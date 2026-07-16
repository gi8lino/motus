package db

import (
	"context"
	"embed"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
)

const schemaVersionLatest = 6

type schemaMigration struct {
	version int
	name    string
	file    string
}

//go:embed migrations/*.sql
var migrationFiles embed.FS

var schemaMigrations = []schemaMigration{
	{
		version: 1,
		name:    "baseline",
		file:    "migrations/001_baseline.sql",
	},
	{
		version: 2,
		name:    "repeat rest name",
		file:    "migrations/002_repeat_rest_name.sql",
	},
	{
		version: 3,
		name:    "exercise sides",
		file:    "migrations/003_exercise_sides.sql",
	},
	{
		version: 4,
		name:    "local auth sessions",
		file:    "migrations/004_local_auth_sessions.sql",
	},
	{
		version: 5,
		name:    "constraints and indexes",
		file:    "migrations/005_constraints_and_indexes.sql",
	},
	{
		version: 6,
		name:    "exercise labels",
		file:    "migrations/006_exercise_labels.sql",
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
