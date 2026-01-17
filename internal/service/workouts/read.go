package workouts

import (
	"context"
	"strings"

	errpkg "github.com/gi8lino/motus/internal/service/errors"
)

// Get returns a workout by id.
func (s *Service) Get(ctx context.Context, id string) (*Workout, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errpkg.NewError(errpkg.ErrorValidation, "workout id is required")
	}
	workout, err := s.store.WorkoutWithSteps(ctx, id)
	if err != nil {
		if isNotFoundError(err) {
			return nil, errpkg.NewError(errpkg.ErrorNotFound, err.Error())
		}
		return nil, errpkg.NewError(errpkg.ErrorInternal, err.Error())
	}
	if workout == nil {
		return nil, errpkg.NewError(errpkg.ErrorNotFound, "workout not found")
	}
	return workout, nil
}

// Export returns a workout for sharing.
func (s *Service) Export(ctx context.Context, id string) (*Workout, error) {
	workout, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return workout, nil
}

// List returns workouts for the given user.
func (s *Service) List(ctx context.Context, userID string) ([]Workout, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, errpkg.NewError(errpkg.ErrorValidation, "user id is required")
	}
	workouts, err := s.store.WorkoutsByUser(ctx, userID)
	if err != nil {
		return nil, errpkg.NewError(errpkg.ErrorInternal, err.Error())
	}
	return workouts, nil
}
