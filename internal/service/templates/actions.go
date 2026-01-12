package templates

import "context"

// Apply clones a template into a new workout.
func (s *Service) Apply(ctx context.Context, templateID, userID, name string) (*Workout, error) {
	workout, err := s.manager.Apply(ctx, templateID, userID, name)
	if err != nil {
		return nil, s.mapError(err)
	}
	return workout, nil
}
