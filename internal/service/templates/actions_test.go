package templates

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domaintemplates "github.com/gi8lino/motus/internal/domain/templates"
	"github.com/gi8lino/motus/internal/service"
)

func TestApply(t *testing.T) {
	t.Parallel()

	t.Run("Validation error", func(t *testing.T) {
		t.Parallel()

		svc := New(&fakeTemplateStore{})
		_, err := svc.Apply(context.Background(), " ", "user", "Name")
		require.Error(t, err)
		assert.True(t, service.IsKind(err, service.ErrorValidation))
	})

	t.Run("Creates new workout", func(t *testing.T) {
		t.Parallel()

		svc := New(&fakeTemplateStore{
			createWorkoutFromTemplate: func(context.Context, string, string, string) (*domaintemplates.Workout, error) {
				return &domaintemplates.Workout{ID: "new", Name: "Copy"}, nil
			},
		})
		workout, err := svc.Apply(context.Background(), "tmpl", "user", "Copy")
		require.NoError(t, err)
		assert.Equal(t, "new", workout.ID)
	})
}
