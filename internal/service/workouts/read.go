package workouts

import "context"

// Get returns a workout by id.
func (s *Service) Get(ctx context.Context, id string) (*Workout, error) {
	workout, err := s.manager.Get(ctx, id)
	if err != nil {
		return nil, s.mapError(err)
	}

	return workout, nil
}

// Export returns a workout for sharing.
func (s *Service) Export(ctx context.Context, id string) (*Workout, error) {
	workout, err := s.manager.Export(ctx, id)
	if err != nil {
		return nil, s.mapError(err)
	}
	return workout, nil
}

// List returns workouts for the given user.
func (s *Service) List(ctx context.Context, userID string) ([]Workout, error) {
	workouts, err := s.manager.List(ctx, userID)
	if err != nil {
		return nil, s.mapError(err)
	}
	return workouts, nil
}
