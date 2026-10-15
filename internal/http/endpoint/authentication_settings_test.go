package endpoint

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	appauth "github.com/gi8lino/motus/internal/application/authentication"
	"github.com/gi8lino/motus/internal/db"
	httpauth "github.com/gi8lino/motus/internal/http/auth"
)

// authenticationSettingsStore records authentication-settings persistence in endpoint tests.
type authenticationSettingsStore struct {
	settings db.AuthenticationSettings
	saved    bool
}

// AuthenticationSettings returns the stored test settings.
func (s *authenticationSettingsStore) AuthenticationSettings(context.Context) (db.AuthenticationSettings, error) {
	return s.settings, nil
}

// SaveAuthenticationSettings records a successful persistence attempt.
func (s *authenticationSettingsStore) SaveAuthenticationSettings(_ context.Context, settings db.AuthenticationSettings) error {
	s.settings = settings
	s.saved = true
	return nil
}

// TestAuthenticationSettings verifies runtime validation occurs before persistence.
func TestAuthenticationSettings(t *testing.T) {
	t.Parallel()

	t.Run("Rejects an unappliable configuration before saving", func(t *testing.T) {
		t.Parallel()
		store := &authenticationSettingsStore{}
		handler := AuthenticationSettings(
			appauth.NewConfiguration(store),
			httpauth.BrowserAuth{Validate: func(context.Context, db.AuthenticationSettings) error {
				return errors.New("provider discovery failed")
			}},
		)
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPut, "/authentication", strings.NewReader(`{"mode":"local"}`))

		handler.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
		assert.False(t, store.saved)
	})

	t.Run("Persists empty proxy headers when deployment overrides validate", func(t *testing.T) {
		t.Parallel()
		store := &authenticationSettingsStore{}
		validated := false
		handler := AuthenticationSettings(
			appauth.NewConfiguration(store),
			httpauth.BrowserAuth{Validate: func(_ context.Context, settings db.AuthenticationSettings) error {
				validated = true
				assert.Empty(t, settings.TrustedUsernameHeaders)
				return nil // The deployment supplies the effective headers.
			}},
		)
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPut, "/authentication", strings.NewReader(`{"mode":"proxy"}`))
		handler.ServeHTTP(recorder, request)
		require.Equal(t, http.StatusNoContent, recorder.Code)
		assert.True(t, validated)
		assert.True(t, store.saved)
		assert.Empty(t, store.settings.TrustedUsernameHeaders)
	})

	t.Run("Persists then applies a validated configuration", func(t *testing.T) {
		t.Parallel()
		store := &authenticationSettingsStore{}
		applied := false
		handler := AuthenticationSettings(
			appauth.NewConfiguration(store),
			httpauth.BrowserAuth{
				Validate: func(context.Context, db.AuthenticationSettings) error { return nil },
				ApplySettings: func(db.AuthenticationSettings) {
					applied = true
				},
			},
		)
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPut, "/authentication", strings.NewReader(`{"mode":"local"}`))

		handler.ServeHTTP(recorder, request)

		require.Equal(t, http.StatusNoContent, recorder.Code)
		assert.True(t, store.saved)
		assert.True(t, applied)
	})
}
