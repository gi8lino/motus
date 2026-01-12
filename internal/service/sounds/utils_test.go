package sounds

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidKey(t *testing.T) {
	t.Parallel()

	t.Run("Empty key is valid", func(t *testing.T) {
		t.Parallel()
		assert.True(t, ValidKey(""))
	})

	t.Run("Known key is valid", func(t *testing.T) {
		t.Parallel()
		assert.True(t, ValidKey("beep"))
	})

	t.Run("Unknown key is invalid", func(t *testing.T) {
		t.Parallel()
		assert.False(t, ValidKey("nope"))
	})
}

func TestURLByKey(t *testing.T) {
	t.Parallel()

	t.Run("Known key returns URL", func(t *testing.T) {
		t.Parallel()
		assert.NotEmpty(t, URLByKey("beep"))
	})

	t.Run("Unknown key returns empty", func(t *testing.T) {
		t.Parallel()
		assert.Empty(t, URLByKey("nope"))
	})
}
