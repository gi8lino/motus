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

// proxyLoginRepository simulates persisted identities and records registration policy decisions.
type proxyLoginRepository struct {
	settings   db.AuthenticationSettings
	lookupErr  error
	resolveErr error
	allowSeen  *bool
}

// AuthenticationSettings reads the stored registration flag for an unset deployment override.
func (r *proxyLoginRepository) AuthenticationSettings(context.Context) (db.AuthenticationSettings, error) {
	return r.settings, nil
}

// GetIdentityUser simulates a previously unseen external identity.
func (r *proxyLoginRepository) GetIdentityUser(context.Context, string, string) (*db.User, error) {
	if r.lookupErr != nil {
		return nil, r.lookupErr
	}
	return nil, pgx.ErrNoRows
}

// ResolveExternalIdentity records whether provisioning was permitted.
func (r *proxyLoginRepository) ResolveExternalIdentity(_ context.Context, _, _, _, _ string, allow bool) (*db.User, error) {
	r.allowSeen = &allow
	if r.resolveErr != nil {
		return nil, r.resolveErr
	}
	return &db.User{ID: "alice@example.com"}, nil
}

// TestProxyRegistrationOverrideAndExistingAccountConflict covers Envoy users with preexisting local accounts.
func TestProxyRegistrationOverrideAndExistingAccountConflict(t *testing.T) {
	t.Parallel()
	repo := &proxyLoginRepository{
		settings:   db.AuthenticationSettings{AllowRegistration: false},
		resolveErr: db.ErrIdentityEmailInUse,
	}
	allow := true
	_, err := NewTrustedProxy(repo, &allow).Login(context.Background(), "alice@example.com", "alice@example.com", "Alice", false, false)
	require.ErrorIs(t, err, appkitauth.ErrForbidden)
	require.ErrorIs(t, err, db.ErrIdentityEmailInUse)
	require.NotNil(t, repo.allowSeen)
	require.True(t, *repo.allowSeen, "the deployment override must win over the persisted false value")
}

// TestProxyRegistrationDisabledIsNotInfrastructureFailure recognizes an expected provisioning denial.
func TestProxyRegistrationDisabledIsNotInfrastructureFailure(t *testing.T) {
	t.Parallel()
	repo := &proxyLoginRepository{resolveErr: db.ErrIdentityRegistrationDisabled}
	_, err := NewTrustedProxy(repo, nil).Login(context.Background(), "alice@example.com", "alice@example.com", "", false, false)
	require.ErrorIs(t, err, appkitauth.ErrForbidden)
	require.ErrorIs(t, err, db.ErrIdentityRegistrationDisabled)
	require.NotNil(t, repo.allowSeen)
	require.False(t, *repo.allowSeen)
}

// TestProxyDatabaseErrorIsPreserved returns the underlying infrastructure failure without reclassifying it.
func TestProxyDatabaseErrorIsPreserved(t *testing.T) {
	t.Parallel()
	underlying := errors.New("database unavailable")
	repo := &proxyLoginRepository{resolveErr: underlying}
	allow := true
	_, err := NewTrustedProxy(repo, &allow).Login(context.Background(), "alice@example.com", "alice@example.com", "", false, false)
	require.ErrorIs(t, err, underlying)
	require.NotErrorIs(t, err, appkitauth.ErrForbidden)
}
