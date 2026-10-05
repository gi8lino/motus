package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gi8lino/motus/internal/db"
	httpauth "github.com/gi8lino/motus/internal/http/auth"
	"github.com/gi8lino/motus/internal/http/response"
)

// adminGetter describes the user lookup required by RequireAdmin.
type adminGetter interface {
	// Get fetches a user for admin checks.
	Get(ctx context.Context, id string) (*db.User, error)
}

// RequireAdmin blocks requests whose authenticated principal is not an administrator.
func RequireAdmin(store adminGetter, fixtureHeader ...string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			principal, _ := httpauth.ContextUser(r.Context())
			userID := principal.ID
			if userID == "" && len(fixtureHeader) == 1 {
				header := fixtureHeader[0]
				if header == "" {
					header = "X-User-ID"
				}
				userID = strings.TrimSpace(r.Header.Get(header))
			}
			if userID == "" {
				response.Problem(w, http.StatusForbidden, "Forbidden.")
				return
			}

			user, err := store.Get(r.Context(), userID)
			if err != nil || user == nil || !user.IsAdmin {
				response.Problem(w, http.StatusForbidden, "Forbidden.")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
