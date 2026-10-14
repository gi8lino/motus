package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	appkithttpauth "github.com/containeroo/appkit/httpauth"
	"github.com/gi8lino/motus/internal/db"
	httpauth "github.com/gi8lino/motus/internal/http/auth"
	"github.com/gi8lino/motus/internal/http/response"
	errpkg "github.com/gi8lino/motus/internal/service/errors"
)

// adminGetter provides current persisted permissions rather than stale session roles.
type adminGetter interface {
	Get(context.Context, string) (*db.User, error)
}

// RequireAdmin applies the shared authorization middleware using current user permissions.
// fixtureHeader is only an explicit test-fixture input and is never set on production routes.
func RequireAdmin(store adminGetter, fixtureHeader ...string) func(http.Handler) http.Handler {
	onFailure := func(w http.ResponseWriter, _ *http.Request, status int) {
		switch status {
		case http.StatusUnauthorized:
			response.Problem(w, status, "Authentication required.")
		case http.StatusForbidden:
			response.Problem(w, status, "Forbidden.")
		default:
			response.InternalServerError(nil, w, errors.New("authorization service unavailable"))
		}
	}
	authorize := appkithttpauth.AuthorizeWithError[db.User](func(r *http.Request, principal db.User) (bool, error) {
		if principal.ID == "" {
			return false, nil
		}
		current, err := store.Get(r.Context(), principal.ID)
		if errpkg.IsKind(err, errpkg.ErrorNotFound) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		return current != nil && current.IsAdmin, nil
	}, onFailure, func(w http.ResponseWriter, _ *http.Request, err error) { response.InternalServerError(nil, w, err) })
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
