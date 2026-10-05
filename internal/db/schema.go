package db

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"slices"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// migration is one validated, versioned SQL migration.
type migration struct {
	version int
	entry   fs.DirEntry
}

// EnsureSchema applies each embedded migration exactly once and records its version.
func (s *Store) EnsureSchema(ctx context.Context, logger *slog.Logger) error {
	migrations, err := loadMigrations(migrationFiles)
	if err != nil {
		return err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := initializeMigrations(ctx, tx); err != nil {
		return err
	}
	for _, item := range migrations {
		applied, err := migrationApplied(ctx, tx, item.version)
		if err != nil {
			return err
		}
		if applied {
			continue
		}
		if err := applyMigration(ctx, tx, item); err != nil {
			return err
		}
		logger.Info("applied database migration", "event", "database_migration_applied", "version", item.version, "migration", item.entry.Name())
	}
	return tx.Commit(ctx)
}

func initializeMigrations(ctx context.Context, tx pgx.Tx) error {
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(734627198236)`); err != nil {
		return err
	}
	_, err := tx.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations(version integer PRIMARY KEY, applied_at timestamptz NOT NULL DEFAULT now())`)
	return err
}

func migrationApplied(ctx context.Context, tx pgx.Tx, version int) (bool, error) {
	var applied bool
	err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version=$1)`, version).Scan(&applied)
	return applied, err
}

func applyMigration(ctx context.Context, tx pgx.Tx, item migration) error {
	sql, err := fs.ReadFile(migrationFiles, "migrations/"+item.entry.Name())
	if err != nil {
		return fmt.Errorf("migration %d: read: %w", item.version, err)
	}
	if _, err := tx.Exec(ctx, string(sql)); err != nil {
		return fmt.Errorf("migration %d: %w", item.version, err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations(version) VALUES($1)`, item.version); err != nil {
		return fmt.Errorf("migration %d: record: %w", item.version, err)
	}
	return nil
}

func loadMigrations(source fs.FS) ([]migration, error) {
	entries, err := fs.ReadDir(source, "migrations")
	if err != nil {
		return nil, err
	}
	migrations := make([]migration, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		prefix, _, found := strings.Cut(entry.Name(), "_")
		if !found {
			return nil, fmt.Errorf("invalid migration %q", entry.Name())
		}
		version, err := strconv.Atoi(prefix)
		if err != nil || version < 1 {
			return nil, fmt.Errorf("invalid migration %q", entry.Name())
		}
		migrations = append(migrations, migration{version: version, entry: entry})
	}
	slices.SortFunc(migrations, func(left, right migration) int { return left.version - right.version })
	for index := 1; index < len(migrations); index++ {
		if migrations[index-1].version == migrations[index].version {
			return nil, fmt.Errorf("duplicate migration version %d", migrations[index].version)
		}
	}
	return migrations, nil
}
