package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// RecordSession stores a completed workout session and optional step timings. Duplicate IDs are ignored.
func (s *Store) RecordSession(ctx context.Context, log SessionLog, steps []SessionStepLog) error {
	if log.ID == "" {
		return fmt.Errorf("session id required")
	}
	if log.StartedAt.IsZero() || log.CompletedAt.IsZero() {
		return fmt.Errorf("session timestamps required")
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) // nolint:errcheck
	if _, err := tx.Exec(ctx, `
		INSERT INTO workout_sessions(
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
		for _, st := range steps {
			batch.Queue(
				`
					INSERT INTO session_steps(
						id,
						session_id,
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

// SessionHistory returns recent sessions for a user.
func (s *Store) SessionHistory(ctx context.Context, userID string, limit int) ([]SessionLog, error) {
	if limit <= 0 {
		limit = 25
	}
	rows, err := s.pool.Query(ctx, `
		SELECT ws.id, ws.workout_id, COALESCE(w.name, ''), ws.user_id, ws.started_at, ws.completed_at
		FROM workout_sessions ws
		LEFT JOIN workouts w ON ws.workout_id = w.id
		WHERE ws.user_id=$1
		ORDER BY ws.started_at DESC
		LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var history []SessionLog
	for rows.Next() {
		var entry SessionLog
		if err := rows.Scan(&entry.ID, &entry.WorkoutID, &entry.WorkoutName, &entry.UserID, &entry.StartedAt, &entry.CompletedAt); err != nil {
			return nil, err
		}
		history = append(history, entry)
	}
	return history, rows.Err()
}

// SessionStepTimings returns stored step durations for a session.
func (s *Store) SessionStepTimings(ctx context.Context, sessionID string) ([]SessionStepLog, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, session_id, step_order, step_type, name, estimated_seconds, elapsed_millis
		FROM session_steps
		WHERE session_id=$1
		ORDER BY step_order ASC`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var steps []SessionStepLog
	for rows.Next() {
		var st SessionStepLog
		if err := rows.Scan(&st.ID, &st.SessionID, &st.StepOrder, &st.Type, &st.Name, &st.EstimatedSeconds, &st.ElapsedMillis); err != nil {
			return nil, err
		}
		steps = append(steps, st)
	}
	return steps, rows.Err()
}
