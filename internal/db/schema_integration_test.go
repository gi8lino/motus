package db

import (
	"context"
	"io"
	"io/fs"
	"log/slog"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gi8lino/motus/internal/utils"
)

func TestEnsureSchemaPostgres(t *testing.T) {
	databaseURL := os.Getenv("MOTUS_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set MOTUS_TEST_DATABASE_URL to run PostgreSQL migration tests")
	}

	ctx := context.Background()
	admin, err := New(ctx, databaseURL)
	require.NoError(t, err)
	defer admin.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	planned, err := loadMigrations(migrationFiles)
	require.NoError(t, err)

	t.Run("fresh schema", func(t *testing.T) {
		store := newMigrationTestStore(t, ctx, admin, databaseURL, "motus_test_")

		require.NoError(t, store.EnsureSchema(ctx, logger))
		require.NoError(t, store.EnsureSchema(ctx, logger), "migrations must be safe to check repeatedly")
		requireMigrationCount(t, ctx, store, len(planned))

		for _, table := range []string{"users", "workouts", "exercises", "exercise_labels", "user_sessions"} {
			var exists bool
			require.NoError(t, store.pool.QueryRow(ctx, `SELECT to_regclass($1) IS NOT NULL`, table).Scan(&exists))
			require.True(t, exists, table)
		}
	})

	t.Run("existing schema without migration history", func(t *testing.T) {
		store := newMigrationTestStore(t, ctx, admin, databaseURL, "motus_legacy_test_")

		for _, item := range planned {
			sql, err := fs.ReadFile(migrationFiles, "migrations/"+item.entry.Name())
			require.NoError(t, err)
			_, err = store.pool.Exec(ctx, string(sql))
			require.NoErrorf(t, err, "apply legacy schema migration %s", item.entry.Name())
		}

		require.NoError(t, store.EnsureSchema(ctx, logger), "existing schemas must be adopted without duplicate DDL failures")
		requireMigrationCount(t, ctx, store, len(planned))
	})
}

func newMigrationTestStore(t *testing.T, ctx context.Context, admin *Store, databaseURL, prefix string) *Store {
	t.Helper()

	schema := prefix + utils.NewID()
	_, err := admin.pool.Exec(ctx, `CREATE SCHEMA `+schema)
	require.NoError(t, err)
	t.Cleanup(func() { _, _ = admin.pool.Exec(context.Background(), `DROP SCHEMA `+schema+` CASCADE`) })

	parsed, err := url.Parse(databaseURL)
	require.NoError(t, err)
	query := parsed.Query()
	query.Set("search_path", schema)
	parsed.RawQuery = query.Encode()

	store, err := New(ctx, parsed.String())
	require.NoError(t, err)
	t.Cleanup(store.Close)
	return store
}

func requireMigrationCount(t *testing.T, ctx context.Context, store *Store, expected int) {
	t.Helper()

	var applied int
	require.NoError(t, store.pool.QueryRow(ctx, `SELECT count(*) FROM schema_migrations`).Scan(&applied))
	require.Equal(t, expected, applied)
}
