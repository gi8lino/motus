package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gi8lino/motus/internal/db"
)

type fakeStore struct {
	getUserFn    func(context.Context, string) (*db.User, error)
	createUserFn func(context.Context, string, string, string) (*db.User, error)
}

func (f *fakeStore) GetUser(ctx context.Context, email string) (*db.User, error) {
	if f.getUserFn == nil {
		return nil, nil
	}
	return f.getUserFn(ctx, email)
}

func (f *fakeStore) CreateUser(ctx context.Context, email, avatarURL, passwordHash string) (*db.User, error) {
	if f.createUserFn == nil {
		return nil, nil
	}
	return f.createUserFn(ctx, email, avatarURL, passwordHash)
}

func TestResolveUserID(t *testing.T) {
	t.Parallel()

	t.Run("Uses auth header when configured", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-User-Email", "User@Example.com")

		id, err := ResolveUserID(req, &fakeStore{}, "X-User-Email", false, "")
		require.NoError(t, err)
		assert.Equal(t, "user@example.com", id)
	})

	t.Run("Uses fallback when no auth header", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		id, err := ResolveUserID(req, &fakeStore{}, "", false, "User@Example.com")
		require.NoError(t, err)
		assert.Equal(t, "user@example.com", id)
	})

	t.Run("Uses local header when fallback missing", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(localAuthHeader, "User@Example.com")
		id, err := ResolveUserID(req, &fakeStore{}, "", false, "")
		require.NoError(t, err)
		assert.Equal(t, "user@example.com", id)
	})

	t.Run("Requires auth header when configured", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		_, err := ResolveUserID(req, &fakeStore{}, "X-User-Email", false, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "auth header")
	})

	t.Run("Auto-create ensures user", func(t *testing.T) {
		t.Parallel()

		created := false
		store := &fakeStore{
			getUserFn: func(context.Context, string) (*db.User, error) {
				return nil, pgx.ErrNoRows
			},
			createUserFn: func(context.Context, string, string, string) (*db.User, error) {
				created = true
				return &db.User{ID: "user@example.com"}, nil
			},
		}
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-User-Email", "user@example.com")

		id, err := ResolveUserID(req, store, "X-User-Email", true, "")
		require.NoError(t, err)
		assert.Equal(t, "user@example.com", id)
		assert.True(t, created, "expected user to be created")
	})

	t.Run("Auto-create ignores race", func(t *testing.T) {
		t.Parallel()

		store := &fakeStore{
			getUserFn: func(context.Context, string) (*db.User, error) {
				return &db.User{ID: "user@example.com"}, nil
			},
			createUserFn: func(context.Context, string, string, string) (*db.User, error) {
				return nil, errors.New("conflict")
			},
		}
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-User-Email", "user@example.com")

		id, err := ResolveUserID(req, store, "X-User-Email", true, "")
		require.NoError(t, err)
		assert.Equal(t, "user@example.com", id)
	})
}
