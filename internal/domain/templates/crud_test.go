package templates

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gi8lino/motus/internal/db"
)

type createOnlyStore struct {
	createFn func(context.Context, string, string) (*db.Workout, error)
}

func (s *createOnlyStore) ListTemplates(ctx context.Context) ([]db.Workout, error) {
	return nil, nil
}

func (s *createOnlyStore) CreateTemplateFromWorkout(ctx context.Context, workoutID, name string) (*db.Workout, error) {
	return s.createFn(ctx, workoutID, name)
}

func (s *createOnlyStore) WorkoutWithSteps(ctx context.Context, id string) (*db.Workout, error) {
	return nil, nil
}

func (s *createOnlyStore) CreateWorkoutFromTemplate(ctx context.Context, templateID, userID, name string) (*db.Workout, error) {
	return nil, nil
}

func TestCreate(t *testing.T) {
	t.Parallel()

	t.Run("Validation", func(t *testing.T) {
		t.Parallel()
		mgr := NewManager(&createOnlyStore{})
		_, err := mgr.Create(context.Background(), " ", "Name")
		assert.Error(t, err)
		var domainErr *Error
		assert.ErrorAs(t, err, &domainErr)
		assert.Equal(t, KindValidation, domainErr.Kind)
	})

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		store := &createOnlyStore{
			createFn: func(ctx context.Context, workoutID, name string) (*db.Workout, error) {
				return &db.Workout{ID: workoutID, Name: name}, nil
			},
		}
		mgr := NewManager(store)
		out, err := mgr.Create(context.Background(), "w1", "Template")
		assert.NoError(t, err)
		assert.Equal(t, "Template", out.Name)
	})
}
