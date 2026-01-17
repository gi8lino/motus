package sounds

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestURLByKey(t *testing.T) {
	t.Parallel()

	t.Run("Normalizes", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "/sounds/chime.wav", URLByKey("ChImE"))
		assert.Equal(t, "", URLByKey("missing"))
	})
}

func TestValidKey(t *testing.T) {
	t.Parallel()

	t.Run("Checks", func(t *testing.T) {
		t.Parallel()
		assert.True(t, ValidKey("BEEP"))
		assert.True(t, ValidKey(""))
		assert.False(t, ValidKey("unknown"))
	})
}
