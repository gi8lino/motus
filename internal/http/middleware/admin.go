package middleware

import (
	"context"
	"net/http"
	"strings"

	appkithttpauth "github.com/containeroo/appkit/httpauth"
	"github.com/gi8lino/motus/internal/db"
	httpauth "github.com/gi8lino/motus/internal/http/auth"
	"github.com/gi8lino/motus/internal/http/response"
)

// adminGetter provides current persisted permissions rather than stale session roles.
type adminGetter interface {
	Get(context.Context, string) (*db.User, error)
}

// RequireAdmin applies the shared authorization middleware using current user permissions.
// fixtureHeader is only an explicit test-fixture input and is never set on production routes.
func RequireAdmin(store adminGetter, fixtureHeader ...string) Middleware {
	onFailure := func(w http.ResponseWriter, _ *http.Request, _ int) {
		response.Problem(w, http.StatusForbidden, "Forbidden.")
	}
	authorize := appkithttpauth.Authorize[db.User](func(r *http.Request, principal db.User) bool {
		if principal.ID == "" {
			return false
		}
		current, err := store.Get(r.Context(), principal.ID)
		return err == nil && current != nil && current.IsAdmin
	}, onFailure)
	return func(next http.Handler) http.Handler {
		protected := authorize(next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := httpauth.User(r); !ok && len(fixtureHeader) == 1 {
				name := fixtureHeader[0]
				if name == "" {
					name = "X-User-ID"
				}
				if id := strings.TrimSpace(r.Header.Get(name)); id != "" {
					r = httpauth.WithUser(r, db.User{ID: id})
				}
			}
			protected.ServeHTTP(w, r)
		})
	}
}
