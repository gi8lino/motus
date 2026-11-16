package authentication

import (
	"context"

	"github.com/containeroo/appkit/auth/provisioning"
	"github.com/gi8lino/motus/internal/db"
)

// externalAccountStore atomically resolves or provisions a linked account.
type externalAccountStore interface {
	ResolveExternalIdentity(context.Context, string, string, string, string, bool) (*db.User, error)
}

// newExternalAccounts adapts verified provider identities to Motus storage keys.
func newExternalAccounts(repo externalAccountStore) provisioning.Accounts[*db.User] {
	return provisioning.Accounts[*db.User]{Resolve: func(ctx context.Context, id provisioning.Identity, allow bool) (*db.User, error) {
		user, err := repo.ResolveExternalIdentity(ctx, id.Issuer, id.Subject, id.Email, id.DisplayName, allow)
		return user, classifyExternalIdentityError(err)
	}}
}
