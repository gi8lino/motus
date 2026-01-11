package users

import (
	"context"

	"github.com/gi8lino/motus/internal/db"
)

// Store defines persistence operations required by the users domain logic.
type Store interface {
	ListUsers(ctx context.Context) ([]db.User, error)
	GetUser(ctx context.Context, id string) (*db.User, error)
	CreateUser(ctx context.Context, email, avatarURL, passwordHash string) (*db.User, error)
	UpdateUserAdmin(ctx context.Context, id string, isAdmin bool) error
	GetUserWithPassword(ctx context.Context, id string) (*db.User, string, error)
	UpdateUserPassword(ctx context.Context, id, passwordHash string) error
	UpdateUserName(ctx context.Context, id, name string) error
}

// Manager orchestrates user domain rules.
type Manager struct {
	store             Store
	authHeader        string
	allowRegistration bool
}

// NewManager creates a Manager instance.
func NewManager(store Store, authHeader string, allowRegistration bool) *Manager {
	return &Manager{store: store, authHeader: authHeader, allowRegistration: allowRegistration}
}
