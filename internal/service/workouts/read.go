package workouts

import (
	"context"

	"github.com/gi8lino/motus/internal/db"
)

// Get returns a workout by id.
func (s *Service) Get(ctx context.Context, id string) (*db.Workout, error) {
	workout, err := s.manager.Get(ctx, id)
	if err != nil {
		return nil, s.mapError(err)
	}

	return workout, nil
}

// Export returns a workout for sharing.
func (s *Service) Export(ctx context.Context, id string) (*db.Workout, error) {
	workout, err := s.manager.Export(ctx, id)
	if err != nil {
		return nil, s.mapError(err)
	}
	return workout, nil
}
