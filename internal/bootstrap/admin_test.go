package bootstrap

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gi8lino/motus/internal/db"
)

type fakeStore struct {
	upsertFn func(context.Context, string, string) (*db.User, bool, error)
}

func (f *fakeStore) UpsertAdminUser(ctx context.Context, email, passwordHash string) (*db.User, bool, error) {
	if f.upsertFn == nil {
		return nil, false, nil
	}
	return f.upsertFn(ctx, email, passwordHash)
}

func TestEnsureAdminUser(t *testing.T) {
	t.Parallel()

	t.Run("No config returns nil", func(t *testing.T) {
		t.Parallel()

		err := EnsureAdminUser(context.Background(), &fakeStore{}, slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)), "", "")
		require.NoError(t, err)
	})

	t.Run("Missing email or password errors", func(t *testing.T) {
		t.Parallel()

		err := EnsureAdminUser(context.Background(), &fakeStore{}, slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)), "admin@example.com", "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "admin email and password")
	})

	t.Run("Invalid email errors", func(t *testing.T) {
		t.Parallel()

		err := EnsureAdminUser(context.Background(), &fakeStore{}, slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)), "not-an-email", "secret")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid admin email")
	})

	t.Run("Upsert error is returned", func(t *testing.T) {
		t.Parallel()

		store := &fakeStore{upsertFn: func(context.Context, string, string) (*db.User, bool, error) {
			return nil, false, errors.New("boom")
		}}
		err := EnsureAdminUser(context.Background(), store, slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)), "admin@example.com", "secret")
		require.Error(t, err)
		assert.EqualError(t, err, "boom")
	})

	t.Run("Logs create and update", func(t *testing.T) {
		t.Parallel()

		var logs bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&logs, nil))

		created := &fakeStore{upsertFn: func(context.Context, string, string) (*db.User, bool, error) {
			return &db.User{ID: "admin"}, true, nil
		}}
		err := EnsureAdminUser(context.Background(), created, logger, "admin@example.com", "secret")
		require.NoError(t, err)
		assert.Contains(t, logs.String(), "created bootstrap admin user")

		logs.Reset()
		updated := &fakeStore{upsertFn: func(context.Context, string, string) (*db.User, bool, error) {
			return &db.User{ID: "admin"}, false, nil
		}}
		err = EnsureAdminUser(context.Background(), updated, logger, "admin@example.com", "secret")
		require.NoError(t, err)
		assert.Contains(t, logs.String(), "updated bootstrap admin user")
	})
}
