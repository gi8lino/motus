package templates

import (
	"context"

	errpkg "github.com/gi8lino/motus/internal/service/errors"
)

// List returns all shared templates.
func (s *Service) List(ctx context.Context) ([]Workout, error) {
	list, err := s.store.ListTemplates(ctx)
	if err != nil {
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorInternal, err.Error(), errorScope)
	}
	return list, nil
}

// Get returns a template by id.
func (s *Service) Get(ctx context.Context, id string) (*Workout, error) {
	tid, err := requireID(id, "template id is required")
	if err != nil {
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorValidation, err.Error(), errorScope)
	}

	template, err := s.store.WorkoutWithSteps(ctx, tid)
	if err != nil {
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorInternal, err.Error(), errorScope)
	}
	if template == nil || !template.IsTemplate {
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorNotFound, "template not found", errorScope)
	}
	return template, nil
}
