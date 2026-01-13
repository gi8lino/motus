package users

import "context"

// List returns all users.
func (s *Service) List(ctx context.Context) ([]User, error) {
	users, err := s.manager.List(ctx)
	if err != nil {
		return nil, s.mapError(err)
	}
	return users, nil
}

// Get returns a user by id.
func (s *Service) Get(ctx context.Context, id string) (*User, error) {
	user, err := s.manager.Get(ctx, id)
	if err != nil {
		return nil, s.mapError(err)
	}
	return user, nil
}
