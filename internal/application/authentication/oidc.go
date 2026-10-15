package authentication

import (
	"context"
	"errors"
	"fmt"

	"github.com/containeroo/appkit/auth"
	"github.com/containeroo/appkit/auth/provisioning"
	"github.com/gi8lino/motus/internal/db"
	"github.com/jackc/pgx/v5"
)

// oidcRepository contains persistence required for OIDC login resolution.
type oidcRepository interface {
	AuthenticationSettings(context.Context) (db.AuthenticationSettings, error)
	GetIdentityUser(context.Context, string, string) (*db.User, error)
	ResolveExternalIdentity(context.Context, string, string, string, string, bool) (*db.User, error)
}

// OIDC coordinates verified OIDC identity login policy with persistence.
type OIDC struct {
	resolver provisioning.Resolver[*db.User]
}

// NewOIDC constructs the OIDC login application service.
func NewOIDC(repository oidcRepository, registrationOverride *bool) *OIDC {
	return &OIDC{
		resolver: provisioning.Resolver[*db.User]{
			Registration: provisioning.RegistrationPolicy{
				Override: registrationOverride,
				Load: func(ctx context.Context) (bool, error) {
					settings, err := repository.AuthenticationSettings(ctx)
					return settings.AllowRegistration, err
				},
			},
			IsMissing: func(err error) bool { return errors.Is(err, pgx.ErrNoRows) },
			Lookup: func(ctx context.Context, identity provisioning.Identity) (*db.User, error) {
				return repository.GetIdentityUser(ctx, "oidc:"+identity.Issuer, identity.Subject)
			},
			Resolve: func(ctx context.Context, identity provisioning.Identity, registrationEnabled bool) (*db.User, error) {
				user, err := repository.ResolveExternalIdentity(ctx, "oidc:"+identity.Issuer, identity.Subject, identity.Email, identity.DisplayName, registrationEnabled)
				return user, classifyExternalIdentityError(err)
			},
		},
	}
}

// Login normalizes a verified OIDC identity and resolves registration policy before persistence.
func (s *OIDC) Login(ctx context.Context, issuer, subject, username, email, displayName string) (*db.User, error) {
	user, err := s.resolver.Login(ctx, provisioning.Identity{
		Issuer: issuer, Subject: subject, Username: username, Email: email, DisplayName: displayName,
	})
	if errors.Is(err, provisioning.ErrIncompleteIdentity) {
		return nil, fmt.Errorf("%w: the identity provider must supply issuer, subject, and username", auth.ErrInvalidCredentials)
	}
	return user, err
}
