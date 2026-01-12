package exercises

import (
	"context"

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
