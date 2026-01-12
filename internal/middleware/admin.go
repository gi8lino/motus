package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gi8lino/motus/internal/db"
	"github.com/gi8lino/motus/internal/utils"
)

// adminGetter describes the user lookup required by RequireAdmin.
type adminGetter interface {
	// Get fetches a user for admin checks.
	Get(ctx context.Context, id string) (*db.User, error)
}

// RequireAdmin blocks requests without an admin user (looked up by X-User-ID).
// This is a lightweight guard; replace with real auth for production.
func RequireAdmin(store adminGetter, authHeader string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := utils.DefaultIfZero(authHeader, "X-User-ID")

			userID := strings.TrimSpace(r.Header.Get(header))
			if userID == "" {
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte("forbidden"))
				return
			}

			user, err := store.Get(r.Context(), userID)
			if err != nil || user == nil || !user.IsAdmin {
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte("forbidden"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
