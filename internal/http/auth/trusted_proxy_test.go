package httpauth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	appkitauth "github.com/containeroo/appkit/auth"
	appkitproxy "github.com/containeroo/appkit/auth/proxy"
	"github.com/gi8lino/motus/internal/db"
	"github.com/stretchr/testify/require"
)

type testProxyLogin struct{}

func (testProxyLogin) Login(_ context.Context, user, email, _ string, _, _ bool) (*db.User, error) {
	return &db.User{ID: user, Name: email}, nil
}

func TestProxyRejectsUntrustedPeers(t *testing.T) {
	provider, err := NewTrustedProxy(testProxyLogin{}, appkitproxy.Config{
		UsernameHeaders: []string{"X-Forwarded-User"},
		EmailHeaders:    []string{"X-Forwarded-Email"},
		TrustedCIDRs:    []string{"127.0.0.1/32"},
	})
	require.NoError(t, err)
	request := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	request.Header.Set("X-Forwarded-User", "alice")
	request.Header.Set("X-Forwarded-Email", "alice@example.com")
	request.RemoteAddr = "192.0.2.10:54321"
	_, err = provider.Authenticate(request)
	require.ErrorIs(t, err, appkitauth.ErrUnauthenticated)
	request.RemoteAddr = "127.0.0.1:54321"
	user, err := provider.Authenticate(request)
	require.NoError(t, err)
	require.Equal(t, "alice", user.ID)
}

func TestProxyRequiresExplicitTrustedCIDR(t *testing.T) {
	_, err := NewTrustedProxy(testProxyLogin{}, appkitproxy.Config{UsernameHeaders: []string{"X-Forwarded-User"}})
	require.ErrorContains(t, err, "trusted-proxy-cidrs")
	_, err = NewTrustedProxy(testProxyLogin{}, appkitproxy.Config{
		UsernameHeaders: []string{"X-Forwarded-User"}, TrustedCIDRs: []string{"not-a-network"},
	})
	require.ErrorContains(t, err, "invalid trusted proxy CIDR")
}

// TestProxyWildcardTrust verifies that wildcard trust still requires a valid immediate peer.
func TestProxyWildcardTrust(t *testing.T) {
	provider, err := NewTrustedProxy(testProxyLogin{}, appkitproxy.Config{
		UsernameHeaders: []string{"X-Forwarded-User"},
		EmailHeaders:    []string{"X-Forwarded-Email"},
		TrustedCIDRs:    []string{"*"},
	})
	require.NoError(t, err)

	request := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	request.Header.Set("X-Forwarded-User", "alice")
	request.Header.Set("X-Forwarded-Email", "alice@example.com")

	for _, peer := range []string{"192.0.2.10:54321", "[2001:db8::1]:54321"} {
		request.RemoteAddr = peer
		user, err := provider.Authenticate(request)
		require.NoError(t, err)
		require.Equal(t, "alice", user.ID)
	}

	request.RemoteAddr = "malformed"
	_, err = provider.Authenticate(request)
	require.ErrorIs(t, err, appkitauth.ErrUnauthenticated)
}

// TestProxyRejectsMixedWildcard ensures configuration mistakes cannot broaden trusted peers.
func TestProxyRejectsMixedWildcard(t *testing.T) {
	_, err := NewTrustedProxy(testProxyLogin{}, appkitproxy.Config{
		UsernameHeaders: []string{"X-Forwarded-User"},
		TrustedCIDRs:    []string{"*", "127.0.0.1/32"},
	})
	require.ErrorContains(t, err, "alone")
}

// TestProxyEmailOnlyIdentity verifies that Motus delegates email-only identities to appkit.
func TestProxyEmailOnlyIdentity(t *testing.T) {
	provider, err := NewTrustedProxy(testProxyLogin{}, appkitproxy.Config{
		TrustedCIDRs: []string{"*"}, EmailHeaders: []string{"X-Auth-Request-Email"},
	})
	require.NoError(t, err)
	request := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	request.RemoteAddr = "192.0.2.10:443"
	request.Header.Set("X-Auth-Request-Email", "alice@example.com")
	request.Header.Set("X-Forwarded-User", "forged")
	user, err := provider.Authenticate(request)
	require.NoError(t, err)
	require.Equal(t, "alice@example.com", user.ID)
}
