package httpauth

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"strings"

	appkitauth "github.com/containeroo/appkit/auth"
	appkitproxy "github.com/containeroo/appkit/auth/proxy"
	"github.com/gi8lino/motus/internal/db"
	"github.com/gi8lino/motus/internal/utils"
)

// TrustedProxyHeaders lists trusted identity headers configured by the administrator.
type TrustedProxyHeaders struct {
	Username    []string
	Email       []string
	DisplayName []string
	Groups      []string
	AdminGroup  string
}

// TrustedProxy delegates header parsing and ingress verification to appkit.
type TrustedProxy struct {
	provider appkitproxy.Authenticator[db.User]
}

// NewTrustedProxy creates a fail-closed authenticator for explicitly allowed peers.
func NewTrustedProxy(loginService trustedProxyLoginService, headers TrustedProxyHeaders, cidrs []string) (*TrustedProxy, error) {
	if len(cidrs) == 0 {
		return nil, errors.New("trusted-proxy authentication requires --trusted-proxy-cidrs")
	}
	peers := make([]netip.Prefix, 0, len(cidrs))
	for _, value := range cidrs {
		parsed, err := netip.ParsePrefix(strings.TrimSpace(value))
		if err != nil {
			return nil, fmt.Errorf("invalid trusted proxy CIDR %q: %w", value, err)
		}
		peers = append(peers, parsed.Masked())
	}
	// A proxy can supply a username or an email-based stable identity.
	usernames := append(append([]string{}, headers.Username...), headers.Email...)
	provider := appkitproxy.Authenticator[db.User]{
		Issuer:             "trusted-proxy",
		UsernameHeaders:    usernames,
		EmailHeaders:       headers.Email,
		DisplayNameHeaders: headers.DisplayName,
		GroupHeaders:       headers.Groups,
		Trust: func(r *http.Request) bool {
			host, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				return false
			}
			addr, err := netip.ParseAddr(host)
			if err != nil {
				return false
			}
			addr = addr.Unmap()
			for _, prefix := range peers {
				if prefix.Contains(addr) {
					return true
				}
			}
			return false
		},
		Resolve: func(ctx context.Context, identity appkitauth.ExternalIdentity) (db.User, error) {
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
		},
	}
	if err := provider.Validate(); err != nil {
		return nil, err
	}
	return &TrustedProxy{provider: provider}, nil
}

// Authenticate verifies the immediate peer before trusting any asserted identity.
func (p *TrustedProxy) Authenticate(r *http.Request) (db.User, error) {
	return p.provider.Authenticate(r)
}

var _ appkitauth.Authenticator[db.User] = (*TrustedProxy)(nil)
