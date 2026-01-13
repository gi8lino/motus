package users

import "context"

// Login validates credentials when using local authentication.
func (s *Service) Login(ctx context.Context, email, password string) (*User, error) {
	user, err := s.manager.Login(ctx, email, password)
	if err != nil {
		return nil, s.mapError(err)
	}
	return user, nil
}

// ChangePassword updates the password for the current user.
func (s *Service) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	if err := s.manager.ChangePassword(ctx, userID, currentPassword, newPassword); err != nil {
		return s.mapError(err)
	}
	return nil
}
