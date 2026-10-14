package httpauth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	appkitauth "github.com/containeroo/appkit/auth"
	"github.com/gi8lino/motus/internal/db"
	"github.com/stretchr/testify/require"
)

type testProxyLogin struct{}

func (testProxyLogin) Login(_ context.Context, user, email, _ string, _, _ bool) (*db.User, error) {
	return &db.User{ID: user, Name: email}, nil
}

func TestProxyRejectsUntrustedPeers(t *testing.T) {
	provider, err := NewTrustedProxy(testProxyLogin{}, TrustedProxyHeaders{
		Username: []string{"X-Forwarded-User"},
		Email:    []string{"X-Forwarded-Email"},
	}, []string{"127.0.0.1/32"})
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
	_, err := NewTrustedProxy(testProxyLogin{}, TrustedProxyHeaders{Username: []string{"X-Forwarded-User"}}, nil)
	require.ErrorContains(t, err, "trusted-proxy-cidrs")
	_, err = NewTrustedProxy(testProxyLogin{}, TrustedProxyHeaders{Username: []string{"X-Forwarded-User"}}, []string{"not-a-network"})
	require.ErrorContains(t, err, "invalid trusted proxy CIDR")
}
