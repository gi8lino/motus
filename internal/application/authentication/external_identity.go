// Package authentication contains application-level authentication policies.
package authentication

import (
	"context"
	"strings"

	"github.com/gi8lino/motus/internal/db"
)

// ExternalIdentityRepository persists native-user provider mappings.
type ExternalIdentityRepository interface {
	// ResolveExternalIdentity resolves or provisions one external identity mapping.
	ResolveExternalIdentity(context.Context, string, string, string, string, bool) (*db.User, error)
}

// ExternalIdentity applies Motus registration policy before persistence.
type ExternalIdentity struct {
	// repository atomically resolves provider mappings.
	repository ExternalIdentityRepository
	// allowRegistration controls first-seen external identity provisioning.
	allowRegistration bool
}

// NewExternalIdentity constructs external-identity login policy.
func NewExternalIdentity(repository ExternalIdentityRepository, allowRegistration bool) *ExternalIdentity {
	return &ExternalIdentity{repository: repository, allowRegistration: allowRegistration}
}

// Login resolves a verified provider identity to one native Motus user.
func (s *ExternalIdentity) Login(ctx context.Context, provider, subject, email, name string) (*db.User, error) {
	return s.repository.ResolveExternalIdentity(
		ctx,
		strings.TrimSpace(provider),
		strings.TrimSpace(subject),
		strings.TrimSpace(email),
		strings.TrimSpace(name),
		s.allowRegistration,
	)
}
