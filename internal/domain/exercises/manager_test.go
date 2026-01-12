package exercises

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gi8lino/motus/internal/db"
)

type fakeExercisesStore struct {
	listFn     func(context.Context, string) ([]db.Exercise, error)
	getUserFn  func(context.Context, string) (*db.User, error)
	createFn   func(context.Context, string, string, bool) (*db.Exercise, error)
	getExFn    func(context.Context, string) (*db.Exercise, error)
	renameFn   func(context.Context, string, string) (*db.Exercise, error)
	deleteFn   func(context.Context, string) error
	backfillFn func(context.Context) error
}

func (f *fakeExercisesStore) ListExercises(ctx context.Context, userID string) ([]db.Exercise, error) {
	if f.listFn == nil {
		return nil, nil
	}
	return f.listFn(ctx, userID)
}

func (f *fakeExercisesStore) GetUser(ctx context.Context, userID string) (*db.User, error) {
	if f.getUserFn == nil {
		return nil, nil
	}
	return f.getUserFn(ctx, userID)
}

func (f *fakeExercisesStore) CreateExercise(ctx context.Context, name, userID string, isCore bool) (*db.Exercise, error) {
	if f.createFn == nil {
		return nil, nil
	}
	return f.createFn(ctx, name, userID, isCore)
}

func (f *fakeExercisesStore) GetExercise(ctx context.Context, id string) (*db.Exercise, error) {
	if f.getExFn == nil {
		return nil, nil
	}
	return f.getExFn(ctx, id)
}

func (f *fakeExercisesStore) RenameExercise(ctx context.Context, id, name string) (*db.Exercise, error) {
	if f.renameFn == nil {
		return nil, nil
	}
	return f.renameFn(ctx, id, name)
}

func (f *fakeExercisesStore) DeleteExercise(ctx context.Context, id string) error {
	if f.deleteFn == nil {
		return nil
	}
	return f.deleteFn(ctx, id)
}

func (f *fakeExercisesStore) BackfillCoreExercises(ctx context.Context) error {
	if f.backfillFn == nil {
		return nil
	}
	return f.backfillFn(ctx)
}

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

func TestBackfill(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		called := false
		store := &fakeExercisesStore{
			backfillFn: func(ctx context.Context) error {
				called = true
				return nil
			},
		}
		assert.NoError(t, NewManager(store).Backfill(context.Background()))
		assert.True(t, called)
	})
}
