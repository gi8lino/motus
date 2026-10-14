// Package httpauth contains HTTP authentication implementations and contracts.
package httpauth

import (
	"context"
	"errors"
	"net/http"

	"github.com/gi8lino/motus/internal/db"
)

var (
	// ErrUnauthenticated indicates that a request has no valid authenticated identity.
	ErrUnauthenticated = errors.New("unauthenticated")
	// ErrInvalidCredentials indicates that local credentials could not be verified.
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrRegistrationDisabled indicates that new identity provisioning is disabled.
	ErrRegistrationDisabled = errors.New("user registration is disabled")
)

// Authenticator resolves the complete authenticated Motus user for a request.
type Authenticator interface {
	Authenticate(*http.Request) (db.User, error)
}

type contextKey struct{}

// User returns the authenticated user attached by authentication middleware.
func User(r *http.Request) (user db.User, ok bool) {
	user, ok = r.Context().Value(contextKey{}).(db.User)
	return user, ok
}

// WithUser returns a request containing user as its authenticated identity.
func WithUser(r *http.Request, user db.User) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), contextKey{}, user))
}

// ContextUser returns the authenticated user from a request context.
func ContextUser(ctx context.Context) (user db.User, ok bool) {
	user, ok = ctx.Value(contextKey{}).(db.User)
	return user, ok
}
