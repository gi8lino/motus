package exercises

import (
	"context"
	"strings"

	"github.com/gi8lino/motus/internal/db"
	"github.com/gi8lino/motus/internal/service"
)

// Store defines the persistence methods needed by the exercises service.
type Store interface {
	// ListExercises loads core plus user-specific exercises for a user.
	ListExercises(ctx context.Context, userID string) ([]db.Exercise, error)
	// GetUser fetches a user by id for permission checks.
	GetUser(ctx context.Context, userID string) (*db.User, error)
	// CreateExercise inserts a catalog entry for a user.
	CreateExercise(ctx context.Context, name, userID string, isCore bool) (*db.Exercise, error)
	// GetExercise returns a single exercise entry.
	GetExercise(ctx context.Context, id string) (*db.Exercise, error)
	// ReplaceExerciseForUser replaces a core exercise with a user copy.
	ReplaceExerciseForUser(ctx context.Context, userID, oldID, newID, newName string) error
	// RenameExercise updates the name of an exercise.
	RenameExercise(ctx context.Context, id, name string) (*db.Exercise, error)
	// DeleteExercise removes an exercise entry.
	DeleteExercise(ctx context.Context, id string) error
	// BackfillCoreExercises migrates workout exercises into the core catalog.
	BackfillCoreExercises(ctx context.Context) error
}

// Service coordinates exercise catalog operations.
type Service struct {
	Store Store
}

// New creates a new exercises service.
func New(store Store) *Service {
	return &Service{Store: store}
}

// List returns the exercise catalog for a user.
func (s *Service) List(ctx context.Context, userID string) ([]db.Exercise, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, service.NewError(service.ErrorValidation, "userId is required")
	}

	// Fetch both core and personal exercises for the user.
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

	// Resolve the user to validate permissions.
	user, err := s.Store.GetUser(ctx, userID)
	if err != nil || user == nil {
		return nil, service.NewError(service.ErrorNotFound, "user not found")
	}

	// Core exercises may only be created by admins.
	if isCore && !user.IsAdmin {
		return nil, service.NewError(service.ErrorForbidden, "core exercise requires admin privileges")
	}

	// Insert the catalog entry for this user.
	exercise, err := s.Store.CreateExercise(ctx, name, userID, isCore)
	if err != nil {
		return nil, service.NewError(service.ErrorValidation, err.Error())
	}

	return exercise, nil
}

// Update renames an exercise or creates a personal copy.
func (s *Service) Update(ctx context.Context, userID, exerciseID, name string) (*db.Exercise, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, service.NewError(service.ErrorValidation, "userId is required")
	}

	exerciseID = strings.TrimSpace(exerciseID)
	if exerciseID == "" {
		return nil, service.NewError(service.ErrorValidation, "exercise id is required")
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return nil, service.NewError(service.ErrorValidation, "exercise name is required")
	}

	// Resolve the user to validate permissions.
	user, err := s.Store.GetUser(ctx, userID)
	if err != nil || user == nil {
		return nil, service.NewError(service.ErrorNotFound, "user not found")
	}

	// Only admins can rename exercises (core or personal).
	if !user.IsAdmin {
		return nil, service.NewError(service.ErrorForbidden, "admin privileges required")
	}

	// Ensure the exercise exists before renaming it.
	if _, err := s.Store.GetExercise(ctx, exerciseID); err != nil {
		return nil, service.NewError(service.ErrorNotFound, "exercise not found")
	}

	// Persist the new name in the catalog.
	updated, err := s.Store.RenameExercise(ctx, exerciseID, name)
	if err != nil {
		return nil, service.NewError(service.ErrorValidation, err.Error())
	}

	return updated, nil
}

// Delete removes an exercise from the catalog.
func (s *Service) Delete(ctx context.Context, userID, exerciseID string) error {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return service.NewError(service.ErrorValidation, "userId is required")
	}

	exerciseID = strings.TrimSpace(exerciseID)
	if exerciseID == "" {
		return service.NewError(service.ErrorValidation, "exercise id is required")
	}

	user, err := s.Store.GetUser(ctx, userID)
	if err != nil || user == nil {
		return service.NewError(service.ErrorNotFound, "user not found")
	}

	if !user.IsAdmin {
		return service.NewError(service.ErrorForbidden, "admin privileges required")
	}

	// Ensure the exercise exists before attempting deletion.
	if _, err := s.Store.GetExercise(ctx, exerciseID); err != nil {
		return service.NewError(service.ErrorNotFound, "exercise not found")
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
