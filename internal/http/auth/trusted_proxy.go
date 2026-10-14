package httpauth

import (
	"net/http"
	"strings"

	"github.com/gi8lino/motus/internal/db"
	"github.com/gi8lino/motus/internal/utils"
)

// TrustedProxy accepts identities supplied by a reverse proxy that strips and sets Header.
type TrustedProxy struct {
	loginService externalLoginService
	header       string
}

// NewTrustedProxy creates a trusted-proxy authenticator backed by provider mappings.
func NewTrustedProxy(loginService externalLoginService, header string) *TrustedProxy {
	return &TrustedProxy{loginService: loginService, header: header}
}

// Authenticate resolves the current reverse-proxy identity to a Motus user.
func (a *TrustedProxy) Authenticate(r *http.Request) (db.User, error) {
	email, err := utils.NormalizeEmail(strings.TrimSpace(r.Header.Get(a.header)))
	if err != nil {
		return db.User{}, ErrUnauthenticated
	}
	user, err := a.loginService.Login(r.Context(), "trusted-proxy", email, email, email)
	if err != nil || user == nil {
		return db.User{}, ErrUnauthenticated
	}
	return *user, nil
}

var _ Authenticator = (*TrustedProxy)(nil)
