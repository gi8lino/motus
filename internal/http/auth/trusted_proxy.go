package httpauth

import (
	"github.com/gi8lino/motus/internal/db"
	"github.com/gi8lino/motus/internal/utils"
	"net/http"
	"strings"
)

// TrustedProxy accepts identities supplied by a reverse proxy that strips and sets Header.
type TrustedProxy struct {
	repository        identityRepository
	header            string
	allowRegistration bool
}

// NewTrustedProxy creates a trusted-proxy authenticator backed by provider mappings.
func NewTrustedProxy(repository identityRepository, header string, allowRegistration bool) *TrustedProxy {
	return &TrustedProxy{repository: repository, header: header, allowRegistration: allowRegistration}
}

// Authenticate resolves the current reverse-proxy identity to a Motus user.
func (a *TrustedProxy) Authenticate(r *http.Request) (db.User, error) {
	email, err := utils.NormalizeEmail(strings.TrimSpace(r.Header.Get(a.header)))
	if err != nil {
		return db.User{}, ErrUnauthenticated
	}
	user, err := a.repository.ResolveExternalIdentity(r.Context(), "trusted-proxy", email, email, email, a.allowRegistration)
	if err != nil || user == nil {
		return db.User{}, ErrUnauthenticated
	}
	return *user, nil
}

var _ Authenticator = (*TrustedProxy)(nil)
