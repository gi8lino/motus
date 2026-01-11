package exercises

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gi8lino/motus/internal/db"
)

func TestCreate(t *testing.T) {
	t.Parallel()

	t.Run("Validation", func(t *testing.T) {
		mgr := NewManager(&fakeExercisesStore{})
		_, err := mgr.Create(context.Background(), " ", "Name", false)
		var domainErr *Error
		assert.ErrorAs(t, err, &domainErr)
		assert.Equal(t, KindValidation, domainErr.Kind)
	})

	t.Run("Forbidden", func(t *testing.T) {
		t.Parallel()

		store := &fakeExercisesStore{
			getUserFn: func(ctx context.Context, id string) (*db.User, error) {
				return &db.User{IsAdmin: false}, nil
			},
		}
		mgr := NewManager(store)
		_, err := mgr.Create(context.Background(), "user", "Burpee", true)
		var domainErr *Error
		assert.ErrorAs(t, err, &domainErr)
		assert.Equal(t, KindForbidden, domainErr.Kind)
	})

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		store := &fakeExercisesStore{
			getUserFn: func(ctx context.Context, id string) (*db.User, error) {
				return &db.User{IsAdmin: true}, nil
			},
			createFn: func(ctx context.Context, name, userID string, isCore bool) (*db.Exercise, error) {
				assert.Equal(t, "Clean", name)
				return &db.Exercise{Name: name}, nil
			},
		}
		mgr := NewManager(store)
		out, err := mgr.Create(context.Background(), "user", "Clean", false)
		assert.NoError(t, err)
		assert.Equal(t, "Clean", out.Name)
	})
}

func TestDelete(t *testing.T) {
	t.Parallel()

	t.Run("Validation", func(t *testing.T) {
		t.Parallel()
		mgr := NewManager(&fakeExercisesStore{})
		err := mgr.Delete(context.Background(), " ", "ex")
		var domainErr *Error
		assert.ErrorAs(t, err, &domainErr)
		assert.Equal(t, KindValidation, domainErr.Kind)
	})

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		store := &fakeExercisesStore{
			getUserFn: func(ctx context.Context, id string) (*db.User, error) {
				return &db.User{IsAdmin: true}, nil
			},
			getExFn: func(ctx context.Context, id string) (*db.Exercise, error) {
				return &db.Exercise{ID: id}, nil
			},
			deleteFn: func(ctx context.Context, id string) error {
				return nil
			},
		}
		mgr := NewManager(store)
		assert.NoError(t, mgr.Delete(context.Background(), "user", "ex"))
	})
}

func TestUpdate(t *testing.T) {
	t.Parallel()

	t.Run("Validation", func(t *testing.T) {
		t.Parallel()
		mgr := NewManager(&fakeExercisesStore{})
		_, err := mgr.Update(context.Background(), " ", "exo", "Name")
		var domainErr *Error
		assert.ErrorAs(t, err, &domainErr)
	})

	t.Run("NotFound", func(t *testing.T) {
		t.Parallel()
		store := &fakeExercisesStore{
			getUserFn: func(ctx context.Context, id string) (*db.User, error) {
				return &db.User{IsAdmin: true}, nil
			},
			getExFn: func(ctx context.Context, id string) (*db.Exercise, error) {
				return nil, nil
			},
		}
		mgr := NewManager(store)
		_, err := mgr.Update(context.Background(), "user", "ex", "Name")
		var domainErr *Error
		assert.ErrorAs(t, err, &domainErr)
		assert.Equal(t, KindNotFound, domainErr.Kind)
	})

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		store := &fakeExercisesStore{
			getUserFn: func(ctx context.Context, id string) (*db.User, error) {
				return &db.User{IsAdmin: true}, nil
			},
			getExFn: func(ctx context.Context, id string) (*db.Exercise, error) {
				return &db.Exercise{ID: id}, nil
			},
			renameFn: func(ctx context.Context, id, name string) (*db.Exercise, error) {
				return &db.Exercise{ID: id, Name: name}, nil
			},
		}
		mgr := NewManager(store)
		out, err := mgr.Update(context.Background(), "user", "ex", "New")
		assert.NoError(t, err)
		assert.Equal(t, "New", out.Name)
	})
}
