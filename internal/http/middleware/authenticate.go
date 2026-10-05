package middleware

import (
	"net/http"

	httpauth "github.com/gi8lino/motus/internal/http/auth"
	"github.com/gi8lino/motus/internal/http/response"
)

// Authenticate resolves an authenticated principal for protected API routes.
func Authenticate(authenticator httpauth.Authenticator) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := authenticator.Authenticate(r)
			if err != nil {
				response.Problem(w, http.StatusUnauthorized, "Authentication required.")
				return
			}

			next.ServeHTTP(w, httpauth.WithUser(r, user))
		})
	}
}
