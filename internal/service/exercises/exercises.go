package exercises

import (
	"context"
	"strings"

	"github.com/gi8lino/motus/internal/db"
	"github.com/gi8lino/motus/internal/service"
)

// store defines the persistence methods needed by the exercises service.
type store interface {
	ListExercises(ctx context.Context, userID string) ([]db.Exercise, error)
	GetUser(ctx context.Context, userID string) (*db.User, error)
	CreateExercise(ctx context.Context, name, userID string, isCore bool) (*db.Exercise, error)
	GetExercise(ctx context.Context, id string) (*db.Exercise, error)
	ReplaceExerciseForUser(ctx context.Context, userID, oldID, newID, newName string) error
	RenameExercise(ctx context.Context, id, name string) (*db.Exercise, error)
	DeleteExercise(ctx context.Context, id string) error
	BackfillCoreExercises(ctx context.Context) error
}

// Service coordinates exercise catalog operations.
type Service struct {
	Store store
}

// New creates a new exercises service.
func New(store store) *Service {
	return &Service{Store: store}
}

// List returns the exercise catalog for a user.
func (s *Service) List(ctx context.Context, userID string) ([]db.Exercise, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, service.NewError(service.ErrorValidation, "userId is required")
	}
	exercises, err := s.Store.ListExercises(ctx, userID)
	if err != nil {
		return nil, service.NewError(service.ErrorInternal, err.Error())
	}
	return exercises, nil
}

// Create adds a new exercise to the catalog.
func (s *Service) Create(ctx context.Context, userID, name string, isCore bool) (*db.Exercise, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, service.NewError(service.ErrorValidation, "userId is required")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, service.NewError(service.ErrorValidation, "exercise name is required")
	}
	user, err := s.Store.GetUser(ctx, userID)
	if err != nil || user == nil {
		return nil, service.NewError(service.ErrorNotFound, "user not found")
	}
	if isCore && !user.IsAdmin {
		return nil, service.NewError(service.ErrorForbidden, "core exercise requires admin privileges")
	}
	exercise, err := s.Store.CreateExercise(ctx, name, userID, isCore)
	if err != nil {
		return nil, service.NewError(service.ErrorValidation, err.Error())
	}
	return exercise, nil
}

// Update renames an exercise or creates a personal copy.
func (s *Service) Update(ctx context.Context, userID, exerciseID, name string) (*db.Exercise, error) {
	userID = strings.TrimSpace(userID)
	exerciseID = strings.TrimSpace(exerciseID)
	name = strings.TrimSpace(name)
	if userID == "" {
		return nil, service.NewError(service.ErrorValidation, "userId is required")
	}
	if exerciseID == "" {
		return nil, service.NewError(service.ErrorValidation, "exercise id is required")
	}
	if name == "" {
		return nil, service.NewError(service.ErrorValidation, "exercise name is required")
	}
	user, err := s.Store.GetUser(ctx, userID)
	if err != nil || user == nil {
		return nil, service.NewError(service.ErrorNotFound, "user not found")
	}
	exercise, err := s.Store.GetExercise(ctx, exerciseID)
	if err != nil {
		return nil, service.NewError(service.ErrorNotFound, "exercise not found")
	}
	if exercise.IsCore && !user.IsAdmin {
		copied, err := s.Store.CreateExercise(ctx, name, userID, false)
		if err != nil {
			return nil, service.NewError(service.ErrorValidation, err.Error())
		}
		if err := s.Store.ReplaceExerciseForUser(ctx, userID, exercise.ID, copied.ID, copied.Name); err != nil {
			return nil, service.NewError(service.ErrorInternal, err.Error())
		}
		return copied, nil
	}
	if !user.IsAdmin && exercise.OwnerUserID != "" && exercise.OwnerUserID != userID {
		return nil, service.NewError(service.ErrorForbidden, "exercise not owned by user")
	}
	updated, err := s.Store.RenameExercise(ctx, exerciseID, name)
	if err != nil {
		return nil, service.NewError(service.ErrorValidation, err.Error())
	}
	return updated, nil
}

// Delete removes an exercise from the catalog.
func (s *Service) Delete(ctx context.Context, userID, exerciseID string) error {
	userID = strings.TrimSpace(userID)
	exerciseID = strings.TrimSpace(exerciseID)
	if userID == "" {
		return service.NewError(service.ErrorValidation, "userId is required")
	}
	if exerciseID == "" {
		return service.NewError(service.ErrorValidation, "exercise id is required")
	}
	user, err := s.Store.GetUser(ctx, userID)
	if err != nil || user == nil {
		return service.NewError(service.ErrorNotFound, "user not found")
	}
	exercise, err := s.Store.GetExercise(ctx, exerciseID)
	if err != nil {
		return service.NewError(service.ErrorNotFound, "exercise not found")
	}
	if exercise.IsCore && !user.IsAdmin {
		return service.NewError(service.ErrorForbidden, "cannot delete core exercise")
	}
	if !user.IsAdmin && exercise.OwnerUserID != "" && exercise.OwnerUserID != userID {
		return service.NewError(service.ErrorForbidden, "exercise not owned by user")
	}
	if err := s.Store.DeleteExercise(ctx, exerciseID); err != nil {
		return service.NewError(service.ErrorValidation, err.Error())
	}
	return nil
}

// Backfill rebuilds core exercises from workout data.
func (s *Service) Backfill(ctx context.Context) error {
	if err := s.Store.BackfillCoreExercises(ctx); err != nil {
		return service.NewError(service.ErrorInternal, err.Error())
	}
	return nil
}
