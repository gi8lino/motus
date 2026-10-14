package authentication

import (
	"context"
	"errors"
	"strings"

	"github.com/gi8lino/motus/internal/db"
	"github.com/jackc/pgx/v5"
)

// oidcRepository contains persistence required for OIDC login resolution.
type oidcRepository interface {
	registrationSettingsRepository
	GetIdentityUser(context.Context, string, string) (*db.User, error)
	ResolveExternalIdentity(context.Context, string, string, string, string, bool) (*db.User, error)
}

// OIDC coordinates verified OIDC identity login policy with persistence.
type OIDC struct {
	repository   oidcRepository
	registration registrationPolicy
}

// NewOIDC constructs the OIDC login application service.
func NewOIDC(repository oidcRepository, registrationOverride *bool) *OIDC {
	return &OIDC{
		repository:   repository,
		registration: newRegistrationPolicy(repository, registrationOverride),
	}
}

// Login normalizes a verified OIDC identity and resolves registration policy before persistence.
func (s *OIDC) Login(ctx context.Context, issuer, subject, username, email, displayName string) (*db.User, error) {
	issuer = strings.TrimSpace(issuer)
	subject = strings.TrimSpace(subject)
	username = strings.TrimSpace(username)
	email = strings.TrimSpace(email)
	displayName = strings.TrimSpace(displayName)
	if issuer == "" || subject == "" || username == "" {
		return nil, errors.New("the identity provider must supply issuer, subject, and username")
	}
	provider := "oidc:" + issuer
	allow := false
	if _, err := s.repository.GetIdentityUser(ctx, provider, subject); err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
		var policyErr error
		allow, policyErr = s.registration.enabled(ctx)
		if policyErr != nil {
			return nil, policyErr
		}
	}
	return s.repository.ResolveExternalIdentity(ctx, provider, subject, email, displayName, allow)
}
