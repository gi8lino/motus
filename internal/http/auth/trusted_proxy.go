package httpauth

import (
	appauth "github.com/gi8lino/motus/internal/application/authentication"
	"github.com/gi8lino/motus/internal/db"
	"github.com/gi8lino/motus/internal/utils"
	"net/http"
	"strings"
)

// TrustedProxy accepts identities supplied by a reverse proxy that strips and sets Header.
type TrustedProxy struct {
	repository *appauth.ExternalIdentity
	header     string
}

// NewTrustedProxy creates a trusted-proxy authenticator backed by provider mappings.
func NewTrustedProxy(repository *appauth.ExternalIdentity, header string) *TrustedProxy {
	return &TrustedProxy{repository: repository, header: header}
}

// Authenticate resolves the current reverse-proxy identity to a Motus user.
func (a *TrustedProxy) Authenticate(r *http.Request) (db.User, error) {
	email, err := utils.NormalizeEmail(strings.TrimSpace(r.Header.Get(a.header)))
	if err != nil {
		return db.User{}, ErrUnauthenticated
	}
	user, err := a.repository.Login(r.Context(), "trusted-proxy", email, email, email)
	if err != nil || user == nil {
		return db.User{}, ErrUnauthenticated
	}
	return *user, nil
}

var _ Authenticator = (*TrustedProxy)(nil)
