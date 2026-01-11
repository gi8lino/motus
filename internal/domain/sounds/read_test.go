package sounds

import "testing"

import "github.com/stretchr/testify/assert"

func TestBuiltinOptions(t *testing.T) {
	t.Parallel()

	t.Run("Chime", func(t *testing.T) {
		t.Parallel()

		found := false
		for _, opt := range BuiltinOptions {
			if opt.Key == "chime" {
				found = true
				assert.Equal(t, "Gentle Chime", opt.Label)
				assert.NotEmpty(t, opt.File)
				break
			}
		}
		assert.True(t, found, "expected chime option present")
	})
}

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
