package templates

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domaintemplates "github.com/gi8lino/motus/internal/domain/templates"
	"github.com/gi8lino/motus/internal/service"
)

type fakeTemplateStore struct {
	listTemplatesFn           func(context.Context) ([]domaintemplates.Workout, error)
	createTemplateFn          func(context.Context, string, string) (*domaintemplates.Workout, error)
	workoutWithStepsFn        func(context.Context, string) (*domaintemplates.Workout, error)
	createWorkoutFromTemplate func(context.Context, string, string, string) (*domaintemplates.Workout, error)
}

func (f *fakeTemplateStore) ListTemplates(ctx context.Context) ([]domaintemplates.Workout, error) {
	if f.listTemplatesFn == nil {
		return nil, nil
	}
	return f.listTemplatesFn(ctx)
}

func (f *fakeTemplateStore) CreateTemplateFromWorkout(ctx context.Context, workoutID, name string) (*domaintemplates.Workout, error) {
	if f.createTemplateFn == nil {
		return nil, nil
	}
	return f.createTemplateFn(ctx, workoutID, name)
}

func (f *fakeTemplateStore) WorkoutWithSteps(ctx context.Context, id string) (*domaintemplates.Workout, error) {
	if f.workoutWithStepsFn == nil {
		return nil, nil
	}
	return f.workoutWithStepsFn(ctx, id)
}

func (f *fakeTemplateStore) CreateWorkoutFromTemplate(ctx context.Context, templateID, userID, name string) (*domaintemplates.Workout, error) {
	if f.createWorkoutFromTemplate == nil {
		return nil, nil
	}
	return f.createWorkoutFromTemplate(ctx, templateID, userID, name)
}

func TestCreate(t *testing.T) {
	t.Parallel()

	t.Run("Validation error", func(t *testing.T) {
		t.Parallel()

		svc := New(&fakeTemplateStore{})
		_, err := svc.Create(context.Background(), " ", "Name")
		require.Error(t, err)
		assert.True(t, service.IsKind(err, service.ErrorValidation))
	})
}
