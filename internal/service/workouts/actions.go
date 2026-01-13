package workouts

import "context"

// Import creates a new workout from exported JSON.
func (s *Service) Import(ctx context.Context, userID string, workout Workout) (*Workout, error) {
	created, err := s.manager.Import(ctx, userID, workout)
	if err != nil {
		return nil, s.mapError(err)
	}

	return created, nil
}
