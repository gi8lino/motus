package exercises

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gi8lino/motus/internal/db"
	"github.com/gi8lino/motus/internal/service"
)

type fakeStore struct {
	listExercisesFn     func(context.Context, string) ([]db.Exercise, error)
	getUserFn           func(context.Context, string) (*db.User, error)
	createExerciseFn    func(context.Context, string, string, bool) (*db.Exercise, error)
	getExerciseFn       func(context.Context, string) (*db.Exercise, error)
	replaceExerciseFn   func(context.Context, string, string, string, string) error
	renameExerciseFn    func(context.Context, string, string) (*db.Exercise, error)
	deleteExerciseFn    func(context.Context, string) error
	backfillExercisesFn func(context.Context) error
}

func (f *fakeStore) ListExercises(ctx context.Context, userID string) ([]db.Exercise, error) {
	if f.listExercisesFn == nil {
		return nil, nil
	}
	return f.listExercisesFn(ctx, userID)
}

func (f *fakeStore) GetUser(ctx context.Context, userID string) (*db.User, error) {
	if f.getUserFn == nil {
		return nil, nil
	}
	return f.getUserFn(ctx, userID)
}

func (f *fakeStore) CreateExercise(ctx context.Context, name, userID string, isCore bool) (*db.Exercise, error) {
	if f.createExerciseFn == nil {
		return nil, nil
	}
	return f.createExerciseFn(ctx, name, userID, isCore)
}

func (f *fakeStore) GetExercise(ctx context.Context, id string) (*db.Exercise, error) {
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

func (f *fakeStore) RenameExercise(ctx context.Context, id, name string) (*db.Exercise, error) {
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

func TestServiceList(t *testing.T) {
	t.Parallel()

	t.Run("Validation error", func(t *testing.T) {
		t.Parallel()

		svc := &Service{Store: &fakeStore{}}
		_, err := svc.List(context.Background(), " ")
		require.Error(t, err)
		assert.True(t, service.IsKind(err, service.ErrorValidation))
	})
}

func TestServiceCreate(t *testing.T) {
	t.Parallel()

	t.Run("User not found", func(t *testing.T) {
		t.Parallel()

		svc := &Service{Store: &fakeStore{getUserFn: func(context.Context, string) (*db.User, error) {
			return nil, nil
		}}}
		_, err := svc.Create(context.Background(), "user", "Burpee", false)
		require.Error(t, err)
		assert.True(t, service.IsKind(err, service.ErrorNotFound))
	})

	t.Run("Core requires admin", func(t *testing.T) {
		t.Parallel()

		called := false
		svc := &Service{Store: &fakeStore{
			getUserFn: func(context.Context, string) (*db.User, error) {
				return &db.User{ID: "user", IsAdmin: false}, nil
			},
			createExerciseFn: func(context.Context, string, string, bool) (*db.Exercise, error) {
				called = true
				return nil, nil
			},
		}}
		_, err := svc.Create(context.Background(), "user", "Burpee", true)
		require.Error(t, err)
		assert.True(t, service.IsKind(err, service.ErrorForbidden))
		assert.False(t, called, "expected CreateExercise not to be called")
	})
}

func TestServiceUpdate(t *testing.T) {
	t.Parallel()

	t.Run("Requires admin", func(t *testing.T) {
		t.Parallel()

		svc := &Service{Store: &fakeStore{
			getUserFn: func(context.Context, string) (*db.User, error) {
				return &db.User{ID: "user", IsAdmin: false}, nil
			},
		}}
		_, err := svc.Update(context.Background(), "user", "core", "Burpee")
		require.Error(t, err)
		assert.True(t, service.IsKind(err, service.ErrorForbidden))
	})

	t.Run("Admin can rename", func(t *testing.T) {
		t.Parallel()

		svc := &Service{Store: &fakeStore{
			getUserFn: func(context.Context, string) (*db.User, error) {
				return &db.User{ID: "admin", IsAdmin: true}, nil
			},
			getExerciseFn: func(context.Context, string) (*db.Exercise, error) {
				return &db.Exercise{ID: "ex", Name: "Burpee", OwnerUserID: "other", IsCore: false}, nil
			},
			renameExerciseFn: func(context.Context, string, string) (*db.Exercise, error) {
				return &db.Exercise{ID: "ex", Name: "Burpee 2"}, nil
			},
		}}
		updated, err := svc.Update(context.Background(), "admin", "ex", "Burpee 2")
		require.NoError(t, err)
		assert.Equal(t, "Burpee 2", updated.Name)
	})
}

func TestServiceDelete(t *testing.T) {
	t.Parallel()

	t.Run("Requires admin", func(t *testing.T) {
		t.Parallel()

		svc := &Service{Store: &fakeStore{
			getUserFn: func(context.Context, string) (*db.User, error) {
				return &db.User{ID: "user", IsAdmin: false}, nil
			},
		}}
		err := svc.Delete(context.Background(), "user", "core")
		require.Error(t, err)
		assert.True(t, service.IsKind(err, service.ErrorForbidden))
	})

	t.Run("Admin can delete", func(t *testing.T) {
		t.Parallel()

		svc := &Service{Store: &fakeStore{
			getUserFn: func(context.Context, string) (*db.User, error) {
				return &db.User{ID: "admin", IsAdmin: true}, nil
			},
			getExerciseFn: func(context.Context, string) (*db.Exercise, error) {
				return &db.Exercise{ID: "ex", Name: "Burpee", OwnerUserID: "other"}, nil
			},
			deleteExerciseFn: func(context.Context, string) error {
				return nil
			},
		}}
		err := svc.Delete(context.Background(), "admin", "ex")
		require.NoError(t, err)
	})
}

func TestServiceBackfill(t *testing.T) {
	t.Parallel()

	t.Run("Internal error", func(t *testing.T) {
		t.Parallel()

		svc := &Service{Store: &fakeStore{backfillExercisesFn: func(context.Context) error {
			return errors.New("boom")
		}}}
		err := svc.Backfill(context.Background())
		require.Error(t, err)
		assert.True(t, service.IsKind(err, service.ErrorInternal))
	})
}
