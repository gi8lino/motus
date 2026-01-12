package templates

import (
	"context"

	"github.com/gi8lino/motus/internal/db"
)

// Apply clones a template to a workout for the user.
func (m *Manager) Apply(ctx context.Context, templateID, userID, name string) (*db.Workout, error) {
	tid, err := requireID(templateID, "template id is required")
	if err != nil {
		return nil, err
	}
	uid, err := requireID(userID, "userId is required")
	if err != nil {
		return nil, err
	}

	workout, err := m.store.CreateWorkoutFromTemplate(ctx, tid, uid, name)
	if err != nil {
		return nil, internal(err)
	}
	return workout, nil
}
