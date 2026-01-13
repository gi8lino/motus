package workouts

import "context"

type fakeStore struct {
	createFn func(context.Context, *Workout) (*Workout, error)
	updateFn func(context.Context, *Workout) (*Workout, error)
	listFn   func(context.Context, string) ([]Workout, error)
	getFn    func(context.Context, string) (*Workout, error)
	deleteFn func(context.Context, string) error
}

func (f *fakeStore) CreateWorkout(ctx context.Context, workout *Workout) (*Workout, error) {
	if f.createFn == nil {
		return workout, nil
	}
	return f.createFn(ctx, workout)
}

func (f *fakeStore) UpdateWorkout(ctx context.Context, workout *Workout) (*Workout, error) {
	if f.updateFn == nil {
		return workout, nil
	}
	return f.updateFn(ctx, workout)
}

func (f *fakeStore) WorkoutsByUser(ctx context.Context, userID string) ([]Workout, error) {
	if f.listFn == nil {
		return nil, nil
	}
	return f.listFn(ctx, userID)
}

func (f *fakeStore) WorkoutWithSteps(ctx context.Context, id string) (*Workout, error) {
	if f.getFn == nil {
		return nil, nil
	}
	return f.getFn(ctx, id)
}

func (f *fakeStore) DeleteWorkout(ctx context.Context, id string) error {
	if f.deleteFn == nil {
		return nil
	}
	return f.deleteFn(ctx, id)
}
