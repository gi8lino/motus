package exercises

import (
	"context"

	"github.com/gi8lino/motus/internal/db"
)

// Create adds a new exercise to the catalog.
func (s *Service) Create(ctx context.Context, userID, name string, isCore bool) (*db.Exercise, error) {
	exercise, err := s.manager.Create(ctx, userID, name, isCore)
	if err != nil {
		return nil, s.mapError(err)
	}
	return exercise, nil
}

// Update renames an exercise entry.
func (s *Service) Update(ctx context.Context, userID, exerciseID, name string) (*db.Exercise, error) {
	updated, err := s.manager.Update(ctx, userID, exerciseID, name)
	if err != nil {
		return nil, s.mapError(err)
	}
	return updated, nil
}

// Delete removes an exercise from the catalog.
func (s *Service) Delete(ctx context.Context, userID, exerciseID string) error {
	if err := s.manager.Delete(ctx, userID, exerciseID); err != nil {
		return s.mapError(err)
	}
	return nil
}
