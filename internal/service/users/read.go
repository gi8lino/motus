package users

import (
	"context"

	"github.com/gi8lino/motus/internal/db"
)

// List returns all users.
func (s *Service) List(ctx context.Context) ([]db.User, error) {
	users, err := s.manager.List(ctx)
	if err != nil {
		return nil, s.mapError(err)
	}
	return users, nil
}
