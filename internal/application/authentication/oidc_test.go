package authentication

import (
	"context"
	"errors"
	"testing"

	appkitauth "github.com/containeroo/appkit/auth"
	"github.com/gi8lino/motus/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

// oidcLoginRepository simulates an unlinked external identity without database dependencies.
type oidcLoginRepository struct {
	resolveErr error
}

// AuthenticationSettings supplies persisted registration settings.
func (*oidcLoginRepository) AuthenticationSettings(context.Context) (db.AuthenticationSettings, error) {
	return db.AuthenticationSettings{AllowRegistration: false}, nil
}

// GetIdentityUser reports a new OIDC identity.
func (*oidcLoginRepository) GetIdentityUser(context.Context, string, string) (*db.User, error) {
	return nil, pgx.ErrNoRows
}

// ResolveExternalIdentity returns the configured result from the external account store.
func (r *oidcLoginRepository) ResolveExternalIdentity(context.Context, string, string, string, string, bool) (*db.User, error) {
	return nil, r.resolveErr
}

// TestOIDCIdentityConflictIsForbidden ensures OIDC uses the same typed denial semantics as proxy mode.
func TestOIDCIdentityConflictIsForbidden(t *testing.T) {
	t.Parallel()
	allow := true
	repo := &oidcLoginRepository{resolveErr: db.ErrIdentityEmailInUse}
	_, err := NewOIDC(repo, &allow).Login(context.Background(), "https://sso.example/", "subject-1", "alice", "alice@example.com", "Alice")
	require.ErrorIs(t, err, appkitauth.ErrForbidden)
	require.ErrorIs(t, err, db.ErrIdentityEmailInUse)
}

// TestOIDCInfrastructureErrorSurfaces preserves unexpected database failures for diagnostics.
func TestOIDCInfrastructureErrorSurfaces(t *testing.T) {
	t.Parallel()
	allow := true
	underlying := errors.New("database unavailable")
	repo := &oidcLoginRepository{resolveErr: underlying}
	_, err := NewOIDC(repo, &allow).Login(context.Background(), "https://sso.example/", "subject-1", "alice", "alice@example.com", "Alice")
	require.ErrorIs(t, err, underlying)
	require.NotErrorIs(t, err, appkitauth.ErrForbidden)
}
