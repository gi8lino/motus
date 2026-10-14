// Package httpauth adapts application identities and local persistence to shared authentication.
package httpauth

import (
	"context"
	"net/http"

	appkitauth "github.com/containeroo/appkit/auth"
	appkithttpauth "github.com/containeroo/appkit/httpauth"
	"github.com/gi8lino/motus/internal/db"
)

var (
	// ErrUnauthenticated is a missing or expired identity.
	ErrUnauthenticated = appkitauth.ErrUnauthenticated
	// ErrInvalidCredentials denotes invalid supplied credentials.
	ErrInvalidCredentials = appkitauth.ErrInvalidCredentials
	// ErrRegistrationDisabled means a new user cannot be provisioned.
	ErrRegistrationDisabled = appkitauth.ErrForbidden
)

// Authenticator resolves a Motus user from an HTTP request.
type Authenticator = appkitauth.Authenticator[db.User]

// User returns an authenticated user from a request.
func User(r *http.Request) (db.User, bool) {
	return appkithttpauth.CurrentUser[db.User](r)
}

// WithUser attaches the authenticated user to request context.
func WithUser(r *http.Request, user db.User) *http.Request {
	return appkithttpauth.WithUser(r, user)
}

// ContextUser returns the authenticated user in a context.
func ContextUser(ctx context.Context) (db.User, bool) {
	return appkithttpauth.CurrentUserFromContext[db.User](ctx)
}
