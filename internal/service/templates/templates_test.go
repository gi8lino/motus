package templates

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gi8lino/motus/internal/db"
	"github.com/gi8lino/motus/internal/service"
)

type fakeTemplateStore struct {
	listTemplatesFn           func(context.Context) ([]db.Workout, error)
	createTemplateFn          func(context.Context, string, string) (*db.Workout, error)
	workoutWithStepsFn        func(context.Context, string) (*db.Workout, error)
	createWorkoutFromTemplate func(context.Context, string, string, string) (*db.Workout, error)
}

func (f *fakeTemplateStore) ListTemplates(ctx context.Context) ([]db.Workout, error) {
	if f.listTemplatesFn == nil {
		return nil, nil
	}
	return f.listTemplatesFn(ctx)
}

func (f *fakeTemplateStore) CreateTemplateFromWorkout(ctx context.Context, workoutID, name string) (*db.Workout, error) {
	if f.createTemplateFn == nil {
		return nil, nil
	}
	return f.createTemplateFn(ctx, workoutID, name)
}

func (f *fakeTemplateStore) WorkoutWithSteps(ctx context.Context, id string) (*db.Workout, error) {
	if f.workoutWithStepsFn == nil {
		return nil, nil
	}
	return f.workoutWithStepsFn(ctx, id)
}

func (f *fakeTemplateStore) CreateWorkoutFromTemplate(ctx context.Context, templateID, userID, name string) (*db.Workout, error) {
	if f.createWorkoutFromTemplate == nil {
		return nil, nil
	}
	return f.createWorkoutFromTemplate(ctx, templateID, userID, name)
}

func TestServiceList(t *testing.T) {
	t.Parallel()

	t.Run("Internal error", func(t *testing.T) {
		t.Parallel()

		svc := &Service{Store: &fakeTemplateStore{
			listTemplatesFn: func(context.Context) ([]db.Workout, error) {
				return nil, errors.New("boom")
			},
		}}
		_, err := svc.List(context.Background())
		require.Error(t, err)
		assert.True(t, service.IsKind(err, service.ErrorInternal))
	})
}

func TestServiceCreate(t *testing.T) {
	t.Parallel()

	t.Run("Validation error", func(t *testing.T) {
		t.Parallel()

		svc := &Service{Store: &fakeTemplateStore{}}
		_, err := svc.Create(context.Background(), " ", "Name")
		require.Error(t, err)
		assert.True(t, service.IsKind(err, service.ErrorValidation))
	})
}

func TestServiceGet(t *testing.T) {
	t.Parallel()

	t.Run("Not a template", func(t *testing.T) {
		t.Parallel()

		svc := &Service{Store: &fakeTemplateStore{
			workoutWithStepsFn: func(context.Context, string) (*db.Workout, error) {
				return &db.Workout{ID: "w1", IsTemplate: false}, nil
			},
		}}
		_, err := svc.Get(context.Background(), "w1")
		require.Error(t, err)
		assert.True(t, service.IsKind(err, service.ErrorNotFound))
	})
}

func TestServiceApply(t *testing.T) {
	t.Parallel()

	t.Run("Validation error", func(t *testing.T) {
		t.Parallel()

		svc := &Service{Store: &fakeTemplateStore{}}
		_, err := svc.Apply(context.Background(), " ", "user", "Name")
		require.Error(t, err)
		assert.True(t, service.IsKind(err, service.ErrorValidation))
	})

	t.Run("Creates new workout", func(t *testing.T) {
		t.Parallel()

		svc := &Service{Store: &fakeTemplateStore{
			createWorkoutFromTemplate: func(context.Context, string, string, string) (*db.Workout, error) {
				return &db.Workout{ID: "new", Name: "Copy"}, nil
			},
		}}
		workout, err := svc.Apply(context.Background(), "tmpl", "user", "Copy")
		require.NoError(t, err)
		assert.Equal(t, "new", workout.ID)
	})
}
