package db

import (
	"context"
	"embed"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
)

const schemaVersionLatest = 5

type schemaMigration struct {
	version    int
	name       string
	statements []string
	file       string
}

//go:embed migrations/*.sql
var migrationFiles embed.FS

var schemaMigrations = []schemaMigration{
	{
		version: 1,
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
            repeat_rest_name TEXT NOT NULL DEFAULT '',
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
			`CREATE TABLE IF NOT EXISTS workout_trainings (
            id TEXT PRIMARY KEY,
            workout_id TEXT NOT NULL REFERENCES workouts(id) ON DELETE CASCADE,
            workout_name TEXT NOT NULL DEFAULT '',
            user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
            started_at TIMESTAMPTZ NOT NULL,
            completed_at TIMESTAMPTZ NOT NULL
        )`,
			`CREATE TABLE IF NOT EXISTS training_steps (
            id TEXT PRIMARY KEY,
            training_id TEXT NOT NULL REFERENCES workout_trainings(id) ON DELETE CASCADE,
            step_order INT NOT NULL,
            step_type TEXT NOT NULL,
            name TEXT NOT NULL,
            estimated_seconds INT NOT NULL,
            elapsed_millis BIGINT NOT NULL DEFAULT 0
        )`,
		},
	},
	{
		version: 2,
		name:    "repeat rest name",
		file:    "migrations/002_repeat_rest_name.sql",
		statements: []string{
			`ALTER TABLE workout_steps
				ADD COLUMN IF NOT EXISTS repeat_rest_name TEXT NOT NULL DEFAULT ''`,
		},
	},
	{
		version: 3,
		name:    "exercise sides",
		file:    "migrations/003_exercise_sides.sql",
		statements: []string{
			`ALTER TABLE workout_subset_exercises ADD COLUMN IF NOT EXISTS side TEXT NOT NULL DEFAULT 'not_applicable'`,
			`ALTER TABLE exercises ADD COLUMN IF NOT EXISTS has_sides BOOLEAN NOT NULL DEFAULT FALSE`,
			`UPDATE workout_subset_exercises SET side = CASE WHEN name ~* '[[:space:]]left$' THEN 'left' WHEN name ~* '[[:space:]]right$' THEN 'right' ELSE side END, name = regexp_replace(name, '[[:space:]]+(left|right)$', '', 'i') WHERE name ~* '[[:space:]](left|right)$'`,
			`DO $$
			DECLARE duplicate RECORD;
			DECLARE target_id TEXT;
			BEGIN
			  FOR duplicate IN SELECT id, regexp_replace(name, '[[:space:]]+(left|right)$', '', 'i') AS base_name FROM exercises WHERE name ~* '[[:space:]](left|right)$' LOOP
			    SELECT id INTO target_id FROM exercises WHERE lower(name) = lower(duplicate.base_name) AND id <> duplicate.id LIMIT 1;
			    IF target_id IS NULL THEN
			      UPDATE exercises SET name = duplicate.base_name, has_sides = TRUE WHERE id = duplicate.id;
			      target_id := duplicate.id;
			    ELSE
			      UPDATE exercises SET has_sides = TRUE WHERE id = target_id;
			      UPDATE workout_subset_exercises SET exercise_id = target_id WHERE exercise_id = duplicate.id;
			      DELETE FROM exercises WHERE id = duplicate.id;
			    END IF;
			    target_id := NULL;
			  END LOOP;
			END $$`,
			`UPDATE workout_subset_exercises wse SET name = e.name FROM exercises e WHERE wse.exercise_id = e.id AND e.has_sides = TRUE`,
		},
	},
	{
		version: 4,
		name:    "local auth sessions",
		file:    "migrations/004_local_auth_sessions.sql",
		statements: []string{
			`CREATE TABLE IF NOT EXISTS user_sessions (
				token TEXT PRIMARY KEY,
				user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
				expires_at TIMESTAMPTZ NOT NULL,
				created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
			)`,
			`CREATE INDEX IF NOT EXISTS user_sessions_user_id_idx ON user_sessions(user_id)`,
			`CREATE INDEX IF NOT EXISTS user_sessions_expires_at_idx ON user_sessions(expires_at)`,
		},
	},
	{
		version: 5,
		name:    "constraints and indexes",
		file:    "migrations/005_constraints_and_indexes.sql",
		statements: []string{
			`ALTER TABLE workout_subset_exercises ADD CONSTRAINT workout_exercise_side_check CHECK (side IN ('left', 'right', 'not_applicable'))`,
			`ALTER TABLE workout_subset_exercises ADD CONSTRAINT workout_exercise_type_check CHECK (exercise_type IN ('rep', 'stopwatch', 'countdown'))`,
			`ALTER TABLE workout_steps ADD CONSTRAINT workout_step_estimate_check CHECK (estimated_seconds >= 0)`,
			`ALTER TABLE workout_subsets ADD CONSTRAINT workout_subset_estimate_check CHECK (estimated_seconds >= 0)`,
			`ALTER TABLE training_steps ADD CONSTRAINT training_step_elapsed_check CHECK (elapsed_millis >= 0)`,
			`CREATE INDEX IF NOT EXISTS workouts_user_id_idx ON workouts(user_id)`,
			`CREATE INDEX IF NOT EXISTS workout_trainings_user_completed_idx ON workout_trainings(user_id, completed_at DESC)`,
			`CREATE INDEX IF NOT EXISTS workout_steps_workout_order_idx ON workout_steps(workout_id, step_order)`,
			`CREATE INDEX IF NOT EXISTS workout_subsets_step_order_idx ON workout_subsets(step_id, subset_order)`,
			`CREATE INDEX IF NOT EXISTS workout_exercises_subset_order_idx ON workout_subset_exercises(subset_id, exercise_order)`,
			`CREATE INDEX IF NOT EXISTS exercises_lower_name_idx ON exercises(lower(name))`,
		},
	},
}

// EnsureSchema applies the baseline schema and any pending migrations.
func (s *Store) EnsureSchema(ctx context.Context, logger *slog.Logger) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) // nolint:errcheck

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
		startVersion := currentVersion
		statements := migration.statements
		if migration.file != "" {
			data, err := migrationFiles.ReadFile(migration.file)
			if err != nil {
				return fmt.Errorf("read migration %d: %w", migration.version, err)
			}
			results, err := tx.Conn().PgConn().Exec(ctx, string(data)).ReadAll()
			if err != nil {
				return fmt.Errorf("execute migration %d: %w", migration.version, err)
			}
			for _, result := range results {
				if result.Err != nil {
					return fmt.Errorf("execute migration %d: %w", migration.version, result.Err)
				}
			}
			statements = nil
		}
		for _, stmt := range statements {
			if _, err := tx.Exec(ctx, stmt); err != nil {
				return err
			}
		}
		if err := writeSchemaVersion(ctx, tx, migration.version); err != nil {
			return err
		}
		logger.Info(
			"db migration applied",
			slog.String("event", "db_migration_applied"),
			slog.Int("from_version", startVersion),
			slog.Int("to_version", migration.version),
			slog.String("name", migration.name),
		)
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
