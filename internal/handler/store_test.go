package handler

import (
	"context"

	"github.com/gi8lino/motus/internal/db"
)

type fakeStore struct {
	pingFn                    func(context.Context) error
	getUserFn                 func(context.Context, string) (*db.User, error)
	listUsersFn               func(context.Context) ([]db.User, error)
	getUserWithPasswordFn     func(context.Context, string) (*db.User, string, error)
	updateUserPasswordFn      func(context.Context, string, string) error
	updateUserAdminFn         func(context.Context, string, bool) error
	updateUserNameFn          func(context.Context, string, string) error
	createUserFn              func(context.Context, string, string, string) (*db.User, error)
	listExercisesFn           func(context.Context, string) ([]db.Exercise, error)
	createExerciseFn          func(context.Context, string, string, bool) (*db.Exercise, error)
	getExerciseFn             func(context.Context, string) (*db.Exercise, error)
	renameExerciseFn          func(context.Context, string, string) (*db.Exercise, error)
	replaceExerciseForUserFn  func(context.Context, string, string, string, string) error
	deleteExerciseFn          func(context.Context, string) error
	backfillCoreExercisesFn   func(context.Context) error
	listTemplatesFn           func(context.Context) ([]db.Workout, error)
	createTemplateFn          func(context.Context, string, string) (*db.Workout, error)
	createWorkoutFromTemplate func(context.Context, string, string, string) (*db.Workout, error)
	workoutsByUserFn          func(context.Context, string) ([]db.Workout, error)
	createWorkoutFn           func(context.Context, *db.Workout) (*db.Workout, error)
	updateWorkoutFn           func(context.Context, *db.Workout) (*db.Workout, error)
	workoutWithStepsFn        func(context.Context, string) (*db.Workout, error)
	deleteWorkoutFn           func(context.Context, string) error
	sessionHistoryFn          func(context.Context, string, int) ([]db.SessionLog, error)
	sessionStepTimingsFn      func(context.Context, string) ([]db.SessionStepLog, error)
	recordSessionFn           func(context.Context, db.SessionLog, []db.SessionStepLog) error
}

func (f *fakeStore) Ping(ctx context.Context) error {
	if f.pingFn == nil {
		return nil
	}
	return f.pingFn(ctx)
}

func (f *fakeStore) GetUser(ctx context.Context, id string) (*db.User, error) {
	if f.getUserFn == nil {
		return nil, nil
	}
	return f.getUserFn(ctx, id)
}

func (f *fakeStore) ListUsers(ctx context.Context) ([]db.User, error) {
	if f.listUsersFn == nil {
		return nil, nil
	}
	return f.listUsersFn(ctx)
}

func (f *fakeStore) GetUserWithPassword(ctx context.Context, id string) (*db.User, string, error) {
	if f.getUserWithPasswordFn == nil {
		return nil, "", nil
	}
	return f.getUserWithPasswordFn(ctx, id)
}

func (f *fakeStore) UpdateUserPassword(ctx context.Context, id, passwordHash string) error {
	if f.updateUserPasswordFn == nil {
		return nil
	}
	return f.updateUserPasswordFn(ctx, id, passwordHash)
}

func (f *fakeStore) UpdateUserAdmin(ctx context.Context, id string, isAdmin bool) error {
	if f.updateUserAdminFn == nil {
		return nil
	}
	return f.updateUserAdminFn(ctx, id, isAdmin)
}

func (f *fakeStore) UpdateUserName(ctx context.Context, id, name string) error {
	if f.updateUserNameFn == nil {
		return nil
	}
	return f.updateUserNameFn(ctx, id, name)
}

func (f *fakeStore) CreateUser(ctx context.Context, email, avatarURL, passwordHash string) (*db.User, error) {
	if f.createUserFn == nil {
		return nil, nil
	}
	return f.createUserFn(ctx, email, avatarURL, passwordHash)
}

func (f *fakeStore) ListExercises(ctx context.Context, userID string) ([]db.Exercise, error) {
	if f.listExercisesFn == nil {
		return nil, nil
	}
	return f.listExercisesFn(ctx, userID)
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

func (f *fakeStore) RenameExercise(ctx context.Context, id, name string) (*db.Exercise, error) {
	if f.renameExerciseFn == nil {
		return nil, nil
	}
	return f.renameExerciseFn(ctx, id, name)
}

func (f *fakeStore) ReplaceExerciseForUser(ctx context.Context, userID, oldID, newID, newName string) error {
	if f.replaceExerciseForUserFn == nil {
		return nil
	}
	return f.replaceExerciseForUserFn(ctx, userID, oldID, newID, newName)
}

func (f *fakeStore) DeleteExercise(ctx context.Context, id string) error {
	if f.deleteExerciseFn == nil {
		return nil
	}
	return f.deleteExerciseFn(ctx, id)
}

func (f *fakeStore) BackfillCoreExercises(ctx context.Context) error {
	if f.backfillCoreExercisesFn == nil {
		return nil
	}
	return f.backfillCoreExercisesFn(ctx)
}

func (f *fakeStore) ListTemplates(ctx context.Context) ([]db.Workout, error) {
	if f.listTemplatesFn == nil {
		return nil, nil
	}
	return f.listTemplatesFn(ctx)
}

func (f *fakeStore) CreateTemplateFromWorkout(ctx context.Context, workoutID, name string) (*db.Workout, error) {
	if f.createTemplateFn == nil {
		return nil, nil
	}
	return f.createTemplateFn(ctx, workoutID, name)
}

func (f *fakeStore) CreateWorkoutFromTemplate(ctx context.Context, templateID, userID, name string) (*db.Workout, error) {
	if f.createWorkoutFromTemplate == nil {
		return nil, nil
	}
	return f.createWorkoutFromTemplate(ctx, templateID, userID, name)
}

func (f *fakeStore) WorkoutsByUser(ctx context.Context, userID string) ([]db.Workout, error) {
	if f.workoutsByUserFn == nil {
		return nil, nil
	}
	return f.workoutsByUserFn(ctx, userID)
}

func (f *fakeStore) CreateWorkout(ctx context.Context, workout *db.Workout) (*db.Workout, error) {
	if f.createWorkoutFn == nil {
		return workout, nil
	}
	return f.createWorkoutFn(ctx, workout)
}

func (f *fakeStore) UpdateWorkout(ctx context.Context, workout *db.Workout) (*db.Workout, error) {
	if f.updateWorkoutFn == nil {
		return workout, nil
	}
	return f.updateWorkoutFn(ctx, workout)
}

func (f *fakeStore) WorkoutWithSteps(ctx context.Context, id string) (*db.Workout, error) {
	if f.workoutWithStepsFn == nil {
		return nil, nil
	}
	return f.workoutWithStepsFn(ctx, id)
}

func (f *fakeStore) DeleteWorkout(ctx context.Context, id string) error {
	if f.deleteWorkoutFn == nil {
		return nil
	}
	return f.deleteWorkoutFn(ctx, id)
}

func (f *fakeStore) SessionHistory(ctx context.Context, userID string, limit int) ([]db.SessionLog, error) {
	if f.sessionHistoryFn == nil {
		return nil, nil
	}
	return f.sessionHistoryFn(ctx, userID, limit)
}

func (f *fakeStore) SessionStepTimings(ctx context.Context, sessionID string) ([]db.SessionStepLog, error) {
	if f.sessionStepTimingsFn == nil {
		return nil, nil
	}
	return f.sessionStepTimingsFn(ctx, sessionID)
}

func (f *fakeStore) RecordSession(ctx context.Context, log db.SessionLog, steps []db.SessionStepLog) error {
	if f.recordSessionFn == nil {
		return nil
	}
	return f.recordSessionFn(ctx, log, steps)
}
