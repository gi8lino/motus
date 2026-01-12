package exercises

import (
	"context"

	"github.com/gi8lino/motus/internal/db"
)

// Store defines persistence operations required by the exercises domain logic.
type Store interface {
	ListExercises(ctx context.Context, userID string) ([]db.Exercise, error)
	GetUser(ctx context.Context, userID string) (*db.User, error)
	CreateExercise(ctx context.Context, name, userID string, isCore bool) (*db.Exercise, error)
	GetExercise(ctx context.Context, id string) (*db.Exercise, error)
	RenameExercise(ctx context.Context, id, name string) (*db.Exercise, error)
	DeleteExercise(ctx context.Context, id string) error
	BackfillCoreExercises(ctx context.Context) error
}

// Manager orchestrates exercise catalog rules.
type Manager struct {
	store Store
}

// NewManager builds a Manager.
func NewManager(store Store) *Manager {
	return &Manager{store: store}
}
