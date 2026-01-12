package workouts

import (
	"context"

	"github.com/gi8lino/motus/internal/db"
)

// Store defines the persistence methods needed by workout operations.
type Store interface {
	// CreateWorkout inserts a workout definition.
	CreateWorkout(ctx context.Context, workout *db.Workout) (*db.Workout, error)
	// UpdateWorkout updates a workout definition.
	UpdateWorkout(ctx context.Context, workout *db.Workout) (*db.Workout, error)
	// WorkoutsByUser returns workouts for a user.
	WorkoutsByUser(ctx context.Context, userID string) ([]db.Workout, error)
	// WorkoutWithSteps loads a workout and its steps.
	WorkoutWithSteps(ctx context.Context, id string) (*db.Workout, error)
	// DeleteWorkout removes a workout by id.
	DeleteWorkout(ctx context.Context, id string) error
}

// Manager orchestrates workout domain rules.
type Manager struct {
	store Store
}

// NewManager builds a workout Manager.
func NewManager(store Store) *Manager {
	return &Manager{store: store}
}
