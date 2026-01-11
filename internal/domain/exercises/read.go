package exercises

import (
	"context"

	"github.com/gi8lino/motus/internal/db"
)

// List returns the available exercises for a user.
func (m *Manager) List(ctx context.Context, userID string) ([]db.Exercise, error) {
	uid, err := requireUserID(userID)
	if err != nil {
		return nil, err
	}
	exercises, err := m.store.ListExercises(ctx, uid)
	if err != nil {
		return nil, internal(err)
	}
	return exercises, nil
}
