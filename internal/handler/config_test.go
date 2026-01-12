package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gi8lino/motus/internal/db"
	"github.com/gi8lino/motus/internal/service/users"
)

func TestConfig(t *testing.T) {
	t.Run("Returns config payload", func(t *testing.T) {
		api := &API{
			AuthHeader:        "X-User",
			AllowRegistration: true,
			Version:           "v1",
			Commit:            "c1",
		}

		h := api.Config()
		req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)

		var payload configResponse
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&payload))
		assert.True(t, payload.AuthHeaderEnabled)
		assert.True(t, payload.AllowRegistration)
		assert.Equal(t, "v1", payload.Version)
		assert.Equal(t, "c1", payload.Commit)
	})
}

func TestCurrentUser(t *testing.T) {
	t.Run("Returns current user", func(t *testing.T) {
		store := &fakeUserStore{
			getUserFn: func(_ context.Context, id string) (*db.User, error) {
				return &db.User{ID: id}, nil
			},
		}

		api := &API{Users: users.New(store, "", false)}
		h := api.CurrentUser()
		req := httptest.NewRequest(http.MethodGet, "/api/users/current", nil)
		req.Header.Set("X-User-ID", "user@example.com")
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)

		var payload db.User
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&payload))
		assert.Equal(t, "user@example.com", payload.ID)
	})

	t.Run("Returns not found", func(t *testing.T) {
		store := &fakeUserStore{
			getUserFn: func(context.Context, string) (*db.User, error) {
				return nil, nil
			},
		}

		api := &API{Users: users.New(store, "", false)}
		h := api.CurrentUser()
		req := httptest.NewRequest(http.MethodGet, "/api/users/current", nil)
		req.Header.Set("X-User-ID", "user@example.com")
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		require.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("Returns bad request when missing user", func(t *testing.T) {
		api := &API{Users: users.New(&fakeUserStore{}, "", false)}
		h := api.CurrentUser()
		req := httptest.NewRequest(http.MethodGet, "/api/users/current", nil)
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		require.Equal(t, http.StatusBadRequest, rec.Code)
	})
}
