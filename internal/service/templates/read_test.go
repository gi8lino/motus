package templates

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domaintemplates "github.com/gi8lino/motus/internal/domain/templates"
	"github.com/gi8lino/motus/internal/service"
)

func TestList(t *testing.T) {
	t.Parallel()

	t.Run("Internal error", func(t *testing.T) {
		t.Parallel()

		svc := New(&fakeTemplateStore{
			listTemplatesFn: func(context.Context) ([]domaintemplates.Workout, error) {
				return nil, errors.New("boom")
			},
		})
		_, err := svc.List(context.Background())
		require.Error(t, err)
		assert.True(t, service.IsKind(err, service.ErrorInternal))
	})
}

func TestGet(t *testing.T) {
	t.Parallel()

	t.Run("Not a template", func(t *testing.T) {
		t.Parallel()

		svc := New(&fakeTemplateStore{
			workoutWithStepsFn: func(context.Context, string) (*domaintemplates.Workout, error) {
				return &domaintemplates.Workout{ID: "w1", IsTemplate: false}, nil
			},
		})
		_, err := svc.Get(context.Background(), "w1")
		require.Error(t, err)
		assert.True(t, service.IsKind(err, service.ErrorNotFound))
	})
}
