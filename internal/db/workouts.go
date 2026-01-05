package db

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/gi8lino/motus/internal/utils"
)

// CreateWorkout inserts a workout and its steps for a user.
func (s *Store) CreateWorkout(ctx context.Context, w *Workout) (*Workout, error) {
	return s.insertWorkout(ctx, w, false)
}

// insertWorkout stores a workout and optionally marks it as a template.
func (s *Store) insertWorkout(ctx context.Context, w *Workout, isTemplate bool) (*Workout, error) {
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

	for idx := range w.Steps {
		step := &w.Steps[idx]
		step.ID = utils.NewID()
		step.WorkoutID = w.ID
		step.Order = idx
		step.CreatedAt = time.Now().UTC()
		if _, err := tx.Exec(ctx, `
			INSERT INTO workout_steps(
				id,
				workout_id,
				step_order,
				step_type,
				name,
				estimated_seconds,
				sound_key,
				exercise,
				amount,
				weight,
				created_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`,
			step.ID, step.WorkoutID, step.Order, step.Type, step.Name, step.EstimatedSeconds, step.SoundKey, step.Exercise, step.Amount, step.Weight, step.CreatedAt); err != nil {
			return nil, err
		}
		if err := s.insertStepExercises(ctx, tx, step.ID, step.Exercises); err != nil {
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
	rows, err := s.pool.Query(ctx, `
		SELECT id, workout_id, step_order, step_type, name, estimated_seconds, sound_key, exercise, amount, weight, created_at
		FROM workout_steps
		WHERE workout_id=$1
		ORDER BY step_order ASC
	`, workoutID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var steps []WorkoutStep
	index := make(map[string]int)
	for rows.Next() {
		var st WorkoutStep
		if err := rows.Scan(&st.ID, &st.WorkoutID, &st.Order, &st.Type, &st.Name, &st.EstimatedSeconds, &st.SoundKey, &st.Exercise, &st.Amount, &st.Weight, &st.CreatedAt); err != nil {
			return nil, err
		}
		steps = append(steps, st)
		index[st.ID] = len(steps) - 1
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(steps) == 0 {
		return steps, nil
	}

	exRows, err := s.pool.Query(ctx, `
		SELECT e.id, e.step_id, e.exercise_order, e.exercise_id, e.name, e.amount, e.weight
		FROM workout_step_exercises e
		JOIN workout_steps s ON e.step_id = s.id
		WHERE s.workout_id=$1
		ORDER BY s.step_order, e.exercise_order
	`, workoutID)
	if err != nil {
		return nil, err
	}
	defer exRows.Close()
	for exRows.Next() {
		var ex StepExercise
		if err := exRows.Scan(&ex.ID, &ex.StepID, &ex.Order, &ex.ExerciseID, &ex.Name, &ex.Amount, &ex.Weight); err != nil {
			return nil, err
		}
		if idx, ok := index[ex.StepID]; ok {
			steps[idx].Exercises = append(steps[idx].Exercises, ex)
		}
	}
	if err := exRows.Err(); err != nil {
		return nil, err
	}
	for i := range steps {
		if len(steps[i].Exercises) > 0 && steps[i].Type != "pause" {
			steps[i].Exercise = steps[i].Exercises[0].Name
			steps[i].Amount = steps[i].Exercises[0].Amount
			steps[i].Weight = steps[i].Exercises[0].Weight
		}
		if strings.EqualFold(steps[i].Weight, "__auto__") {
			steps[i].PauseOptions = PauseOptions{AutoAdvance: true}
		}
	}
	return steps, nil
}

// WorkoutWithSteps retrieves a workout by id.
func (s *Store) WorkoutWithSteps(ctx context.Context, workoutID string) (*Workout, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, user_id, name, is_template, created_at
		FROM workouts
		WHERE id=$1
	`, workoutID)
	var w Workout
	if err := row.Scan(&w.ID, &w.UserID, &w.Name, &w.IsTemplate, &w.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("workout not found")
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
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		UPDATE workouts
		SET name=$1
		WHERE id=$2
	`, w.Name, w.ID); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `
		DELETE FROM workout_step_exercises
		WHERE step_id IN (SELECT id FROM workout_steps WHERE workout_id=$1)
	`, w.ID); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `
		DELETE FROM workout_steps
		WHERE workout_id=$1
	`, w.ID); err != nil {
		return nil, err
	}
	for idx := range w.Steps {
		step := &w.Steps[idx]
		step.ID = utils.NewID()
		step.WorkoutID = w.ID
		step.Order = idx
		step.CreatedAt = time.Now().UTC()
		if _, err := tx.Exec(ctx, `
			INSERT INTO workout_steps(
				id,
				workout_id,
				step_order,
				step_type,
				name,
				estimated_seconds,
				sound_key,
				exercise,
				amount,
				weight,
				created_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`,
			step.ID, step.WorkoutID, step.Order, step.Type, step.Name, step.EstimatedSeconds, step.SoundKey, step.Exercise, step.Amount, step.Weight, step.CreatedAt); err != nil {
			return nil, err
		}
		if err := s.insertStepExercises(ctx, tx, step.ID, step.Exercises); err != nil {
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
		result[i].Exercises = make([]StepExercise, len(src[i].Exercises))
		for j := range src[i].Exercises {
			result[i].Exercises[j] = src[i].Exercises[j]
			result[i].Exercises[j].ID = ""
			result[i].Exercises[j].StepID = ""
		}
	}
	return result
}

// insertStepExercises saves the exercise rows for a workout step.
func (s *Store) insertStepExercises(ctx context.Context, tx pgx.Tx, stepID string, exercises []StepExercise) error {
	for idx := range exercises {
		ex := exercises[idx]
		name := strings.TrimSpace(ex.Name)
		if name == "" && strings.TrimSpace(ex.Amount) == "" && strings.TrimSpace(ex.Weight) == "" {
			continue
		}
		ex.ID = utils.NewID()
		ex.StepID = stepID
		ex.Order = idx
		if _, err := tx.Exec(ctx, `
			INSERT INTO workout_step_exercises(
				id,
				step_id,
				exercise_order,
				exercise_id,
				name,
				amount,
				weight
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`,
			ex.ID, ex.StepID, ex.Order, strings.TrimSpace(ex.ExerciseID), name, strings.TrimSpace(ex.Amount), strings.TrimSpace(ex.Weight)); err != nil {
			return err
		}
	}
	return nil
}

// DeleteWorkout removes a workout and cascades its steps.
func (s *Store) DeleteWorkout(ctx context.Context, workoutID string) error {
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
