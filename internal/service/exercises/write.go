package exercises

import (
	"context"

	errpkg "github.com/gi8lino/motus/internal/service/errors"
)

// Create adds a new exercise to the catalog.
func (s *Service) Create(ctx context.Context, userID, name string, isCore bool) (*Exercise, error) {
	uid, err := requireUserID(userID)
	if err != nil {
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorValidation, err.Error(), errorScope)
	}
	clean, err := requireName(name)
	if err != nil {
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorValidation, err.Error(), errorScope)
	}

	user, err := s.store.GetUser(ctx, uid)
	if err != nil {
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorInternal, err.Error(), errorScope)
	}
	if user == nil {
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorNotFound, "user not found", errorScope)
	}
	if isCore && !user.IsAdmin {
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorForbidden, "core exercises require admin permissions", errorScope)
	}

	exercise, err := s.store.CreateExercise(ctx, clean, uid, isCore)
	if err != nil {
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorInternal, err.Error(), errorScope)
	}
	return exercise, nil
}

// Update renames an exercise entry.
func (s *Service) Update(ctx context.Context, userID, exerciseID, name string) (*Exercise, error) {
	uid, err := requireUserID(userID)
	if err != nil {
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorValidation, err.Error(), errorScope)
	}
	clean, err := requireName(name)
	if err != nil {
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorValidation, err.Error(), errorScope)
	}
	id, err := requireEntityID(exerciseID, "exercise id is required")
	if err != nil {
		return nil, err
	}

	user, err := s.store.GetUser(ctx, uid)
	if err != nil {
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorInternal, err.Error(), errorScope)
	}
	if user == nil {
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorNotFound, "user not found", errorScope)
	}

	exercise, err := s.store.GetExercise(ctx, id)
	if err != nil {
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorInternal, err.Error(), errorScope)
	}
	if exercise == nil {
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorNotFound, "exercise not found", errorScope)
	}

	if exercise.IsCore && !user.IsAdmin {
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorForbidden, "core exercises require admin permissions", errorScope)
	}

	if exercise.OwnerUserID != "" && exercise.OwnerUserID != uid && !user.IsAdmin {
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorForbidden, "exercise belongs to another user", errorScope)
	}

	updated, err := s.store.RenameExercise(ctx, id, clean)
	if err != nil {
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorInternal, err.Error(), errorScope)
	}
	return updated, nil
}

// Delete removes an exercise from the catalog.
func (s *Service) Delete(ctx context.Context, userID, exerciseID string) error {
	uid, err := requireUserID(userID)
	if err != nil {
		return errpkg.NewErrorWithScope(errpkg.ErrorValidation, err.Error(), errorScope)
	}
	id, err := requireEntityID(exerciseID, "exercise id is required")
	if err != nil {
		return errpkg.NewErrorWithScope(errpkg.ErrorValidation, err.Error(), errorScope)
	}

	user, err := s.store.GetUser(ctx, uid)
	if err != nil {
		return errpkg.NewErrorWithScope(errpkg.ErrorInternal, err.Error(), errorScope)
	}
	if user == nil {
		return errpkg.NewErrorWithScope(errpkg.ErrorNotFound, "user not found", errorScope)
	}

	exercise, err := s.store.GetExercise(ctx, id)
	if err != nil {
		return errpkg.NewErrorWithScope(errpkg.ErrorInternal, err.Error(), errorScope)
	}
	if exercise == nil {
		return errpkg.NewErrorWithScope(errpkg.ErrorNotFound, "exercise not found", errorScope)
	}

	if exercise.IsCore && !user.IsAdmin {
		return errpkg.NewErrorWithScope(errpkg.ErrorForbidden, "core exercises require admin permissions", errorScope)
	}

	if exercise.OwnerUserID != "" && exercise.OwnerUserID != uid && !user.IsAdmin {
		return errpkg.NewErrorWithScope(errpkg.ErrorForbidden, "exercise belongs to another user", errorScope)
	}

	if err := s.store.DeleteExercise(ctx, id); err != nil {
		return errpkg.NewErrorWithScope(errpkg.ErrorInternal, err.Error(), errorScope)
	}
	return nil
}
