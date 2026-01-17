package workouts

import "context"

// Store defines the persistence methods needed by workout operations.
type Store interface {
	CreateWorkout(ctx context.Context, workout *Workout) (*Workout, error)
	UpdateWorkout(ctx context.Context, workout *Workout) (*Workout, error)
	WorkoutsByUser(ctx context.Context, userID string) ([]Workout, error)
	WorkoutWithSteps(ctx context.Context, id string) (*Workout, error)
	DeleteWorkout(ctx context.Context, id string) error
}
