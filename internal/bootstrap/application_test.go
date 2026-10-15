package bootstrap

import (
	"testing"

	appkitauth "github.com/containeroo/appkit/auth"
	appkitproxy "github.com/containeroo/appkit/auth/proxy"
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

// TestNewBrowserConfigDelegatesProxyDefaults ensures Motus doesn't reimplement appkit defaults.
func TestNewBrowserConfigDelegatesProxyDefaults(t *testing.T) {
	cfg, err := flags.Parse([]string{"--database-url", "postgres://localhost/motus", "--auth-mode", "proxy", "--trusted-proxy-cidrs", "*"}, "")
	require.NoError(t, err)

	browser := newBrowserConfig(cfg)
	require.Equal(t, "proxy", string(browser.ModeOverride))
	require.NotNil(t, browser.ProxyOverrides.TrustedCIDRs)
	require.Equal(t, []string{"*"}, *browser.ProxyOverrides.TrustedCIDRs)
	require.Nil(t, browser.ProxyOverrides.UsernameHeaders)
	require.Nil(t, browser.ProxyOverrides.EmailHeaders)

	resolved, err := appkitproxy.Resolve(appkitproxy.Options{Overrides: browser.ProxyOverrides})
	require.NoError(t, err)
	require.Contains(t, resolved.EmailHeaders, "X-Auth-Request-Email")
	require.Contains(t, resolved.UsernameHeaders, "X-Forwarded-User")
}

// TestNewBrowserConfigRetainsExplicitHeaderOverrides checks CLI settings win over defaults.
func TestNewBrowserConfigRetainsExplicitHeaderOverrides(t *testing.T) {
	cfg, err := flags.Parse([]string{
		"--database-url", "postgres://localhost/motus",
		"--trusted-proxy-cidrs", "127.0.0.1/32",
		"--trusted-email-headers", "X-Verified-Email",
	}, "")
	require.NoError(t, err)

	overrides := newBrowserConfig(cfg).ProxyOverrides
	require.NotNil(t, overrides.EmailHeaders)
	require.Equal(t, []string{"X-Verified-Email"}, *overrides.EmailHeaders)
	require.Nil(t, overrides.UsernameHeaders)
	resolved, err := appkitproxy.Resolve(appkitproxy.Options{Overrides: overrides})
	require.NoError(t, err)
	require.Equal(t, []string{"X-Verified-Email"}, resolved.EmailHeaders)
}
