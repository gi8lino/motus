package users

import (
	"context"

	"github.com/gi8lino/motus/internal/db"
)

// Store defines persistence operations required by the users domain logic.
type Store interface {
	ListUsers(ctx context.Context) ([]User, error)
	GetUser(ctx context.Context, id string) (*User, error)
	CreateUser(ctx context.Context, email, avatarURL, passwordHash string) (*User, error)
	UpdateUserAdmin(ctx context.Context, id string, isAdmin bool) error
	UpdateUserName(ctx context.Context, id, name string) error
	UpdateUserPassword(ctx context.Context, id, passwordHash string) error
	DeleteUserSessions(ctx context.Context, id string) error
	LinkAuthIdentity(ctx context.Context, userID string, identity db.AuthIdentity) error
	DeleteAuthIdentity(ctx context.Context, userID, provider, subject string) error
}
