package authentication

import (
	"context"
	"errors"
	"strings"

	"github.com/containeroo/appkit/auth/provisioning"
	"github.com/gi8lino/motus/internal/db"
	"github.com/jackc/pgx/v5"
)

// trustedProxyRepository contains persistence required for trusted-proxy login resolution.
type trustedProxyRepository interface {
	AuthenticationSettings(context.Context) (db.AuthenticationSettings, error)
	GetIdentityUser(context.Context, string, string) (*db.User, error)
	ResolveExternalIdentity(context.Context, string, string, string, string, bool) (*db.User, error)
}

// TrustedProxy owns trusted-proxy identity normalization and registration policy.
type TrustedProxy struct {
	repository   trustedProxyRepository
	registration provisioning.RegistrationPolicy
}

// NewTrustedProxy constructs the trusted-proxy login application service.
func NewTrustedProxy(repository trustedProxyRepository, registrationOverride *bool) *TrustedProxy {
	return &TrustedProxy{repository: repository, registration: provisioning.RegistrationPolicy{
		Override: registrationOverride,
		Load: func(ctx context.Context) (bool, error) {
			settings, err := repository.AuthenticationSettings(ctx)
			return settings.AllowRegistration, err
		},
	}}
}

// Login resolves a trusted-proxy identity and provisions unknown identities only when allowed.
func (s *TrustedProxy) Login(ctx context.Context, username, email, displayName string, _ bool, _ bool) (*db.User, error) {
	username, email, displayName = strings.TrimSpace(username), strings.TrimSpace(email), strings.TrimSpace(displayName)
	if username == "" {
		return nil, errors.New("the trusted proxy must supply a username")
	}
	allow := false
	if _, err := s.repository.GetIdentityUser(ctx, "trusted-proxy", username); err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
		var policyErr error
		allow, policyErr = s.registration.Enabled(ctx)
		if policyErr != nil {
			return nil, policyErr
		}
	}
	return s.repository.ResolveExternalIdentity(ctx, "trusted-proxy", username, email, displayName, allow)
}
