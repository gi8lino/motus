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
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorValidation, "workout id is required", errorScope)
	}
	workout, err := s.store.WorkoutWithSteps(ctx, id)
	if err != nil {
		if isNotFoundError(err) {
			return nil, errpkg.NewErrorWithScope(errpkg.ErrorNotFound, err.Error(), errorScope)
		}
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorInternal, err.Error(), errorScope)
	}
	if workout == nil {
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorNotFound, "workout not found", errorScope)
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
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorValidation, "user id is required", errorScope)
	}
	workouts, err := s.store.WorkoutsByUser(ctx, userID)
	if err != nil {
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorInternal, err.Error(), errorScope)
	}
	return workouts, nil
}
