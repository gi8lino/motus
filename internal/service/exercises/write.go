package exercises

import (
	"context"

	errpkg "github.com/gi8lino/motus/internal/service/errors"
)

// Create adds a new exercise to the catalog.
func (s *Service) Create(ctx context.Context, userID, name string, isCore bool) (*Exercise, error) {
	uid, err := requireUserID(userID)
	if err != nil {
		return nil, err
	}
	clean, err := requireName(name)
	if err != nil {
		return nil, err
	}

	user, err := s.store.GetUser(ctx, uid)
	if err != nil {
		return nil, errpkg.NewError(errpkg.ErrorInternal, err.Error())
	}
	if user == nil {
		return nil, errpkg.NewError(errpkg.ErrorNotFound, "user not found")
	}
	if isCore && !user.IsAdmin {
		return nil, errpkg.NewError(errpkg.ErrorForbidden, "core exercises require admin permissions")
	}

	exercise, err := s.store.CreateExercise(ctx, clean, uid, isCore)
	if err != nil {
		return nil, errpkg.NewError(errpkg.ErrorInternal, err.Error())
	}
	return exercise, nil
}

// Update renames an exercise entry.
func (s *Service) Update(ctx context.Context, userID, exerciseID, name string) (*Exercise, error) {
	uid, err := requireUserID(userID)
	if err != nil {
		return nil, err
	}
	clean, err := requireName(name)
	if err != nil {
		return nil, err
	}
	id, err := requireEntityID(exerciseID, "exercise id is required")
	if err != nil {
		return nil, err
	}

	user, err := s.store.GetUser(ctx, uid)
	if err != nil {
		return nil, errpkg.NewError(errpkg.ErrorInternal, err.Error())
	}
	if user == nil {
		return nil, errpkg.NewError(errpkg.ErrorNotFound, "user not found")
	}

	exercise, err := s.store.GetExercise(ctx, id)
	if err != nil {
		return nil, errpkg.NewError(errpkg.ErrorInternal, err.Error())
	}
	if exercise == nil {
		return nil, errpkg.NewError(errpkg.ErrorNotFound, "exercise not found")
	}

	if exercise.IsCore && !user.IsAdmin {
		return nil, errpkg.NewError(errpkg.ErrorForbidden, "core exercises require admin permissions")
	}

	if exercise.OwnerUserID != "" && exercise.OwnerUserID != uid && !user.IsAdmin {
		return nil, errpkg.NewError(errpkg.ErrorForbidden, "exercise belongs to another user")
	}

	updated, err := s.store.RenameExercise(ctx, id, clean)
	if err != nil {
		return nil, errpkg.NewError(errpkg.ErrorInternal, err.Error())
	}
	return updated, nil
}

// Delete removes an exercise from the catalog.
func (s *Service) Delete(ctx context.Context, userID, exerciseID string) error {
	uid, err := requireUserID(userID)
	if err != nil {
		return err
	}
	id, err := requireEntityID(exerciseID, "exercise id is required")
	if err != nil {
		return err
	}

	user, err := s.store.GetUser(ctx, uid)
	if err != nil {
		return errpkg.NewError(errpkg.ErrorInternal, err.Error())
	}
	if user == nil {
		return errpkg.NewError(errpkg.ErrorNotFound, "user not found")
	}

	exercise, err := s.store.GetExercise(ctx, id)
	if err != nil {
		return errpkg.NewError(errpkg.ErrorInternal, err.Error())
	}
	if exercise == nil {
		return errpkg.NewError(errpkg.ErrorNotFound, "exercise not found")
	}

	if exercise.IsCore && !user.IsAdmin {
		return errpkg.NewError(errpkg.ErrorForbidden, "core exercises require admin permissions")
	}

	if exercise.OwnerUserID != "" && exercise.OwnerUserID != uid && !user.IsAdmin {
		return errpkg.NewError(errpkg.ErrorForbidden, "exercise belongs to another user")
	}

	if err := s.store.DeleteExercise(ctx, id); err != nil {
		return errpkg.NewError(errpkg.ErrorInternal, err.Error())
	}
	return nil
}
