package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// RecordTraining stores a completed workout training and optional step timings. Duplicate IDs are ignored.
func (s *Store) RecordTraining(ctx context.Context, log TrainingLog, steps []TrainingStepLog) error {
	// Persist the training log and optional step timings in one transaction.
	if log.ID == "" {
		return fmt.Errorf("training id required")
	}
	if log.StartedAt.IsZero() || log.CompletedAt.IsZero() {
		return fmt.Errorf("training timestamps required")
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) // nolint:errcheck
	if _, err := tx.Exec(ctx, `
		INSERT INTO workout_trainings(
			id,
			workout_id,
			workout_name,
			user_id,
			started_at,
			completed_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO NOTHING
	`,
		log.ID, log.WorkoutID, log.WorkoutName, log.UserID, log.StartedAt, log.CompletedAt); err != nil {
		return err
	}
	if len(steps) > 0 {
		batch := &pgx.Batch{}
		// Queue each step timing insert in the batch.
		for _, st := range steps {
			batch.Queue(
				`
					INSERT INTO training_steps(
						id,
						training_id,
						step_order,
						step_type,
						name,
						estimated_seconds,
						elapsed_millis
					)
					VALUES ($1, $2, $3, $4, $5, $6, $7)
					ON CONFLICT (id) DO NOTHING
				`,
				st.ID, log.ID, st.StepOrder, st.Type, st.Name, st.EstimatedSeconds, st.ElapsedMillis,
			)
		}
		if err := tx.SendBatch(ctx, batch).Close(); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// TrainingHistory returns recent trainings for a user.
func (s *Store) TrainingHistory(ctx context.Context, userID string, limit int) ([]TrainingLog, error) {
	// Load recent training logs for a user.
	limit = max(limit, 25)
	rows, err := s.pool.Query(ctx, `
		SELECT ws.id, ws.workout_id, COALESCE(w.name, ''), ws.user_id, ws.started_at, ws.completed_at
		FROM workout_trainings ws
		LEFT JOIN workouts w ON ws.workout_id = w.id
		WHERE ws.user_id=$1
		ORDER BY ws.started_at DESC
		LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var history []TrainingLog
	// Collect each training log row.
	for rows.Next() {
		var entry TrainingLog
		if err := rows.Scan(&entry.ID, &entry.WorkoutID, &entry.WorkoutName, &entry.UserID, &entry.StartedAt, &entry.CompletedAt); err != nil {
			return nil, err
		}
		history = append(history, entry)
	}
	return history, rows.Err()
}

// TrainingStepTimings returns stored step durations for a training.
func (s *Store) TrainingStepTimings(ctx context.Context, trainingID string) ([]TrainingStepLog, error) {
	// Load stored step durations for a training.
	rows, err := s.pool.Query(ctx, `
		SELECT id, training_id, step_order, step_type, name, estimated_seconds, elapsed_millis
		FROM training_steps
		WHERE training_id=$1
		ORDER BY step_order ASC`, trainingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var steps []TrainingStepLog
	// Collect each step timing row.
	for rows.Next() {
		var st TrainingStepLog
		if err := rows.Scan(&st.ID, &st.TrainingID, &st.StepOrder, &st.Type, &st.Name, &st.EstimatedSeconds, &st.ElapsedMillis); err != nil {
			return nil, err
		}
		steps = append(steps, st)
	}
	return steps, rows.Err()
}
