package users

import "context"

// Create registers a new local user.
func (s *Service) Create(ctx context.Context, email, avatarURL, password string) (*User, error) {
	user, err := s.manager.Create(ctx, email, avatarURL, password)
	if err != nil {
		return nil, s.mapError(err)
	}
	return user, nil
}

// UpdateRole toggles admin access.
func (s *Service) UpdateRole(ctx context.Context, id string, isAdmin bool) error {
	if err := s.manager.UpdateRole(ctx, id, isAdmin); err != nil {
		return s.mapError(err)
	}
	return nil
}

// UpdateName changes the display name for the user profile.
func (s *Service) UpdateName(ctx context.Context, userID, name string) error {
	if err := s.manager.UpdateName(ctx, userID, name); err != nil {
		return s.mapError(err)
	}
	return nil
}
