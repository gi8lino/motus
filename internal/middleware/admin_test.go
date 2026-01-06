package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gi8lino/motus/internal/db"
)

type stubAdminStore struct {
	users map[string]*db.User
	err   error
}

// GetUser returns a user or error for testing RequireAdmin.
func (s stubAdminStore) GetUser(_ context.Context, id string) (*db.User, error) {
	if s.err != nil {
		return nil, s.err
	}
	user, ok := s.users[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	return user, nil
}

// TestRequireAdmin covers admin guard behavior for missing, invalid, and valid users.
func TestRequireAdmin(t *testing.T) {
	t.Parallel()

	t.Run("missing header", func(t *testing.T) {
		t.Parallel()

		handler := RequireAdmin(stubAdminStore{}, "")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Equal(t, "forbidden", rec.Body.String())
	})

	t.Run("non-admin user", func(t *testing.T) {
		t.Parallel()

		store := stubAdminStore{
			users: map[string]*db.User{
				"user@example.com": {ID: "user@example.com", IsAdmin: false},
			},
		}
		handler := RequireAdmin(store, "")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-User-ID", "user@example.com")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Equal(t, "forbidden", rec.Body.String())
	})

	t.Run("admin user", func(t *testing.T) {
		t.Parallel()

		store := stubAdminStore{
			users: map[string]*db.User{
				"admin@example.com": {ID: "admin@example.com", IsAdmin: true},
			},
		}
		called := false
		handler := RequireAdmin(store, "")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-User-ID", "admin@example.com")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		require.True(t, called)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "ok", rec.Body.String())
	})

	t.Run("custom auth header", func(t *testing.T) {
		t.Parallel()

		store := stubAdminStore{
			users: map[string]*db.User{
				"admin@example.com": {ID: "admin@example.com", IsAdmin: true},
			},
		}
		handler := RequireAdmin(store, "X-User-Email")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-User-Email", "admin@example.com")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
}
