package templates

import (
	"context"

	"github.com/gi8lino/motus/internal/db"
)

// Store defines persistence used by the template domain.
type Store interface {
	ListTemplates(ctx context.Context) ([]db.Workout, error)
	CreateTemplateFromWorkout(ctx context.Context, workoutID, name string) (*db.Workout, error)
	WorkoutWithSteps(ctx context.Context, id string) (*db.Workout, error)
	CreateWorkoutFromTemplate(ctx context.Context, templateID, userID, name string) (*db.Workout, error)
}

// Manager contains the template business rules.
type Manager struct {
	store Store
}

// NewManager builds a template Manager.
func NewManager(store Store) *Manager {
	return &Manager{store: store}
}
