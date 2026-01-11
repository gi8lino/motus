package exercises

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBackfill(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		called := false
		store := &fakeExercisesStore{
			backfillFn: func(ctx context.Context) error {
				called = true
				return nil
			},
		}
		assert.NoError(t, NewManager(store).Backfill(context.Background()))
		assert.True(t, called)
	})
}
