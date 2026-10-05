package flags

import (
	"os"
	"strings"
	"testing"

	"github.com/containeroo/tinyflags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	httpauth "github.com/gi8lino/motus/internal/http/auth"
	"github.com/gi8lino/motus/internal/logging"
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
		assert.Equal(t, "127.0.0.1:8080", cfg.ApplicationListenAddress)
		assert.Equal(t, "127.0.0.1:8081", cfg.ManagementListenAddress)
		assert.Equal(t, "http://localhost:8080", cfg.PublicURL)
		assert.Empty(t, cfg.RoutePrefix)
		assert.Equal(t, logging.LogFormatJSON, cfg.LogFormat)
		assert.False(t, cfg.Debug)
		assert.False(t, cfg.AccessLog)
		assert.Equal(t, testDatabaseURL, cfg.DatabaseURL)
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

	t.Run("custom values", func(t *testing.T) {
		clearEnv(t)
		cfg, err := Parse([]string{
			"--database-url", testDatabaseURL,
			"--listen-address", "127.0.0.1:9000",
			"--management-listen-address", "127.0.0.1:9001",
			"--route-prefix", "/motus/",
			"--public-url", "https://example.com",
			"--auth-mode", "trusted-proxy",
			"--trusted-proxy-user-header", "X-User-Email",
			"--allow-registration",
			"--debug",
			"--access-log",
			"--log-format", "text",
		}, "0.0.0")
		require.NoError(t, err)
		assert.Equal(t, "127.0.0.1:9000", cfg.ApplicationListenAddress)
		assert.Equal(t, "127.0.0.1:9001", cfg.ManagementListenAddress)
		assert.Equal(t, "/motus", cfg.RoutePrefix)
		assert.Equal(t, "https://example.com/motus", cfg.PublicURL)
		assert.Equal(t, httpauth.ModeTrustedProxy, cfg.AuthMode)
		assert.Equal(t, "X-User-Email", cfg.TrustedProxyUserHeader)
		assert.True(t, cfg.AllowRegistration)
		assert.True(t, cfg.Debug)
		assert.True(t, cfg.AccessLog)
		assert.Equal(t, logging.LogFormatText, cfg.LogFormat)
		assert.NotEmpty(t, cfg.Overrides)
	})

	t.Run("invalid route prefix", func(t *testing.T) {
		clearEnv(t)
		_, err := Parse([]string{"--database-url", testDatabaseURL, "--route-prefix", "motus"}, "")
		require.Error(t, err)
	})

	t.Run("OIDC requires complete configuration", func(t *testing.T) {
		clearEnv(t)
		_, err := Parse([]string{"--database-url", testDatabaseURL, "--auth-mode", "oidc", "--oidc-issuer", "https://issuer.example", "--oidc-client-id", "motus"}, "")
		require.NoError(t, err)
	})

	t.Run("OIDC conflicts with trusted header auth", func(t *testing.T) {
		clearEnv(t)
		_, err := Parse([]string{
			"--database-url", testDatabaseURL,
			"--oidc-issuer", "https://issuer.example",
			"--oidc-client-id", "motus",
			"--oidc-client-secret", "secret",
			"--oidc-session-secret", strings.Repeat("x", 32),
			"--trusted-proxy-user-header", "X-User",
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
