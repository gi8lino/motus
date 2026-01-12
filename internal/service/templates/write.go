package templates

import "context"

// Create marks a workout as a template.
func (s *Service) Create(ctx context.Context, workoutID, name string) (*Workout, error) {
	template, err := s.manager.Create(ctx, workoutID, name)
	if err != nil {
		return nil, s.mapError(err)
	}
	return template, nil
}
