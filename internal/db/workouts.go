package db

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/gi8lino/motus/internal/utils"
)

// CreateWorkout inserts a workout and its steps for a user.
func (s *Store) CreateWorkout(ctx context.Context, w *Workout) (*Workout, error) {
	// Persist workout and steps as a new workout.
	return s.insertWorkout(ctx, w, false)
}

// insertWorkout stores a workout and optionally marks it as a template.
func (s *Store) insertWorkout(ctx context.Context, w *Workout, isTemplate bool) (*Workout, error) {
	// Start a transaction so workout and steps are created together.
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) // nolint:errcheck

	w.ID = utils.NewID()
	w.CreatedAt = time.Now().UTC()
	if _, err := tx.Exec(ctx, `
		INSERT INTO workouts(id, user_id, name, is_template, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, w.ID, w.UserID, w.Name, isTemplate, w.CreatedAt); err != nil {
		return nil, err
	}

	// Insert each step and its nested exercises.
	for idx := range w.Steps {
		step := &w.Steps[idx]
		step.ID = utils.NewID()
		step.WorkoutID = w.ID
		step.Order = idx
		step.CreatedAt = time.Now().UTC()
		step.NormalizeRepeatSettings()

		if _, err := tx.Exec(ctx, `
			INSERT INTO workout_steps(
				id,
				workout_id,
				step_order,
				step_type,
				name,
				estimated_seconds,
				sound_key,
				pause_auto_advance,
				repeat_count,
				repeat_rest_seconds,
				repeat_rest_after_last,
				repeat_rest_sound_key,
				repeat_rest_auto_advance,
				repeat_rest_name,
				created_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		`,
			step.ID,
			step.WorkoutID,
			step.Order,
			step.Type,
			step.Name,
			step.EstimatedSeconds,
			step.SoundKey,
			step.PauseOptions.AutoAdvance,
			step.RepeatCount,
			step.RepeatRestSeconds,
			step.RepeatRestAfterLast,
			step.RepeatRestSoundKey,
			step.RepeatRestAutoAdvance,
			step.RepeatRestName,
			step.CreatedAt,
		); err != nil {
			return nil, err
		}

		if err := s.insertStepSubsets(ctx, tx, step.ID, step.Subsets); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	w.IsTemplate = isTemplate
	return w, nil
}

// WorkoutsByUser returns workouts for a user.
func (s *Store) WorkoutsByUser(ctx context.Context, userID string) ([]Workout, error) {
	// Load workouts and their steps for the given user.
	rows, err := s.pool.Query(ctx, `
		SELECT id, user_id, name, is_template, created_at
		FROM workouts
		WHERE user_id=$1 AND is_template=FALSE
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workouts []Workout
	// Collect workouts and hydrate each with steps.
	for rows.Next() {
		var w Workout
		if err := rows.Scan(&w.ID, &w.UserID, &w.Name, &w.IsTemplate, &w.CreatedAt); err != nil {
			return nil, err
		}
		steps, err := s.WorkoutSteps(ctx, w.ID)
		if err != nil {
			return nil, err
		}
		w.Steps = steps
		workouts = append(workouts, w)
	}
	return workouts, rows.Err()
}

// WorkoutSteps fetches all steps for a workout ordered by step order.
func (s *Store) WorkoutSteps(ctx context.Context, workoutID string) ([]WorkoutStep, error) {
	// Load steps and associated exercises for a workout.
	rows, err := s.pool.Query(ctx, `
		SELECT id,
			workout_id,
			step_order,
			step_type,
			name,
			estimated_seconds,
			sound_key,
			pause_auto_advance,
			repeat_count,
			repeat_rest_seconds,
			repeat_rest_after_last,
			repeat_rest_sound_key,
			repeat_rest_auto_advance,
			repeat_rest_name,
			created_at
		FROM workout_steps
		WHERE workout_id=$1
		ORDER BY step_order ASC
	`, workoutID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var steps []WorkoutStep
	// Collect step rows.
	for rows.Next() {
		var st WorkoutStep
		if err := rows.Scan(
			&st.ID,
			&st.WorkoutID,
			&st.Order,
			&st.Type,
			&st.Name,
			&st.EstimatedSeconds,
			&st.SoundKey,
			&st.PauseOptions.AutoAdvance,
			&st.RepeatCount,
			&st.RepeatRestSeconds,
			&st.RepeatRestAfterLast,
			&st.RepeatRestSoundKey,
			&st.RepeatRestAutoAdvance,
			&st.RepeatRestName,
			&st.CreatedAt,
		); err != nil {
			return nil, err
		}
		if st.RepeatCount == 0 {
			st.RepeatCount = 1
		}
		steps = append(steps, st)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(steps) == 0 {
		return steps, nil
	}

	stepIDs := make([]string, 0, len(steps))
	for _, step := range steps {
		stepIDs = append(stepIDs, step.ID)
	}

	type subsetBuilder struct {
		subset    WorkoutSubset
		exercises []SubsetExercise
	}
	stepSubsets := make(map[string][]*subsetBuilder)
	subsetByID := make(map[string]*subsetBuilder)
	subsetRows, err := s.pool.Query(ctx, `
		SELECT id, step_id, subset_order, name, estimated_seconds, sound_key, superset, created_at
		FROM workout_subsets
		WHERE step_id = ANY($1)
		ORDER BY step_id, subset_order ASC
	`, stepIDs)
	if err != nil {
		return nil, err
	}
	defer subsetRows.Close()
	for subsetRows.Next() {
		var b subsetBuilder
		if err := subsetRows.Scan(
			&b.subset.ID,
			&b.subset.StepID,
			&b.subset.Order,
			&b.subset.Name,
			&b.subset.EstimatedSeconds,
			&b.subset.SoundKey,
			&b.subset.Superset,
			&b.subset.CreatedAt,
		); err != nil {
			return nil, err
		}
		stepSubsets[b.subset.StepID] = append(stepSubsets[b.subset.StepID], &b)
		subsetByID[b.subset.ID] = &b
	}
	if err := subsetRows.Err(); err != nil {
		return nil, err
	}

	if len(subsetByID) > 0 {
		subsetIDs := make([]string, 0, len(subsetByID))
		for id := range subsetByID {
			subsetIDs = append(subsetIDs, id)
		}
		exRows, err := s.pool.Query(ctx, `
			SELECT id, subset_id, exercise_order, exercise_id, name, exercise_type, reps, weight, duration, sound_key
			FROM workout_subset_exercises
			WHERE subset_id = ANY($1)
			ORDER BY subset_id, exercise_order
		`, subsetIDs)
		if err != nil {
			return nil, err
		}
		defer exRows.Close()
		for exRows.Next() {
			var ex SubsetExercise
			if err := exRows.Scan(
				&ex.ID,
				&ex.SubsetID,
				&ex.Order,
				&ex.ExerciseID,
				&ex.Name,
				&ex.Type,
				&ex.Reps,
				&ex.Weight,
				&ex.Duration,
				&ex.SoundKey,
			); err != nil {
				return nil, err
			}
			ex.Type = utils.NormalizeExerciseType(ex.Type)
			if builder, ok := subsetByID[ex.SubsetID]; ok {
				builder.exercises = append(builder.exercises, ex)
			}
		}
		if err := exRows.Err(); err != nil {
			return nil, err
		}
	}

	for idx := range steps {
		if builders, ok := stepSubsets[steps[idx].ID]; ok {
			steps[idx].Subsets = make([]WorkoutSubset, len(builders))
			for i, builder := range builders {
				builder.subset.Exercises = builder.exercises
				steps[idx].Subsets[i] = builder.subset
			}
		}
	}

	return steps, nil
}

// WorkoutWithSteps retrieves a workout by id.
func (s *Store) WorkoutWithSteps(ctx context.Context, workoutID string) (*Workout, error) {
	// Fetch the workout row and hydrate its steps.
	row := s.pool.QueryRow(ctx, `
		SELECT id, user_id, name, is_template, created_at
		FROM workouts
		WHERE id=$1
	`, workoutID)
	var w Workout
	if err := row.Scan(&w.ID, &w.UserID, &w.Name, &w.IsTemplate, &w.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("workout not found")
		}
		return nil, err
	}
	steps, err := s.WorkoutSteps(ctx, w.ID)
	if err != nil {
		return nil, err
	}
	w.Steps = steps
	return &w, nil
}

// UpdateWorkout replaces the workout name and steps.
func (s *Store) UpdateWorkout(ctx context.Context, w *Workout) (*Workout, error) {
	// Replace workout name and step definitions in a transaction.
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) // nolint:errcheck

	if _, err := tx.Exec(ctx, `
		UPDATE workouts
		SET name=$1
		WHERE id=$2
	`, w.Name, w.ID); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `
		DELETE FROM workout_steps
		WHERE workout_id=$1
	`, w.ID); err != nil {
		return nil, err
	}

	// Recreate steps after clearing previous definitions.
	for idx := range w.Steps {
		step := &w.Steps[idx]
		step.ID = utils.NewID()
		step.WorkoutID = w.ID
		step.Order = idx
		step.CreatedAt = time.Now().UTC()
		step.NormalizeRepeatSettings()

		if _, err := tx.Exec(ctx, `
			INSERT INTO workout_steps(
				id,
				workout_id,
				step_order,
				step_type,
				name,
				estimated_seconds,
				sound_key,
				pause_auto_advance,
				repeat_count,
				repeat_rest_seconds,
				repeat_rest_after_last,
				repeat_rest_sound_key,
				repeat_rest_auto_advance,
				repeat_rest_name,
				created_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		`,
			step.ID,
			step.WorkoutID,
			step.Order,
			step.Type,
			step.Name,
			step.EstimatedSeconds,
			step.SoundKey,
			step.PauseOptions.AutoAdvance,
			step.RepeatCount,
			step.RepeatRestSeconds,
			step.RepeatRestAfterLast,
			step.RepeatRestSoundKey,
			step.RepeatRestAutoAdvance,
			step.RepeatRestName,
			step.CreatedAt,
		); err != nil {
			return nil, err
		}
		if err := s.insertStepSubsets(ctx, tx, step.ID, step.Subsets); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return w, nil
}

// cloneSteps copies workout steps for template/workout reuse.
func cloneSteps(src []WorkoutStep) []WorkoutStep {
	result := make([]WorkoutStep, len(src))
	for i := range src {
		result[i] = src[i]
		result[i].ID = ""
		result[i].Subsets = make([]WorkoutSubset, len(src[i].Subsets))
		for j := range src[i].Subsets {
			result[i].Subsets[j] = src[i].Subsets[j]
			result[i].Subsets[j].ID = ""
			result[i].Subsets[j].StepID = ""
			result[i].Subsets[j].Exercises = make([]SubsetExercise, len(src[i].Subsets[j].Exercises))
			for k := range src[i].Subsets[j].Exercises {
				result[i].Subsets[j].Exercises[k] = src[i].Subsets[j].Exercises[k]
				result[i].Subsets[j].Exercises[k].ID = ""
				result[i].Subsets[j].Exercises[k].SubsetID = ""
			}
		}
	}
	return result
}

// DeleteWorkout removes a workout and cascades its steps.
func (s *Store) DeleteWorkout(ctx context.Context, workoutID string) error {
	// Delete the workout and rely on cascading deletes for related rows.
	tag, err := s.pool.Exec(ctx, `
		DELETE FROM workouts
		WHERE id=$1
	`, workoutID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// insertStepSubsets saves subset rows for a workout step.
func (s *Store) insertStepSubsets(ctx context.Context, tx pgx.Tx, stepID string, subsets []WorkoutSubset) error {
	for idx := range subsets {
		sub := subsets[idx]
		name := strings.TrimSpace(sub.Name)
		sub.ID = utils.NewID()
		sub.StepID = stepID
		sub.Order = idx
		sub.Name = name
		sub.EstimatedSeconds = max(sub.EstimatedSeconds, 0)
		sub.CreatedAt = time.Now().UTC()
		if _, err := tx.Exec(ctx, `
			INSERT INTO workout_subsets(
				id,
				step_id,
				subset_order,
				name,
				estimated_seconds,
				sound_key,
				superset,
				created_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`,
			sub.ID,
			sub.StepID,
			sub.Order,
			sub.Name,
			sub.EstimatedSeconds,
			sub.SoundKey,
			sub.Superset,
			sub.CreatedAt,
		); err != nil {
			return err
		}
		if err := s.insertSubsetExercises(ctx, tx, sub.ID, sub.Exercises); err != nil {
			return err
		}
	}
	return nil
}

// insertSubsetExercises saves the exercise rows for a workout subset.
func (s *Store) insertSubsetExercises(ctx context.Context, tx pgx.Tx, subsetID string, exercises []SubsetExercise) error {
	for idx := range exercises {
		ex := exercises[idx]
		name := strings.TrimSpace(ex.Name)
		if isEmptySubsetExercise(ex) && name == "" {
			continue
		}
		ex.ID = utils.NewID()
		ex.SubsetID = subsetID
		ex.Order = idx
		token := utils.DefaultIfZero(utils.NormalizeToken(ex.Type), utils.ExerciseTypeRep)
		exType := utils.NormalizeExerciseType(token)

		if _, err := tx.Exec(ctx, `
			INSERT INTO workout_subset_exercises(
				id,
				subset_id,
				exercise_order,
				exercise_id,
				name,
				exercise_type,
				reps,
				weight,
				duration,
				sound_key
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`,
			ex.ID,
			ex.SubsetID,
			ex.Order,
			strings.TrimSpace(ex.ExerciseID),
			name,
			exType,
			strings.TrimSpace(ex.Reps),
			strings.TrimSpace(ex.Weight),
			strings.TrimSpace(ex.Duration),
			strings.TrimSpace(ex.SoundKey),
		); err != nil {
			return err
		}
	}
	return nil
}

// isEmptySubsetExercise returns true when an exercise row has no meaningful content.
func isEmptySubsetExercise(ex SubsetExercise) bool {
	return strings.TrimSpace(ex.Name) == "" &&
		strings.TrimSpace(ex.Reps) == "" &&
		strings.TrimSpace(ex.Weight) == "" &&
		strings.TrimSpace(ex.Duration) == ""
}
