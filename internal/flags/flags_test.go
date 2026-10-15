package flags

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/containeroo/tinyflags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	appkitlogging "github.com/containeroo/appkit/logging"
)

const testDatabaseURL = "postgres://motus:motus@localhost:5432/motus?sslmode=disable"

func clearEnv(t *testing.T) {
	t.Helper()
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, "MOTUS_") {
			continue
		}
		key := strings.SplitN(entry, "=", 2)[0]
		value, existed := os.LookupEnv(key)
		if existed {
			t.Cleanup(func() { _ = os.Setenv(key, value) })
		} else {
			t.Cleanup(func() { _ = os.Unsetenv(key) })
		}
		_ = os.Unsetenv(key)
	}
}

func TestParse(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		clearEnv(t)

		cfg, err := Parse([]string{"--database-url", testDatabaseURL}, "vX.Y.Z")
		require.NoError(t, err)
		assert.Equal(t, "127.0.0.1:8080", cfg.Server.ListenAddress)
		assert.Equal(t, "127.0.0.1:8081", cfg.Server.ManagementListenAddress)
		assert.Equal(t, "http://localhost:8080", cfg.Server.PublicURL)
		assert.Empty(t, cfg.Server.RoutePrefix)
		assert.Equal(t, appkitlogging.LogFormatJSON, cfg.Logging.Format)
		assert.False(t, cfg.Logging.Debug)
		assert.False(t, cfg.Logging.AccessLog)
		assert.Equal(t, testDatabaseURL, cfg.Database.URL)
		assert.Nil(t, cfg.Authentication.AllowRegistrationOverride)
		assert.Empty(t, cfg.Authentication.ModeOverride)
	})

	t.Run("version", func(t *testing.T) {
		clearEnv(t)
		_, err := Parse([]string{"--version"}, "1.2.3")
		require.Error(t, err)
		assert.True(t, tinyflags.IsVersionRequested(err))
		assert.EqualError(t, err, "1.2.3")
	})

	t.Run("help is sectioned", func(t *testing.T) {
		clearEnv(t)
		_, err := Parse([]string{"--help"}, "")
		require.Error(t, err)
		assert.True(t, tinyflags.IsHelpRequested(err))
		usage := err.Error()
		assert.Contains(t, usage, "Usage: motus [flags]")
		assert.Contains(t, usage, "Server:")
		assert.Contains(t, usage, "Database:")
		assert.Contains(t, usage, "Authentication:")
		assert.Contains(t, usage, "OIDC:")
		assert.Contains(t, usage, "Logging and diagnostics:")
		assert.Contains(t, usage, "--management-listen-address")
		assert.Contains(t, usage, "--public-url")
		assert.Contains(t, usage, "--access-log")
	})

	t.Run("invalid route prefix", func(t *testing.T) {
		clearEnv(t)
		_, err := Parse([]string{"--database-url", testDatabaseURL, "--route-prefix", "motus"}, "")
		require.Error(t, err)
	})

	t.Run("OIDC settings belong to the database", func(t *testing.T) {
		clearEnv(t)
		_, err := Parse([]string{"--database-url", testDatabaseURL, "--auth-mode", "oidc"}, "")
		require.NoError(t, err)
	})

	t.Run("OIDC secrets remain deployment-managed", func(t *testing.T) {
		clearEnv(t)
		_, err := Parse([]string{
			"--database-url", testDatabaseURL,
			"--oidc-client-secret", "secret",
			"--oidc-session-secret", strings.Repeat("x", 32),
		}, "")
		require.NoError(t, err)
	})

	t.Run("missing database URL", func(t *testing.T) {
		clearEnv(t)
		_, err := Parse(nil, "")
		require.ErrorContains(t, err, "database-url")
	})

	t.Run("invalid log format", func(t *testing.T) {
		clearEnv(t)
		_, err := Parse([]string{"--database-url", testDatabaseURL, "--log-format", "xml"}, "")
		require.EqualError(t, err, "invalid value for flag --log-format: \"xml\" must be one of: text, json")
	})
}

func TestProxyTrustAndSessionFlags(t *testing.T) {
	clearEnv(t)
	cfg, err := Parse([]string{
		"--database-url", testDatabaseURL,
		"--trusted-proxy-cidrs", "127.0.0.1/32",
		"--session-lifetime", "48h",
		"--auth-mode", "proxy",
	}, "")
	require.NoError(t, err)
	require.Equal(t, []string{"127.0.0.1/32"}, cfg.Proxy.TrustedCIDRs)
	require.Equal(t, 48*time.Hour, cfg.Authentication.SessionLifetime)
	require.Equal(t, "proxy", string(*cfg.Authentication.ModeOverride))
}

func TestProxyWildcardFlag(t *testing.T) {
	clearEnv(t)
	cfg, err := Parse([]string{
		"--database-url", testDatabaseURL,
		"--auth-mode", "proxy",
		"--trusted-proxy-cidrs", "*",
	}, "")
	require.NoError(t, err)
	assert.Equal(t, []string{"*"}, cfg.Proxy.TrustedCIDRs)
}

// TestProxyEnvironmentOverrides verifies Envoy identity headers retain their explicit override origins.
func TestProxyEnvironmentOverrides(t *testing.T) {
	clearEnv(t)
	t.Setenv("MOTUS__AUTH_MODE", "proxy")
	t.Setenv("MOTUS__TRUSTED_PROXY_CIDRS", "*")
	t.Setenv("MOTUS__TRUSTED_USERNAME_HEADERS", "X-Auth-Request-Email")
	t.Setenv("MOTUS__TRUSTED_EMAIL_HEADERS", "X-Auth-Request-Email")

	cfg, err := Parse([]string{"--database-url", testDatabaseURL}, "")
	require.NoError(t, err)
	require.NotNil(t, cfg.Authentication.ModeOverride)
	assert.Equal(t, "proxy", string(*cfg.Authentication.ModeOverride))
	assert.Equal(t, []string{"*"}, cfg.Proxy.TrustedCIDRs)
	assert.Equal(t, []string{"X-Auth-Request-Email"}, cfg.Proxy.UsernameHeaders)
	assert.Equal(t, []string{"X-Auth-Request-Email"}, cfg.Proxy.EmailHeaders)

	values := cfg.Overrides.Values()
	_, usernameOverridden := values["trusted-username-headers"]
	_, emailOverridden := values["trusted-email-headers"]
	assert.True(t, usernameOverridden)
	assert.True(t, emailOverridden)
}
