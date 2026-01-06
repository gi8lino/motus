package utils_test

import (
	"encoding/hex"
	"testing"

	"github.com/gi8lino/motus/internal/utils"
	"github.com/stretchr/testify/require"
)

func TestNewID(t *testing.T) {
	t.Parallel()

	t.Run("New ID", func(t *testing.T) {
		t.Parallel()
		id := utils.NewID()
		require.Len(t, id, 32)
		_, err := hex.DecodeString(id)
		require.NoError(t, err)
	})

	t.Run("Unique IDs", func(t *testing.T) {
		t.Parallel()
		seen := make(map[string]struct{})
		for range 100 {
			id := utils.NewID()
			_, exists := seen[id]
			require.False(t, exists)
			seen[id] = struct{}{}
		}
	})
}
