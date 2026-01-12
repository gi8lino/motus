package templates

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gi8lino/motus/internal/db"
)

type readOnlyStore struct {
	listFn func(context.Context) ([]db.Workout, error)
	getFn  func(context.Context, string) (*db.Workout, error)
}

func (s *readOnlyStore) ListTemplates(ctx context.Context) ([]db.Workout, error) {
	if s.listFn == nil {
		return nil, nil
	}
	return s.listFn(ctx)
}

func (s *readOnlyStore) CreateTemplateFromWorkout(ctx context.Context, workoutID, name string) (*db.Workout, error) {
	return nil, nil
}

func (s *readOnlyStore) WorkoutWithSteps(ctx context.Context, id string) (*db.Workout, error) {
	if s.getFn == nil {
		return nil, nil
	}
	return s.getFn(ctx, id)
}

func (s *readOnlyStore) CreateWorkoutFromTemplate(ctx context.Context, templateID, userID, name string) (*db.Workout, error) {
	return nil, nil
}

func TestList(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		store := &readOnlyStore{
			listFn: func(ctx context.Context) ([]db.Workout, error) {
				return []db.Workout{{ID: "t1"}}, nil
			},
		}
		mgr := NewManager(store)
		items, err := mgr.List(context.Background())
		assert.NoError(t, err)
		assert.Len(t, items, 1)
	})
}

func TestGet(t *testing.T) {
	t.Parallel()

	t.Run("Validation", func(t *testing.T) {
		t.Parallel()
		mgr := NewManager(&readOnlyStore{})
		_, err := mgr.Get(context.Background(), " ")
		assert.Error(t, err)
		var domainErr *Error
		assert.ErrorAs(t, err, &domainErr)
		assert.Equal(t, KindValidation, domainErr.Kind)
	})

	t.Run("NotTemplate", func(t *testing.T) {
		t.Parallel()
		store := &readOnlyStore{
			getFn: func(ctx context.Context, id string) (*db.Workout, error) {
				return &db.Workout{IsTemplate: false}, nil
			},
		}
		mgr := NewManager(store)
		_, err := mgr.Get(context.Background(), "tmpl")
		assert.Error(t, err)
		var domainErr *Error
		assert.ErrorAs(t, err, &domainErr)
		assert.Equal(t, KindNotFound, domainErr.Kind)
	})
}
