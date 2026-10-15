package middleware

import (
	"errors"
	"net/http"

	appkithttpauth "github.com/containeroo/appkit/httpauth"
	"github.com/gi8lino/motus/internal/db"
	httpauth "github.com/gi8lino/motus/internal/http/auth"
	"github.com/gi8lino/motus/internal/http/response"
)

// Authenticate resolves the principal and reports the original cause of unexpected failures.
func Authenticate(authenticator httpauth.Authenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := authenticator.Authenticate(r)
			if err == nil {
				next.ServeHTTP(w, appkithttpauth.WithUser(r, user))
				return
			}
			switch {
			case errors.Is(err, httpauth.ErrUnauthenticated), errors.Is(err, httpauth.ErrInvalidCredentials):
				response.Problem(w, http.StatusUnauthorized, "Authentication required.")
			case errors.Is(err, db.ErrIdentityEmailInUse):
				response.Problem(w, http.StatusForbidden, "The external identity must be linked to your existing account by an administrator.")
			case errors.Is(err, httpauth.ErrRegistrationDisabled):
				response.Problem(w, http.StatusForbidden, "Login is not permitted.")
			default:
				response.InternalServerError(nil, w, err)
			}
		})
	}
}
