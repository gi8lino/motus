package users

import "context"

// Store defines persistence operations required by the users domain logic.
type Store interface {
	ListUsers(ctx context.Context) ([]User, error)
	GetUser(ctx context.Context, id string) (*User, error)
	CreateUser(ctx context.Context, email, avatarURL, passwordHash string) (*User, error)
	UpdateUserAdmin(ctx context.Context, id string, isAdmin bool) error
	GetUserWithPassword(ctx context.Context, id string) (*User, string, error)
	UpdateUserPassword(ctx context.Context, id, passwordHash string) error
	UpdateUserName(ctx context.Context, id, name string) error
}
