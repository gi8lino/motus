package workouts

import "context"

// Create stores a new workout for the user.
func (s *Service) Create(ctx context.Context, req WorkoutRequest) (*Workout, error) {
	workout, err := s.manager.Create(ctx, req)
	if err != nil {
		return nil, s.mapError(err)
	}
	return workout, nil
}

// Update replaces a workout and its steps.
func (s *Service) Update(ctx context.Context, id string, req WorkoutRequest) (*Workout, error) {
	updated, err := s.manager.Update(ctx, id, req)
	if err != nil {
		return nil, s.mapError(err)
	}
	return updated, nil
}

// Delete removes a workout by id.
func (s *Service) Delete(ctx context.Context, id string) error {
	if err := s.manager.Delete(ctx, id); err != nil {
		return s.mapError(err)
	}
	return nil
}
