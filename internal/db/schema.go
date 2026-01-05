package db

import "context"

// EnsureSchema creates required tables.
func (s *Store) EnsureSchema(ctx context.Context) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS users (
            id TEXT PRIMARY KEY,
            name TEXT NOT NULL,
            is_admin BOOLEAN NOT NULL DEFAULT FALSE,
            avatar_url TEXT NOT NULL DEFAULT '',
            password_hash TEXT NOT NULL DEFAULT '',
            created_at TIMESTAMPTZ NOT NULL
        )`,
		`CREATE TABLE IF NOT EXISTS workouts (
            id TEXT PRIMARY KEY,
            user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
            name TEXT NOT NULL,
            is_template BOOLEAN NOT NULL DEFAULT FALSE,
            created_at TIMESTAMPTZ NOT NULL
        )`,
		`CREATE TABLE IF NOT EXISTS workout_steps (
            id TEXT PRIMARY KEY,
            workout_id TEXT NOT NULL REFERENCES workouts(id) ON DELETE CASCADE,
            step_order INT NOT NULL,
            step_type TEXT NOT NULL,
            name TEXT NOT NULL,
            estimated_seconds INT NOT NULL,
            sound_key TEXT NOT NULL DEFAULT '',
            exercise TEXT NOT NULL DEFAULT '',
            amount TEXT NOT NULL DEFAULT '',
            weight TEXT NOT NULL DEFAULT '',
            created_at TIMESTAMPTZ NOT NULL
        )`,
		`CREATE TABLE IF NOT EXISTS workout_step_exercises (
            id TEXT PRIMARY KEY,
            step_id TEXT NOT NULL REFERENCES workout_steps(id) ON DELETE CASCADE,
            exercise_order INT NOT NULL,
            name TEXT NOT NULL,
            amount TEXT NOT NULL DEFAULT '',
            weight TEXT NOT NULL DEFAULT ''
        )`,
		`CREATE TABLE IF NOT EXISTS exercises (
            id TEXT PRIMARY KEY,
            name TEXT NOT NULL UNIQUE,
            owner_user_id TEXT,
            is_core BOOLEAN NOT NULL DEFAULT FALSE,
            created_at TIMESTAMPTZ NOT NULL
        )`,
		`CREATE TABLE IF NOT EXISTS workout_sessions (
	            id TEXT PRIMARY KEY,
	            workout_id TEXT NOT NULL REFERENCES workouts(id) ON DELETE CASCADE,
	            workout_name TEXT NOT NULL DEFAULT '',
	            user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	            started_at TIMESTAMPTZ NOT NULL,
	            completed_at TIMESTAMPTZ NOT NULL
	        )`,
		`CREATE TABLE IF NOT EXISTS session_steps (
	            id TEXT PRIMARY KEY,
	            session_id TEXT NOT NULL REFERENCES workout_sessions(id) ON DELETE CASCADE,
	            step_order INT NOT NULL,
	            step_type TEXT NOT NULL,
	            name TEXT NOT NULL,
	            estimated_seconds INT NOT NULL,
	            elapsed_millis BIGINT NOT NULL DEFAULT 0
	        )`,
		`ALTER TABLE workout_sessions ADD COLUMN IF NOT EXISTS workout_name TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS is_admin BOOLEAN NOT NULL DEFAULT FALSE`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE workouts ADD COLUMN IF NOT EXISTS is_template BOOLEAN NOT NULL DEFAULT FALSE`,
		`ALTER TABLE workout_steps ADD COLUMN IF NOT EXISTS sound_key TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE workout_steps ADD COLUMN IF NOT EXISTS exercise TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE workout_steps ADD COLUMN IF NOT EXISTS amount TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE workout_steps ADD COLUMN IF NOT EXISTS weight TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE workout_step_exercises ADD COLUMN IF NOT EXISTS amount TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE workout_step_exercises ADD COLUMN IF NOT EXISTS weight TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE workout_step_exercises ADD COLUMN IF NOT EXISTS exercise_id TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE exercises ADD COLUMN IF NOT EXISTS owner_user_id TEXT`,
		`ALTER TABLE exercises ADD COLUMN IF NOT EXISTS is_core BOOLEAN NOT NULL DEFAULT FALSE`,
		`ALTER TABLE exercises DROP COLUMN IF EXISTS is_admin`,
	}
	for _, stmt := range stmts {
		if _, err := s.pool.Exec(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}
