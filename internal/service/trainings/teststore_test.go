package trainings

import "context"

type fakeStore struct {
	stepTimingsFn func(context.Context, string) ([]TrainingStepLog, error)
	workoutFn     func(context.Context, string) (*Workout, error)
	recordFn      func(context.Context, TrainingLog, []TrainingStepLog) error
	historyFn     func(context.Context, string, int) ([]TrainingLog, error)
}

func (f *fakeStore) TrainingStepTimings(ctx context.Context, trainingID string) ([]TrainingStepLog, error) {
	if f.stepTimingsFn == nil {
		return nil, nil
	}
	return f.stepTimingsFn(ctx, trainingID)
}

func (f *fakeStore) WorkoutWithSteps(ctx context.Context, id string) (*Workout, error) {
	if f.workoutFn == nil {
		return nil, nil
	}
	return f.workoutFn(ctx, id)
}

func (f *fakeStore) RecordTraining(ctx context.Context, log TrainingLog, steps []TrainingStepLog) error {
	if f.recordFn == nil {
		return nil
	}
	return f.recordFn(ctx, log, steps)
}

func (f *fakeStore) TrainingHistory(ctx context.Context, userID string, limit int) ([]TrainingLog, error) {
	if f.historyFn == nil {
		return nil, nil
	}
	return f.historyFn(ctx, userID, limit)
}
