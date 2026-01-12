package exercises

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequireUserID(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		val, err := requireUserID(" USER ")
		assert.NoError(t, err)
		assert.Equal(t, "user", val)
	})

	t.Run("Missing", func(t *testing.T) {
		t.Parallel()
		_, err := requireUserID(" ")
		assert.Error(t, err)
	})
}

func TestRequireName(t *testing.T) {
	t.Parallel()

	t.Run("Missing", func(t *testing.T) {
		t.Parallel()
		_, err := requireName(" ")
		assert.Error(t, err)
	})
}

func TestRequireEntityID(t *testing.T) {
	t.Parallel()

	t.Run("Missing", func(t *testing.T) {
		t.Parallel()
		_, err := requireEntityID(" ", "msg")
		assert.Error(t, err)
	})
}
