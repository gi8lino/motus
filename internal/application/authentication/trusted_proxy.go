package authentication

import (
	"context"
	"errors"
	"fmt"

	"github.com/containeroo/appkit/auth"
	"github.com/containeroo/appkit/auth/provisioning"
	"github.com/gi8lino/motus/internal/db"
)

// TrustedProxy resolves verified external identities under the effective registration policy.
type TrustedProxy struct {
	// accounts applies the shared access policy using the Motus directory.
	accounts provisioning.Accounts[*db.User]
	// policy contains the effective deployment and application settings.
	policy provisioning.AccessPolicy
}

// NewTrustedProxy binds the resolved registration setting to the local identity store.
func NewTrustedProxy(repo externalAccountStore, allowRegistration bool) *TrustedProxy {
	return &TrustedProxy{
		accounts: newExternalAccounts(repo),
		policy:   provisioning.AccessPolicy{AllowRegistration: allowRegistration},
	}
}

// Login resolves a trusted-proxy identity and provisions unknown identities only when allowed.
func (s *TrustedProxy) Login(ctx context.Context, username, email, displayName string, _ bool, _ bool) (*db.User, error) {
	user, err := s.accounts.Login(
		ctx,
		s.policy,
		provisioning.Identity{
			Issuer:      "trusted-proxy",
			Subject:     username,
			Username:    username,
			Email:       email,
			DisplayName: displayName,
		},
	)
	if errors.Is(err, provisioning.ErrIncompleteIdentity) {
		return nil, fmt.Errorf("%w: the trusted proxy must supply a username", auth.ErrInvalidCredentials)
	}
	return user, err
}
