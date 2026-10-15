package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	appkitauth "github.com/containeroo/appkit/auth"
	appkithttpauth "github.com/containeroo/appkit/httpauth"
	"github.com/gi8lino/motus/internal/db"
	"github.com/stretchr/testify/require"
)

// stubAuthenticator adapts authentication results for middleware response tests.
type stubAuthenticator func(*http.Request) (db.User, error)

// Authenticate implements the shared appkit authenticator contract.
func (fn stubAuthenticator) Authenticate(r *http.Request) (db.User, error) {
	return fn(r)
}

// TestAuthenticatePropagatesPrincipal keeps appkit's authenticated-user context contract intact.
func TestAuthenticatePropagatesPrincipal(t *testing.T) {
	t.Parallel()
	middleware := Authenticate(stubAuthenticator(func(*http.Request) (db.User, error) {
		return db.User{ID: "alice@example.com"}, nil
	}))
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := appkithttpauth.CurrentUser[db.User](r)
		require.True(t, ok)
		require.Equal(t, "alice@example.com", user.ID)
		w.WriteHeader(http.StatusNoContent)
	}))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/me", nil))
	require.Equal(t, http.StatusNoContent, recorder.Code)
}

// TestAuthenticateClassifiesFailures distinguishes expected denials from infrastructure failures.
func TestAuthenticateClassifiesFailures(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		err     error
		status  int
		message string
	}{
		{"no identity", appkitauth.ErrUnauthenticated, http.StatusUnauthorized, "Authentication required."},
		{"invalid credentials", appkitauth.ErrInvalidCredentials, http.StatusUnauthorized, "Authentication required."},
		{"registration disabled", fmt.Errorf("%w: %w", appkitauth.ErrForbidden, db.ErrIdentityRegistrationDisabled), http.StatusForbidden, "Login is not permitted."},
		{"existing local user", fmt.Errorf("%w: %w", appkitauth.ErrForbidden, db.ErrIdentityEmailInUse), http.StatusForbidden, "The external identity must be linked to your existing account by an administrator."},
		{"missing provider email", fmt.Errorf("%w: %w", appkitauth.ErrForbidden, db.ErrIdentityEmailInvalid), http.StatusForbidden, "Login is not permitted."},
		{"database failure", errors.New("private database connection details"), http.StatusInternalServerError, "The request could not be processed. Reference:"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			middleware := Authenticate(stubAuthenticator(func(*http.Request) (db.User, error) {
				return db.User{}, tc.err
			}))
			handler := middleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				t.Fatal("request must not reach the endpoint")
			}))
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/me", nil))
			require.Equal(t, tc.status, recorder.Code)
			require.Contains(t, recorder.Body.String(), tc.message)
			require.NotContains(t, recorder.Body.String(), "private database connection details")
		})
	}
}
