package exercises

import "context"

// List returns the exercise catalog for a user.
func (s *Service) List(ctx context.Context, userID string) ([]Exercise, error) {
	exercises, err := s.manager.List(ctx, userID)
	if err != nil {
		return nil, s.mapError(err)
	}
	return exercises, nil
}
