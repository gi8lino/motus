package endpoint

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gi8lino/motus/internal/db"
	httpauth "github.com/gi8lino/motus/internal/http/auth"
)

func TestConfig(t *testing.T) {
	t.Parallel()
	t.Run("Returns config payload", func(t *testing.T) {
		t.Parallel()
		api := &testEndpoints{Logger: slog.Default(),
			AllowRegistration: true,
			Version:           "v1",
			Commit:            "c1",
		}

		h := ClientConfiguration{
			Current: func() httpauth.BrowserState {
				return httpauth.BrowserState{Mode: httpauth.ModeLocal, AllowRegistration: api.AllowRegistration, LocalLoginEnabled: true}
			},
			Version: api.Version,
			Commit:  api.Commit,
		}.Handler()
		req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)

		var payload configResponse
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&payload))
		assert.False(t, payload.AuthHeaderEnabled)
		assert.True(t, payload.AllowRegistration)
		assert.Equal(t, "v1", payload.Version)
		assert.Equal(t, "c1", payload.Commit)
	})

	t.Run("Returns process-local setup state", func(t *testing.T) {
		t.Parallel()
		h := ClientConfiguration{
			SetupRequired: func() bool { return true },
		}.Handler()
		req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		var payload configResponse
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&payload))
		assert.True(t, payload.SetupRequired)
	})
}

func TestCurrentUser(t *testing.T) {
	t.Parallel()
	t.Run("Returns current user", func(t *testing.T) {
		t.Parallel()
		h := CurrentUser()
		req := httptest.NewRequest(http.MethodGet, "/api/users/current", nil)
		req = httpauth.WithUser(req, db.User{ID: "user@example.com"})
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)

		var payload db.User
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&payload))
		assert.Equal(t, "user@example.com", payload.ID)
	})

	t.Run("Rejects missing user", func(t *testing.T) {
		t.Parallel()
		h := CurrentUser()
		req := httptest.NewRequest(http.MethodGet, "/api/users/current", nil)
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}
