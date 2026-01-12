package sounds

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
