package utils_test

import (
	"testing"

	"github.com/gi8lino/motus/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeEmail(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		mail, err := utils.NormalizeEmail("test@example.com")
		require.NoError(t, err)
		assert.Equal(t, "test@example.com", mail)
	})

	t.Run("Empty address", func(t *testing.T) {
		t.Parallel()
		_, err := utils.NormalizeEmail("")
		require.Error(t, err)
	})

	t.Run("Email parse error", func(t *testing.T) {
		t.Parallel()
		_, err := utils.NormalizeEmail("not-an-email")
		require.Error(t, err)
	})
}
