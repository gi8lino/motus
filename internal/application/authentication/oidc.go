package authentication

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/containeroo/appkit/auth"
	"github.com/containeroo/appkit/auth/provisioning"
	"github.com/gi8lino/motus/internal/db"
)

// OIDC resolves verified external identities under the effective registration policy.
type OIDC struct {
	// accounts applies the shared access policy using the Motus directory.
	accounts provisioning.Accounts[*db.User]
	// policy contains the effective deployment and application settings.
	policy provisioning.AccessPolicy
}

// NewOIDC binds the resolved registration setting to the local identity store.
func NewOIDC(repo externalAccountStore, allowRegistration bool) *OIDC {
	return &OIDC{
		accounts: newExternalAccounts(repo),
		policy:   provisioning.AccessPolicy{AllowRegistration: allowRegistration},
	}
}

// Login normalizes a verified OIDC identity and resolves registration policy before persistence.
func (s *OIDC) Login(ctx context.Context, issuer, subject, username, email, displayName string) (*db.User, error) {
	if strings.TrimSpace(issuer) == "" || strings.TrimSpace(username) == "" {
		return nil, auth.ErrInvalidCredentials
	}
	user, err := s.accounts.Login(ctx, s.policy, provisioning.Identity{
		Issuer: "oidc:" + strings.TrimSpace(issuer), Subject: subject, Username: username, Email: email, DisplayName: displayName,
	})
	if errors.Is(err, provisioning.ErrIncompleteIdentity) {
		return nil, fmt.Errorf("%w: the identity provider must supply issuer, subject, and username", auth.ErrInvalidCredentials)
	}
	return user, err
}
