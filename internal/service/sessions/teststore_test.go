package sessions

import "context"

type fakeStore struct {
	stepTimingsFn func(context.Context, string) ([]SessionStepLog, error)
	workoutFn     func(context.Context, string) (*Workout, error)
	recordFn      func(context.Context, SessionLog, []SessionStepLog) error
	historyFn     func(context.Context, string, int) ([]SessionLog, error)
}

func (f *fakeStore) SessionStepTimings(ctx context.Context, sessionID string) ([]SessionStepLog, error) {
	if f.stepTimingsFn == nil {
		return nil, nil
	}
	return f.stepTimingsFn(ctx, sessionID)
}

func (f *fakeStore) WorkoutWithSteps(ctx context.Context, id string) (*Workout, error) {
	if f.workoutFn == nil {
		return nil, nil
	}
	return f.workoutFn(ctx, id)
}

func (f *fakeStore) RecordSession(ctx context.Context, log SessionLog, steps []SessionStepLog) error {
	if f.recordFn == nil {
		return nil
	}
	return f.recordFn(ctx, log, steps)
}

func (f *fakeStore) SessionHistory(ctx context.Context, userID string, limit int) ([]SessionLog, error) {
	if f.historyFn == nil {
		return nil, nil
	}
	return f.historyFn(ctx, userID, limit)
}
