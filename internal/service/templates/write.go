package templates

import (
	"context"
	errpkg "github.com/gi8lino/motus/internal/service/errors"
)

// Create marks a workout as a template.
func (s *Service) Create(ctx context.Context, workoutID, name string) (*Workout, error) {
	wid, err := requireID(workoutID, "workoutId is required")
	if err != nil {
		return nil, err
	}

	template, err := s.store.CreateTemplateFromWorkout(ctx, wid, name)
	if err != nil {
		return nil, errpkg.NewError(errpkg.ErrorInternal, err.Error())
	}
	return template, nil
}
