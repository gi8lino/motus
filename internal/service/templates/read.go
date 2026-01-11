package templates

import (
	"context"

	"github.com/gi8lino/motus/internal/db"
)

// List returns all shared templates.
func (s *Service) List(ctx context.Context) ([]db.Workout, error) {
	templates, err := s.manager.List(ctx)
	if err != nil {
		return nil, s.mapError(err)
	}
	return templates, nil
}

// Get returns a template by id.
func (s *Service) Get(ctx context.Context, id string) (*db.Workout, error) {
	template, err := s.manager.Get(ctx, id)
	if err != nil {
		return nil, s.mapError(err)
	}
	return template, nil
}
