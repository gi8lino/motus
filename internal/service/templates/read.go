package templates

import (
	"context"
	errpkg "github.com/gi8lino/motus/internal/service/errors"
)

// List returns all shared templates.
func (s *Service) List(ctx context.Context) ([]Workout, error) {
	list, err := s.store.ListTemplates(ctx)
	if err != nil {
		return nil, errpkg.NewError(errpkg.ErrorInternal, err.Error())
	}
	return list, nil
}

// Get returns a template by id.
func (s *Service) Get(ctx context.Context, id string) (*Workout, error) {
	tid, err := requireID(id, "template id is required")
	if err != nil {
		return nil, err
	}

	template, err := s.store.WorkoutWithSteps(ctx, tid)
	if err != nil {
		return nil, errpkg.NewError(errpkg.ErrorInternal, err.Error())
	}
	if template == nil || !template.IsTemplate {
		return nil, errpkg.NewError(errpkg.ErrorNotFound, "template not found")
	}
	return template, nil
}
