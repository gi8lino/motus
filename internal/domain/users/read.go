package users

import (
	"context"

	"github.com/gi8lino/motus/internal/db"
)

// List returns all users.
func (m *Manager) List(ctx context.Context) ([]db.User, error) {
	users, err := m.store.ListUsers(ctx)
	if err != nil {
		return nil, internal(err)
	}
	return users, nil
}
