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
