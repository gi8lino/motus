package db

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/gi8lino/motus/internal/utils"
)

// BackfillCoreExercises creates core exercises from existing workout data and links them.
func (s *Store) BackfillCoreExercises(ctx context.Context) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) // nolint:errcheck

	names, err := s.collectDistinctExerciseNames(tx)
	if err != nil {
		return err
	}
	if len(names) == 0 {
		return tx.Commit(ctx)
	}

	existing, err := s.collectExistingExercises(ctx, tx)
	if err != nil {
		return err
	}

	if err := s.insertMissingCoreExercises(ctx, tx, names, existing); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `
		UPDATE workout_subset_exercises wse
		SET exercise_id = e.id
		FROM exercises e
		WHERE wse.exercise_id = ''
		AND wse.name <> ''
		AND LOWER(wse.name) = LOWER(e.name)`); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// collectDistinctExerciseNames returns a list of trimmed, unique exercise names from subsets.
func (s *Store) collectDistinctExerciseNames(tx pgx.Tx) ([]string, error) {
	rows, err := tx.Query(context.Background(), `
		SELECT DISTINCT name FROM workout_subset_exercises WHERE name <> ''`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var unique []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		trimmed := strings.TrimSpace(name)
		if trimmed != "" {
			unique = append(unique, trimmed)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return unique, nil
}

// collectExistingExercises loads existing exercise IDs keyed by normalized name.
func (s *Store) collectExistingExercises(ctx context.Context, tx pgx.Tx) (map[string]string, error) {
	rows, err := tx.Query(ctx, `
		SELECT id, name
		FROM exercises
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	existing := make(map[string]string)
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		existing[utils.NormalizeToken(name)] = id
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return existing, nil
}

// insertMissingCoreExercises inserts any names that are not already present in the catalog.
func (s *Store) insertMissingCoreExercises(ctx context.Context, tx pgx.Tx, names []string, existing map[string]string) error {
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
	return nil
}

// BackfillCoreExercises creates core exercises from existing workout data and links them.

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
		return nil, errors.New("exercise name required")
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
		return nil, errors.New("exercise name required")
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
		UPDATE workout_subset_exercises
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
		UPDATE workout_subset_exercises
		SET exercise_id=$1, name=$2
		WHERE exercise_id=$3
		AND subset_id IN (
			SELECT su.id
			FROM workout_subsets su
			JOIN workout_steps ws ON su.step_id = ws.id
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
		UPDATE workout_subset_exercises
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
