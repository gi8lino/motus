package exercises

import (
	"context"
	"fmt"
)

// Manager orchestrates exercise catalog rules.
type Manager struct {
	store Store
}

// NewManager builds a Manager.
func NewManager(store Store) *Manager {
	return &Manager{store: store}
}

// Create adds a new exercise entry to the catalog.
func (m *Manager) Create(ctx context.Context, userID, name string, isCore bool) (*Exercise, error) {
	uid, err := requireUserID(userID)
	if err != nil {
		return nil, err
	}
	clean, err := requireName(name)
	if err != nil {
		return nil, err
	}

	user, err := m.store.GetUser(ctx, uid)
	if err != nil {
		return nil, internal(err)
	}
	if user == nil {
		return nil, notFound("user not found")
	}
	if isCore && !user.IsAdmin {
		return nil, forbidden("core exercises require admin permissions")
	}

	exercise, err := m.store.CreateExercise(ctx, clean, uid, isCore)
	if err != nil {
		return nil, internal(err)
	}
	return exercise, nil
}

// Update renames an exercise entry.
func (m *Manager) Update(ctx context.Context, userID, exerciseID, name string) (*Exercise, error) {
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

	user, err := m.store.GetUser(ctx, uid)
	if err != nil {
		return nil, internal(err)
	}
	if user == nil {
		return nil, notFound("user not found")
	}

	exercise, err := m.store.GetExercise(ctx, id)
	if err != nil {
		return nil, internal(err)
	}
	if exercise == nil {
		return nil, notFound("exercise not found")
	}

	if exercise.IsCore && !user.IsAdmin {
		return nil, forbidden("core exercises require admin permissions")
	}

	if exercise.OwnerUserID != "" && exercise.OwnerUserID != uid && !user.IsAdmin {
		return nil, forbidden("exercise belongs to another user")
	}

	updated, err := m.store.RenameExercise(ctx, id, clean)
	if err != nil {
		return nil, internal(err)
	}
	return updated, nil
}

// Delete removes an exercise entry.
func (m *Manager) Delete(ctx context.Context, userID, exerciseID string) error {
	uid, err := requireUserID(userID)
	if err != nil {
		return err
	}
	id, err := requireEntityID(exerciseID, "exercise id is required")
	if err != nil {
		return err
	}

	user, err := m.store.GetUser(ctx, uid)
	if err != nil {
		return internal(err)
	}
	if user == nil {
		return notFound("user not found")
	}

	exercise, err := m.store.GetExercise(ctx, id)
	if err != nil {
		return internal(err)
	}
	if exercise == nil {
		return notFound("exercise not found")
	}

	if exercise.IsCore && !user.IsAdmin {
		return forbidden("core exercises require admin permissions")
	}

	if exercise.OwnerUserID != "" && exercise.OwnerUserID != uid && !user.IsAdmin {
		return forbidden("exercise belongs to another user")
	}

	if err := m.store.DeleteExercise(ctx, id); err != nil {
		return internal(err)
	}
	return nil
}

// List returns the exercise catalog for a user.
func (m *Manager) List(ctx context.Context, userID string) ([]Exercise, error) {
	uid, err := requireUserID(userID)
	if err != nil {
		return nil, err
	}
	exercises, err := m.store.ListExercises(ctx, uid)
	if err != nil {
		return nil, internal(err)
	}
	return exercises, nil
}

// Backfill rebuilds core exercises from workout data.
func (m *Manager) Backfill(ctx context.Context) error {
	if err := m.store.BackfillCoreExercises(ctx); err != nil {
		return internal(fmt.Errorf("backfill core exercises: %w", err))
	}
	return nil
}
