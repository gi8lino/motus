package httpauth

import (
	"context"
	"net/http"

	appkitauth "github.com/containeroo/appkit/auth"
	appkitproxy "github.com/containeroo/appkit/auth/proxy"
	"github.com/gi8lino/motus/internal/db"
	"github.com/gi8lino/motus/internal/utils"
)

// TrustedProxy delegates header parsing and ingress verification to appkit.
type TrustedProxy struct {
	provider appkitproxy.Authenticator[db.User]
}

// NewTrustedProxy creates a fail-closed authenticator for explicitly allowed peers.
func NewTrustedProxy(loginService trustedProxyLoginService, settings appkitproxy.Config) (*TrustedProxy, error) {
	provider, err := appkitproxy.New(settings, "trusted-proxy", func(ctx context.Context, identity appkitauth.ExternalIdentity) (db.User, error) {
		email, err := utils.NormalizeEmail(identity.Email)
		if err != nil {
			return db.User{}, appkitauth.ErrInvalidCredentials
		}
		user, err := loginService.Login(ctx, identity.Username, email, identity.DisplayName, false, false)
		if err != nil {
			return db.User{}, err
		}
		if user == nil {
			return db.User{}, appkitauth.ErrUnauthenticated
		}
		return *user, nil
	})
	if err != nil {
		return nil, err
	}
	return &TrustedProxy{provider: provider}, nil
}

// Authenticate verifies the immediate peer before trusting any asserted identity.
func (p *TrustedProxy) Authenticate(r *http.Request) (db.User, error) {
	return p.provider.Authenticate(r)
}

var _ appkitauth.Authenticator[db.User] = (*TrustedProxy)(nil)
