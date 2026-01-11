package templates

import (
	"context"

	"github.com/gi8lino/motus/internal/db"
)

// Create records a workout as a template.
func (m *Manager) Create(ctx context.Context, workoutID, name string) (*db.Workout, error) {
	wid, err := requireID(workoutID, "workoutId is required")
	if err != nil {
		return nil, err
	}

	template, err := m.store.CreateTemplateFromWorkout(ctx, wid, name)
	if err != nil {
		return nil, internal(err)
	}
	return template, nil
}
