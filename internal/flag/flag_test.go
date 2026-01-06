package flag

import (
	"os"
	"strings"
	"testing"

	"github.com/containeroo/tinyflags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gi8lino/motus/internal/logging"
)

const testDatabaseURL = "postgres://motus:motus@localhost:5432/motus?sslmode=disable"

func clearEnv(t *testing.T) {
	t.Helper()
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, "MOTUS_") {
			continue
		}
		parts := strings.SplitN(entry, "=", 2)
		key := parts[0]
		value, ok := os.LookupEnv(key)
		if ok {
			t.Cleanup(func() {
				_ = os.Setenv(key, value)
			})
		} else {
			t.Cleanup(func() {
				_ = os.Unsetenv(key)
			})
		}
		_ = os.Unsetenv(key)
	}
}

func TestParseFlags(t *testing.T) {
	// clearEnv(t) // since using t.Setenv Parallel will panic

	t.Run("use defaults", func(t *testing.T) {
		clearEnv(t)

		cfg, err := ParseFlags([]string{"--database-url", testDatabaseURL}, "vX.Y.Z")
		assert.NoError(t, err)
		assert.Equal(t, ":8080", cfg.ListenAddr, "default listen address")
		assert.Equal(t, "http://localhost:8080", cfg.SiteRoot, "default site root")
		assert.False(t, cfg.Debug, "default debug flag")
		assert.Equal(t, logging.LogFormat("json"), cfg.LogFormat, "default log format")
		assert.Equal(t, "", cfg.RoutePrefix, "default route prefix")
		assert.Equal(t, "", cfg.AuthHeader, "default auth header")
		assert.False(t, cfg.AllowRegistration, "default allow registration")
		assert.False(t, cfg.AutoCreateUsers, "default auto-create users")
		assert.Equal(t, testDatabaseURL, cfg.DatabaseURL, "database url")
		assert.Equal(t, "", cfg.AdminEmail, "default admin email")
		assert.Equal(t, "", cfg.AdminPassword, "default admin password")
	})

	t.Run("show version", func(t *testing.T) {
		clearEnv(t)

		_, err := ParseFlags([]string{"--version"}, "1.2.3")
		assert.Error(t, err)
		assert.True(t, tinyflags.IsVersionRequested(err))
		assert.EqualError(t, err, "1.2.3")
	})

	t.Run("show help", func(t *testing.T) {
		clearEnv(t)

		_, err := ParseFlags([]string{"--help"}, "")
		assert.Error(t, err)
		assert.True(t, tinyflags.IsHelpRequested(err))
		require.Error(t, err)
		usage := err.Error()
		assert.True(t, strings.HasPrefix(usage, "Usage: motus [flags]"))
		assert.Contains(t, usage, "--database-url")
		assert.Contains(t, usage, "--route-prefix")
		assert.Contains(t, usage, "--auth-header")
		assert.Contains(t, usage, "--allow-registration")
		assert.Contains(t, usage, "--auto-create-users")
		assert.Contains(t, usage, "--site-root")
		assert.Contains(t, usage, "--admin-email")
		assert.Contains(t, usage, "--admin-password")
		assert.Contains(t, usage, "--log-format")
	})

	t.Run("custom values", func(t *testing.T) {
		clearEnv(t)

		args := []string{
			"--database-url", testDatabaseURL,
			"-a", "127.0.0.1:9000",
			"--route-prefix", "motus",
			"--auth-header", "X-User-Email",
			"--allow-registration",
			"--auto-create-users",
			"-r", "https://example.com",
			"--admin-email", "admin@example.com",
			"--admin-password", "secret",
			"-d",
			"-l", "text",
		}
		cfg, err := ParseFlags(args, "0.0.0")
		assert.NoError(t, err)
		assert.Equal(t, "127.0.0.1:9000", cfg.ListenAddr)
		assert.Equal(t, "/motus", cfg.RoutePrefix)
		assert.Equal(t, "https://example.com/motus", cfg.SiteRoot)
		assert.Equal(t, "X-User-Email", cfg.AuthHeader)
		assert.True(t, cfg.AllowRegistration)
		assert.True(t, cfg.AutoCreateUsers)
		assert.Equal(t, "admin@example.com", cfg.AdminEmail)
		assert.Equal(t, "secret", cfg.AdminPassword)
		assert.True(t, cfg.Debug)
		assert.Equal(t, logging.LogFormat("text"), cfg.LogFormat)
	})

	t.Run("parsing error", func(t *testing.T) {
		clearEnv(t)
		args := []string{"--database-url", testDatabaseURL, "--invalid"}
		_, err := ParseFlags(args, "")

		assert.Error(t, err)
		assert.EqualError(t, err, "unknown flag: --invalid")
	})

	t.Run("missing database url", func(t *testing.T) {
		clearEnv(t)

		args := []string{}
		_, err := ParseFlags(args, "")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database-url")
	})

	t.Run("valid log format", func(t *testing.T) {
		clearEnv(t)
		args := []string{"--database-url", testDatabaseURL, "--log-format", "json"}
		_, err := ParseFlags(args, "")

		assert.NoError(t, err)
	})

	t.Run("invalid log format", func(t *testing.T) {
		clearEnv(t)
		args := []string{"--database-url", testDatabaseURL, "--log-format", "xml"}
		_, err := ParseFlags(args, "")

		assert.Error(t, err)
		assert.EqualError(t, err, "invalid value for flag --log-format: must be one of: text, json.")
	})

	t.Run("test route prefix", func(t *testing.T) {
		clearEnv(t)
		args := []string{"--database-url", testDatabaseURL, "--route-prefix", "motus"}
		flags, err := ParseFlags(args, "")

		require.NoError(t, err)
		assert.Equal(t, "/motus", flags.RoutePrefix)
		assert.Equal(t, "http://localhost:8080/motus", flags.SiteRoot)
	})
}
