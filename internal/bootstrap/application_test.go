package bootstrap

import (
	"testing"

	appkitauth "github.com/containeroo/appkit/auth"
	"github.com/stretchr/testify/require"

	"github.com/gi8lino/motus/internal/flags"
)

func TestAuthenticationModeOverride(t *testing.T) {
	require.Empty(t, authenticationMode(nil))
	mode := appkitauth.ModeNone
	require.Equal(t, "none", string(authenticationMode(&mode)))
	mode = appkitauth.ModeProxy
	require.Equal(t, "proxy", string(authenticationMode(&mode)))
}

// TestNewBrowserConfigPreservesProxyDefaults verifies that the appkit defaults
// reach browser authentication even when no environment override is set.
func TestNewBrowserConfigPreservesProxyDefaults(t *testing.T) {
	cfg, err := flags.Parse([]string{"--database-url", "postgres://localhost/motus", "--auth-mode", "proxy", "--trusted-proxy-cidrs", "*"}, "")
	require.NoError(t, err)

	browser := newBrowserConfig(cfg)
	require.Equal(t, "proxy", string(browser.ModeOverride))
	require.Contains(t, browser.ProxyDefaults.Email, "X-Auth-Request-Email")
	require.Contains(t, browser.ProxyDefaults.Username, "X-Forwarded-User")
	require.Nil(t, browser.ProxyOverrides.Username)
	require.Nil(t, browser.ProxyOverrides.Email)
}
