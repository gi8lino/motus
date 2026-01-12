package templates

import "context"

// Store defines persistence used by the template domain.
type Store interface {
	ListTemplates(ctx context.Context) ([]Workout, error)
	CreateTemplateFromWorkout(ctx context.Context, workoutID, name string) (*Workout, error)
	WorkoutWithSteps(ctx context.Context, id string) (*Workout, error)
	CreateWorkoutFromTemplate(ctx context.Context, templateID, userID, name string) (*Workout, error)
}
