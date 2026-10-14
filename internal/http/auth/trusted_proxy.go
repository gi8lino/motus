package httpauth

import (
	"net/http"
	"strings"

	"github.com/gi8lino/motus/internal/db"
	"github.com/gi8lino/motus/internal/utils"
)

// TrustedProxy accepts identities supplied by a reverse proxy that strips and sets Header.
type TrustedProxy struct {
	loginService trustedProxyLoginService
	headers      TrustedProxyHeaders
}

// TrustedProxyHeaders identifies configured trusted reverse-proxy identity headers.
type TrustedProxyHeaders struct {
	// Username lists headers that may contain the stable proxy identity.
	Username []string
	// Email lists headers that may contain the asserted email address.
	Email []string
	// DisplayName lists headers that may contain the asserted display name.
	DisplayName []string
	// Groups lists headers that may contain asserted external groups.
	Groups []string
	// AdminGroup grants administrator access when present in Groups.
	AdminGroup string
}

// NewTrustedProxy creates a trusted-proxy authenticator backed by provider mappings.
func NewTrustedProxy(loginService trustedProxyLoginService, headers TrustedProxyHeaders) *TrustedProxy {
	return &TrustedProxy{loginService: loginService, headers: headers}
}

// Authenticate resolves the current reverse-proxy identity to a Motus user.
func (a *TrustedProxy) Authenticate(r *http.Request) (db.User, error) {
	username := firstConfiguredHeaderValue(r, a.headers.Username)
	email, err := utils.NormalizeEmail(firstConfiguredHeaderValue(r, a.headers.Email))
	if err != nil {
		return db.User{}, ErrUnauthenticated
	}
	if username == "" {
		username = email
	}
	user, err := a.loginService.Login(r.Context(), username, email, firstConfiguredHeaderValue(r, a.headers.DisplayName), false, false)
	if err != nil || user == nil {
		return db.User{}, ErrUnauthenticated
	}
	return *user, nil
}

// firstConfiguredHeaderValue returns the first non-empty value among headers.
func firstConfiguredHeaderValue(r *http.Request, headers []string) string {
	for _, header := range headers {
		if value := strings.TrimSpace(r.Header.Get(header)); value != "" {
			return value
		}
	}
	return ""
}

var _ Authenticator = (*TrustedProxy)(nil)
