package exercises

import "context"

// Store defines persistence operations required by the exercises domain.
type Store interface {
	ListExercises(ctx context.Context, userID string) ([]Exercise, error)
	GetUser(ctx context.Context, userID string) (*User, error)
	CreateExercise(ctx context.Context, name, userID string, isCore bool) (*Exercise, error)
	GetExercise(ctx context.Context, id string) (*Exercise, error)
	RenameExercise(ctx context.Context, id, name string) (*Exercise, error)
	DeleteExercise(ctx context.Context, id string) error
	BackfillCoreExercises(ctx context.Context) error
}
