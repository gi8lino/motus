package users

import (
	"context"

	errpkg "github.com/gi8lino/motus/internal/service/errors"
)

// List returns all users.
func (s *Service) List(ctx context.Context) ([]User, error) {
	users, err := s.store.ListUsers(ctx)
	if err != nil {
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorInternal, err.Error(), errorScope)
	}
	return users, nil
}

// Get returns a user by id.
func (s *Service) Get(ctx context.Context, id string) (*User, error) {
	cleanID, err := requireEntityID(id, "user id is required")
	if err != nil {
		return nil, err
	}
	user, err := s.store.GetUser(ctx, cleanID)
	if err != nil {
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorInternal, err.Error(), errorScope)
	}
	if user == nil {
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorNotFound, "user not found", errorScope)
	}
	return user, nil
}
