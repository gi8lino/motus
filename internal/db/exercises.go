package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gi8lino/motus/internal/utils"
)

// BackfillCoreExercises creates core exercises from existing workout data and links them.
func (s *Store) BackfillCoreExercises(ctx context.Context) error {
	// Build core exercises from distinct names in existing workouts.
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) // nolint:errcheck

	nameRows, err := tx.Query(ctx, `
		SELECT DISTINCT name FROM workout_step_exercises WHERE name <> ''`)
	if err != nil {
		return err
	}
	defer nameRows.Close()

	var names []string
	// Gather distinct exercise names from existing workouts.
	for nameRows.Next() {
		var name string
		if err := nameRows.Scan(&name); err != nil {
			return err
		}
		trimmed := strings.TrimSpace(name)
		if trimmed != "" {
			names = append(names, trimmed)
		}
	}
	if err := nameRows.Err(); err != nil {
		return err
	}
	if len(names) == 0 {
		return tx.Commit(ctx)
	}

	existingRows, err := tx.Query(ctx, `
		SELECT id, name
		FROM exercises
	`)
	if err != nil {
		return err
	}
	defer existingRows.Close()

	existing := make(map[string]string)
	// Load existing exercise names to avoid duplicates.
	for existingRows.Next() {
		var id, name string
		if err := existingRows.Scan(&id, &name); err != nil {
			return err
		}
		existing[strings.ToLower(strings.TrimSpace(name))] = id
	}
	if err := existingRows.Err(); err != nil {
		return err
	}

	// Insert any missing core exercise rows.
	for _, name := range names {
		key := strings.ToLower(name)
		if key == "" || existing[key] != "" {
			continue
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO exercises(id, name, owner_user_id, is_core, created_at)
			VALUES ($1, $2, '', TRUE, $3)
		`,
			utils.NewID(), name, time.Now().UTC()); err != nil {
			return err
		}
		existing[key] = "inserted"
	}

	if _, err := tx.Exec(ctx, `
		UPDATE workout_step_exercises wse
		SET exercise_id = e.id
		FROM exercises e
		WHERE wse.exercise_id = ''
		AND wse.name <> ''
		AND LOWER(wse.name) = LOWER(e.name)`); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// ListExercises returns core exercises plus user-owned exercises.
func (s *Store) ListExercises(ctx context.Context, userID string) ([]Exercise, error) {
	// Return core exercises plus user-owned entries.
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, owner_user_id, (is_core OR owner_user_id IS NULL OR owner_user_id = '') AS is_core, created_at
		FROM exercises
		WHERE is_core = TRUE OR owner_user_id = $1 OR owner_user_id IS NULL OR owner_user_id = ''
		ORDER BY is_core DESC, name ASC`, strings.TrimSpace(userID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exercises []Exercise
	// Collect each exercise row for the response.
	for rows.Next() {
		var ex Exercise
		var ownerID *string
		if err := rows.Scan(&ex.ID, &ex.Name, &ownerID, &ex.IsCore, &ex.CreatedAt); err != nil {
			return nil, err
		}
		if ownerID != nil {
			ex.OwnerUserID = *ownerID
		}
		exercises = append(exercises, ex)
	}
	return exercises, rows.Err()
}

// GetExercise fetches a single exercise by id.
func (s *Store) GetExercise(ctx context.Context, id string) (*Exercise, error) {
	// Fetch a single exercise row by id.
	row := s.pool.QueryRow(ctx, `
		SELECT id, name, owner_user_id, (is_core OR owner_user_id IS NULL OR owner_user_id = '') AS is_core, created_at
		FROM exercises
		WHERE id=$1`, strings.TrimSpace(id))
	var ex Exercise
	var ownerID *string
	if err := row.Scan(&ex.ID, &ex.Name, &ownerID, &ex.IsCore, &ex.CreatedAt); err != nil {
		return nil, err
	}
	if ownerID != nil {
		ex.OwnerUserID = *ownerID
	}
	return &ex, nil
}

// CreateExercise inserts a new exercise entry.
func (s *Store) CreateExercise(ctx context.Context, name, ownerUserID string, isCore bool) (*Exercise, error) {
	// Insert a new exercise row.
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return nil, fmt.Errorf("exercise name required")
	}
	ex := &Exercise{
		ID:          utils.NewID(),
		Name:        trimmed,
		OwnerUserID: strings.TrimSpace(ownerUserID),
		IsCore:      isCore,
		CreatedAt:   time.Now().UTC(),
	}
	if ex.IsCore {
		ex.OwnerUserID = ""
	}
	if _, err := s.pool.Exec(ctx, `
		INSERT INTO exercises(id, name, owner_user_id, is_core, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`,
		ex.ID, ex.Name, ex.OwnerUserID, ex.IsCore, ex.CreatedAt); err != nil {
		return nil, err
	}
	return ex, nil
}

// RenameExercise updates the catalog name and linked workout exercise names.
func (s *Store) RenameExercise(ctx context.Context, id, name string) (*Exercise, error) {
	// Update exercise name and propagate to workout references.
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return nil, fmt.Errorf("exercise name required")
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) // nolint:errcheck
	if _, err := tx.Exec(ctx, `
		UPDATE exercises
		SET name=$1
		WHERE id=$2
	`, trimmed, strings.TrimSpace(id)); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE workout_step_exercises
		SET name=$1
		WHERE exercise_id=$2
	`, trimmed, strings.TrimSpace(id)); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.GetExercise(ctx, id)
}

// ReplaceExerciseForUser swaps a user workout's exercise references to a new exercise id.
func (s *Store) ReplaceExerciseForUser(ctx context.Context, userID, fromID, toID, toName string) error {
	// Swap exercise references for a user's workouts.
	_, err := s.pool.Exec(ctx, `
		UPDATE workout_step_exercises
		SET exercise_id=$1, name=$2
		WHERE exercise_id=$3
		AND step_id IN (
			SELECT ws.id
			FROM workout_steps ws
			JOIN workouts w ON ws.workout_id = w.id
			WHERE w.user_id=$4
		)`, strings.TrimSpace(toID), strings.TrimSpace(toName), strings.TrimSpace(fromID), strings.TrimSpace(userID))
	return err
}

// DeleteExercise removes an exercise and clears linked workout rows.
func (s *Store) DeleteExercise(ctx context.Context, id string) error {
	// Delete exercise and clear references from workout steps.
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) // nolint:errcheck
	if _, err := tx.Exec(ctx, `
		UPDATE workout_step_exercises
		SET exercise_id=''
		WHERE exercise_id=$1
	`, strings.TrimSpace(id)); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		DELETE FROM exercises
		WHERE id=$1
	`, strings.TrimSpace(id)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
