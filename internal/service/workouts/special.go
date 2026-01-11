package workouts

import (
	"context"

	"github.com/gi8lino/motus/internal/db"
)

// Import creates a new workout from exported JSON.
func (s *Service) Import(ctx context.Context, userID string, workout db.Workout) (*db.Workout, error) {
	created, err := s.manager.Import(ctx, userID, workout)
	if err != nil {
		return nil, s.mapError(err)
	}

	return created, nil
}
