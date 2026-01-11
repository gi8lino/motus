package workouts

import (
	"context"
	"strings"

	"github.com/gi8lino/motus/internal/db"
)

// List returns workouts for the given user.
func (m *Manager) List(ctx context.Context, userID string) ([]db.Workout, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, validation("userId is required")
	}
	workouts, err := m.store.WorkoutsByUser(ctx, userID)
	if err != nil {
		return nil, internal(err)
	}
	return workouts, nil
}

// Get fetches a workout by id.
func (m *Manager) Get(ctx context.Context, id string) (*db.Workout, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, validation("workout id is required")
	}
	workout, err := m.store.WorkoutWithSteps(ctx, id)
	if err != nil {
		if isNotFoundError(err) {
			return nil, notFound("workout not found")
		}
		return nil, internal(err)
	}
	if workout == nil {
		return nil, notFound("workout not found")
	}
	return workout, nil
}

// Export returns a workout for sharing.
func (m *Manager) Export(ctx context.Context, id string) (*db.Workout, error) {
	return m.Get(ctx, id)
}
