package templates

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gi8lino/motus/internal/db"
)

type applyOnlyStore struct {
	applyFn func(context.Context, string, string, string) (*db.Workout, error)
}

func (s *applyOnlyStore) ListTemplates(ctx context.Context) ([]db.Workout, error) {
	return nil, nil
}

func (s *applyOnlyStore) CreateTemplateFromWorkout(ctx context.Context, workoutID, name string) (*db.Workout, error) {
	return nil, nil
}

func (s *applyOnlyStore) WorkoutWithSteps(ctx context.Context, id string) (*db.Workout, error) {
	return nil, nil
}

func (s *applyOnlyStore) CreateWorkoutFromTemplate(ctx context.Context, templateID, userID, name string) (*db.Workout, error) {
	return s.applyFn(ctx, templateID, userID, name)
}

func TestApply(t *testing.T) {
	t.Parallel()

	t.Run("Validation", func(t *testing.T) {
		t.Parallel()
		mgr := NewManager(&applyOnlyStore{})
		_, err := mgr.Apply(context.Background(), " ", "user", "Name")
		assert.Error(t, err)
		var domainErr *Error
		assert.ErrorAs(t, err, &domainErr)
		assert.Equal(t, KindValidation, domainErr.Kind)
	})

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		store := &applyOnlyStore{
			applyFn: func(ctx context.Context, templateID, userID, name string) (*db.Workout, error) {
				return &db.Workout{ID: templateID, Name: name}, nil
			},
		}
		mgr := NewManager(store)
		out, err := mgr.Apply(context.Background(), "tmpl", "user", "Copy")
		assert.NoError(t, err)
		assert.Equal(t, "Copy", out.Name)
	})
}
