package db

import (
	"context"
	"io"
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
	schema := "motus_test_" + utils.NewID()
	_, err = admin.pool.Exec(ctx, `CREATE SCHEMA `+schema)
	require.NoError(t, err)
	t.Cleanup(func() { _, _ = admin.pool.Exec(context.Background(), `DROP SCHEMA `+schema+` CASCADE`) })

	parsed, err := url.Parse(databaseURL)
	require.NoError(t, err)
	query := parsed.Query()
	query.Set("search_path", schema)
	parsed.RawQuery = query.Encode()
	store, err := New(ctx, parsed.String())
	require.NoError(t, err)
	defer store.Close()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	require.NoError(t, store.EnsureSchema(ctx, logger))
	require.NoError(t, store.EnsureSchema(ctx, logger), "migrations must be safe to check repeatedly")
	var version int
	require.NoError(t, store.pool.QueryRow(ctx, `SELECT version FROM schema_version WHERE id=1`).Scan(&version))
	require.Equal(t, schemaVersionLatest, version)
	for _, table := range []string{"users", "workouts", "exercises", "user_sessions"} {
		var exists bool
		require.NoError(t, store.pool.QueryRow(ctx, `SELECT to_regclass($1) IS NOT NULL`, table).Scan(&exists))
		require.True(t, exists, table)
	}
}
