package middleware

import (
	"errors"
	"net/http"

	appkithttpauth "github.com/containeroo/appkit/httpauth"
	"github.com/gi8lino/motus/internal/db"
	httpauth "github.com/gi8lino/motus/internal/http/auth"
	"github.com/gi8lino/motus/internal/http/response"
)

// Authenticate resolves the principal and distinguishes denial from infrastructure errors.
func Authenticate(authenticator httpauth.Authenticator) func(http.Handler) http.Handler {
	return appkithttpauth.Require[db.User](func(w http.ResponseWriter, r *http.Request, status int) {
		switch status {
		case http.StatusUnauthorized:
			response.Problem(w, status, "Authentication required.")
		case http.StatusForbidden:
			response.Problem(w, status, "Forbidden.")
		default:
			response.InternalServerError(nil, w, errAuthenticationUnavailable)
		}
	}, authenticator)
}

var errAuthenticationUnavailable = errors.New("authentication service unavailable")
