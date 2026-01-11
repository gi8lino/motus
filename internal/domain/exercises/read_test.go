package exercises

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gi8lino/motus/internal/db"
)

func TestList(t *testing.T) {
	t.Parallel()

	t.Run("Validation", func(t *testing.T) {
		t.Parallel()
		mgr := NewManager(&fakeExercisesStore{})
		_, err := mgr.List(context.Background(), " ")
		var domainErr *Error
		assert.ErrorAs(t, err, &domainErr)
		assert.Equal(t, KindValidation, domainErr.Kind)
	})

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		called := false
		store := &fakeExercisesStore{
			listFn: func(ctx context.Context, userID string) ([]db.Exercise, error) {
				called = true
				assert.Equal(t, "user", userID)
				return []db.Exercise{{ID: "ex"}}, nil
			},
		}
		mgr := NewManager(store)
		out, err := mgr.List(context.Background(), "user")
		assert.NoError(t, err)
		assert.True(t, called)
		assert.Len(t, out, 1)
	})
}
