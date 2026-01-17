package workouts

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"

	errpkg "github.com/gi8lino/motus/internal/service/errors"
	"github.com/gi8lino/motus/internal/service/sounds"
)

// Create stores a new workout for the user.
func (s *Service) Create(ctx context.Context, req WorkoutRequest) (*Workout, error) {
	req.UserID = strings.TrimSpace(req.UserID)
	req.Name = strings.TrimSpace(req.Name)
	if req.UserID == "" || req.Name == "" || len(req.Steps) == 0 {
		return nil, errpkg.NewError(errpkg.ErrorValidation, "userId, name, and steps are required")
	}

	steps, err := NormalizeSteps(req.Steps, sounds.ValidKey)
	if err != nil {
		return nil, errpkg.NewError(errpkg.ErrorValidation, err.Error())
	}

	workout := &Workout{UserID: req.UserID, Name: req.Name, Steps: steps}
	created, err := s.store.CreateWorkout(ctx, workout)
	if err != nil {
		return nil, errpkg.NewError(errpkg.ErrorInternal, err.Error())
	}

	return created, nil
}

// Update replaces a workout and its steps.
func (s *Service) Update(ctx context.Context, id string, req WorkoutRequest) (*Workout, error) {
	id = strings.TrimSpace(id)
	req.UserID = strings.TrimSpace(req.UserID)
	req.Name = strings.TrimSpace(req.Name)
	if id == "" {
		return nil, errpkg.NewError(errpkg.ErrorValidation, "workout id is required")
	}
	if req.Name == "" || len(req.Steps) == 0 {
		return nil, errpkg.NewError(errpkg.ErrorValidation, "name and steps are required")
	}

	steps, err := NormalizeSteps(req.Steps, sounds.ValidKey)
	if err != nil {
		return nil, errpkg.NewError(errpkg.ErrorValidation, err.Error())
	}

	workout := &Workout{ID: id, UserID: req.UserID, Name: req.Name, Steps: steps}
	updated, err := s.store.UpdateWorkout(ctx, workout)
	if err != nil {
		return nil, errpkg.NewError(errpkg.ErrorInternal, err.Error())
	}

	return updated, nil
}

// Delete removes a workout by id.
func (s *Service) Delete(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return errpkg.NewError(errpkg.ErrorValidation, "workout id is required")
	}

	if err := s.store.DeleteWorkout(ctx, id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) || isNotFoundError(err) {
			return errpkg.NewError(errpkg.ErrorNotFound, err.Error())
		}
		return errpkg.NewError(errpkg.ErrorInternal, err.Error())
	}

	return nil
}
