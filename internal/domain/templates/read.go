package templates

import (
	"context"

	"github.com/gi8lino/motus/internal/db"
)

// List returns all available templates.
func (m *Manager) List(ctx context.Context) ([]db.Workout, error) {
	list, err := m.store.ListTemplates(ctx)
	if err != nil {
		return nil, internal(err)
	}
	return list, nil
}

// Get fetches a template by id.
func (m *Manager) Get(ctx context.Context, id string) (*db.Workout, error) {
	tid, err := requireID(id, "template id is required")
	if err != nil {
		return nil, err
	}

	template, err := m.store.WorkoutWithSteps(ctx, tid)
	if err != nil {
		return nil, internal(err)
	}
	if template == nil || !template.IsTemplate {
		return nil, notFound("template not found")
	}
	return template, nil
}
