package templates

import (
	"context"

	errpkg "github.com/gi8lino/motus/internal/service/errors"
)

// Create marks a workout as a template.
func (s *Service) Create(ctx context.Context, workoutID, name string) (*Workout, error) {
	wid, err := requireID(workoutID, "workoutId is required")
	if err != nil {
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorValidation, err.Error(), errorScope)
	}

	template, err := s.store.CreateTemplateFromWorkout(ctx, wid, name)
	if err != nil {
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorInternal, err.Error(), errorScope)
	}
	return template, nil
}
