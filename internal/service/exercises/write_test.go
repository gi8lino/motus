package exercises

import (
	"context"
	"testing"

	"github.com/gi8lino/motus/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeStore struct {
	listExercisesFn     func(context.Context, string) ([]Exercise, error)
	getUserFn           func(context.Context, string) (*User, error)
	createExerciseFn    func(context.Context, string, string, bool) (*Exercise, error)
	getExerciseFn       func(context.Context, string) (*Exercise, error)
	replaceExerciseFn   func(context.Context, string, string, string, string) error
	renameExerciseFn    func(context.Context, string, string) (*Exercise, error)
	deleteExerciseFn    func(context.Context, string) error
	backfillExercisesFn func(context.Context) error
}

func (f *fakeStore) ListExercises(ctx context.Context, userID string) ([]Exercise, error) {
	if f.listExercisesFn == nil {
		return nil, nil
	}
	return f.listExercisesFn(ctx, userID)
}

func (f *fakeStore) GetUser(ctx context.Context, userID string) (*User, error) {
	if f.getUserFn == nil {
		return nil, nil
	}
	return f.getUserFn(ctx, userID)
}

func (f *fakeStore) CreateExercise(ctx context.Context, name, userID string, isCore bool) (*Exercise, error) {
	if f.createExerciseFn == nil {
		return nil, nil
	}
	return f.createExerciseFn(ctx, name, userID, isCore)
}

func (f *fakeStore) GetExercise(ctx context.Context, id string) (*Exercise, error) {
	if f.getExerciseFn == nil {
		return nil, nil
	}
	return f.getExerciseFn(ctx, id)
}

func (f *fakeStore) ReplaceExerciseForUser(ctx context.Context, userID, oldID, newID, newName string) error {
	if f.replaceExerciseFn == nil {
		return nil
	}
	return f.replaceExerciseFn(ctx, userID, oldID, newID, newName)
}

func (f *fakeStore) RenameExercise(ctx context.Context, id, name string) (*Exercise, error) {
	if f.renameExerciseFn == nil {
		return nil, nil
	}
	return f.renameExerciseFn(ctx, id, name)
}

func (f *fakeStore) DeleteExercise(ctx context.Context, id string) error {
	if f.deleteExerciseFn == nil {
		return nil
	}
	return f.deleteExerciseFn(ctx, id)
}

func (f *fakeStore) BackfillCoreExercises(ctx context.Context) error {
	if f.backfillExercisesFn == nil {
		return nil
	}
	return f.backfillExercisesFn(ctx)
}

func TestCreate(t *testing.T) {
	t.Parallel()

	t.Run("User not found", func(t *testing.T) {
		t.Parallel()

		svc := New(&fakeStore{
			getUserFn: func(context.Context, string) (*User, error) {
				return nil, nil
			},
		})
		_, err := svc.Create(context.Background(), "user", "Burpee", false)
		require.Error(t, err)
		assert.True(t, service.IsKind(err, service.ErrorNotFound))
	})

	t.Run("Core requires admin", func(t *testing.T) {
		t.Parallel()

		called := false
		svc := New(&fakeStore{
			getUserFn: func(context.Context, string) (*User, error) {
				return &User{ID: "user", IsAdmin: false}, nil
			},
			createExerciseFn: func(context.Context, string, string, bool) (*Exercise, error) {
				called = true
				return nil, nil
			},
		})
		_, err := svc.Create(context.Background(), "user", "Burpee", true)
		require.Error(t, err)
		assert.True(t, service.IsKind(err, service.ErrorForbidden))
		assert.False(t, called, "expected CreateExercise not to be called")
	})
}

func TestUpdate(t *testing.T) {
	t.Parallel()

	t.Run("Requires admin", func(t *testing.T) {
		t.Parallel()

		svc := New(&fakeStore{
			getUserFn: func(context.Context, string) (*User, error) {
				return &User{ID: "user", IsAdmin: false}, nil
			},
			getExerciseFn: func(context.Context, string) (*Exercise, error) {
				return &Exercise{ID: "core", IsCore: true}, nil
			},
		})
		_, err := svc.Update(context.Background(), "user", "core", "Burpee")
		require.Error(t, err)
		assert.True(t, service.IsKind(err, service.ErrorForbidden))
	})

	t.Run("Admin can rename", func(t *testing.T) {
		t.Parallel()

		svc := New(&fakeStore{
			getUserFn: func(context.Context, string) (*User, error) {
				return &User{ID: "admin", IsAdmin: true}, nil
			},
			getExerciseFn: func(context.Context, string) (*Exercise, error) {
				return &Exercise{ID: "ex", Name: "Burpee", OwnerUserID: "other", IsCore: false}, nil
			},
			renameExerciseFn: func(context.Context, string, string) (*Exercise, error) {
				return &Exercise{ID: "ex", Name: "Burpee 2"}, nil
			},
		})
		updated, err := svc.Update(context.Background(), "admin", "ex", "Burpee 2")
		require.NoError(t, err)
		assert.Equal(t, "Burpee 2", updated.Name)
	})
}

func TestDelete(t *testing.T) {
	t.Parallel()

	t.Run("Requires admin", func(t *testing.T) {
		t.Parallel()

		svc := New(&fakeStore{
			getUserFn: func(context.Context, string) (*User, error) {
				return &User{ID: "user", IsAdmin: false}, nil
			},
			getExerciseFn: func(context.Context, string) (*Exercise, error) {
				return &Exercise{ID: "core", IsCore: true}, nil
			},
		})
		err := svc.Delete(context.Background(), "user", "core")
		require.Error(t, err)
		assert.True(t, service.IsKind(err, service.ErrorForbidden))
	})

	t.Run("Admin can delete", func(t *testing.T) {
		t.Parallel()

		svc := New(&fakeStore{
			getUserFn: func(context.Context, string) (*User, error) {
				return &User{ID: "admin", IsAdmin: true}, nil
			},
			getExerciseFn: func(context.Context, string) (*Exercise, error) {
				return &Exercise{ID: "ex", Name: "Burpee", OwnerUserID: "other"}, nil
			},
			deleteExerciseFn: func(context.Context, string) error {
				return nil
			},
		})
		err := svc.Delete(context.Background(), "admin", "ex")
		require.NoError(t, err)
	})
}
