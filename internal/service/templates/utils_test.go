package templates

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequireID(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		val, err := requireID("  TMP  ", "missing")
		assert.NoError(t, err)
		assert.Equal(t, "tmp", val)
	})

	t.Run("Missing", func(t *testing.T) {
		t.Parallel()
		_, err := requireID(" ", "missing")
		assert.Error(t, err)
	})
}
