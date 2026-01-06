package app

import (
	"bytes"
	"context"
	"embed"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func TestRun(t *testing.T) {
	clearEnv(t)
	t.Run("Help returns usage", func(t *testing.T) {
		var out bytes.Buffer
		err := Run(context.Background(), embed.FS{}, "v1.2.3", "abc", []string{"--help"}, &out)
		require.NoError(t, err)
		assert.Contains(t, out.String(), "Usage: motus")
	})

	t.Run("Version returns output", func(t *testing.T) {
		var out bytes.Buffer
		err := Run(context.Background(), embed.FS{}, "v9.9.9", "abc", []string{"--version"}, &out)
		require.NoError(t, err)
		assert.Contains(t, out.String(), "v9.9.9")
	})

	t.Run("Missing database url fails", func(t *testing.T) {
		var out bytes.Buffer
		err := Run(context.Background(), embed.FS{}, "v1", "abc", []string{}, &out)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "CLI flags error")
	})

	t.Run("Invalid database url fails", func(t *testing.T) {
		var out bytes.Buffer
		err := Run(context.Background(), embed.FS{}, "v1", "abc", []string{"--database-url", "not-a-url"}, &out)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "connect db")
	})
}
