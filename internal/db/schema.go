package db

import (
	"context"

	"github.com/jackc/pgx/v5"
)

const schemaVersionLatest = 1

type schemaMigration struct {
	version    int
	name       string
	statements []string
}

var schemaMigrations = []schemaMigration{
	{
		version: schemaVersionLatest,
		name:    "baseline",
		statements: []string{
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
            pause_auto_advance BOOLEAN NOT NULL DEFAULT FALSE,
            repeat_count INT NOT NULL DEFAULT 1,
            repeat_rest_seconds INT NOT NULL DEFAULT 0,
            repeat_rest_after_last BOOLEAN NOT NULL DEFAULT FALSE,
            repeat_rest_sound_key TEXT NOT NULL DEFAULT '',
            repeat_rest_auto_advance BOOLEAN NOT NULL DEFAULT FALSE,
            created_at TIMESTAMPTZ NOT NULL
        )`,
			`CREATE TABLE IF NOT EXISTS workout_subsets (
            id TEXT PRIMARY KEY,
            step_id TEXT NOT NULL REFERENCES workout_steps(id) ON DELETE CASCADE,
            subset_order INT NOT NULL,
            name TEXT NOT NULL,
            estimated_seconds INT NOT NULL,
            sound_key TEXT NOT NULL DEFAULT '',
            superset BOOLEAN NOT NULL DEFAULT FALSE,
            created_at TIMESTAMPTZ NOT NULL
        )`,

			`CREATE TABLE IF NOT EXISTS workout_subset_exercises (
            id TEXT PRIMARY KEY,
            subset_id TEXT NOT NULL REFERENCES workout_subsets(id) ON DELETE CASCADE,
            exercise_order INT NOT NULL,
            exercise_id TEXT NOT NULL DEFAULT '',
            name TEXT NOT NULL,
            exercise_type TEXT NOT NULL DEFAULT 'rep',
            reps TEXT NOT NULL DEFAULT '',
            weight TEXT NOT NULL DEFAULT '',
            duration TEXT NOT NULL DEFAULT '',
            sound_key TEXT NOT NULL DEFAULT ''
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
		},
	},
}

// EnsureSchema applies the baseline schema and any pending migrations.
func (s *Store) EnsureSchema(ctx context.Context) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) // nolint: errcheck

	if err := ensureSchemaVersionTable(ctx, tx); err != nil {
		return err
	}
	currentVersion, err := readSchemaVersion(ctx, tx)
	if err != nil {
		return err
	}

	for _, migration := range schemaMigrations {
		if migration.version <= currentVersion {
			// Skip migrations that have already been applied.
			continue
		}
		for _, stmt := range migration.statements {
			if _, err := tx.Exec(ctx, stmt); err != nil {
				return err
			}
		}
		if err := writeSchemaVersion(ctx, tx, migration.version); err != nil {
			return err
		}
		currentVersion = migration.version
	}

	return tx.Commit(ctx)
}

// ensureSchemaVersionTable creates the schema version tracker if missing.
func ensureSchemaVersionTable(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_version (
        id INT PRIMARY KEY,
        version INT NOT NULL,
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    )`)
	return err
}

// readSchemaVersion returns the current schema version (or 0 when missing).
func readSchemaVersion(ctx context.Context, tx pgx.Tx) (int, error) {
	var version int
	err := tx.QueryRow(ctx, `SELECT version FROM schema_version WHERE id = 1`).Scan(&version)
	if err == pgx.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return version, nil
}

// writeSchemaVersion persists the latest schema version.
func writeSchemaVersion(ctx context.Context, tx pgx.Tx, version int) error {
	_, err := tx.Exec(ctx, `INSERT INTO schema_version (id, version, updated_at)
        VALUES (1, $1, NOW())
        ON CONFLICT (id) DO UPDATE SET version = EXCLUDED.version, updated_at = NOW()`, version)
	return err
}
