package exercises

import (
	"context"

	"github.com/gi8lino/motus/internal/db"
)

// Create adds a new exercise entry to the catalog.
func (m *Manager) Create(ctx context.Context, userID, name string, isCore bool) (*db.Exercise, error) {
	uid, err := requireUserID(userID)
	if err != nil {
		return nil, err
	}
	nm, err := requireName(name)
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
		return nil, forbidden("core exercise requires admin privileges")
	}

	exercise, err := m.store.CreateExercise(ctx, nm, uid, isCore)
	if err != nil {
		return nil, internal(err)
	}
	return exercise, nil
}

// Delete removes an exercise from the catalog.
func (m *Manager) Delete(ctx context.Context, userID, exerciseID string) error {
	userID, err := requireUserID(userID)
	if err != nil {
		return err
	}
	exerciseID, err = requireEntityID(exerciseID, "exercise id is required")
	if err != nil {
		return err
	}

	user, err := m.store.GetUser(ctx, userID)
	if err != nil {
		return internal(err)
	}
	if user == nil {
		return notFound("user not found")
	}
	if !user.IsAdmin {
		return forbidden("admin privileges required")
	}

	exercise, err := m.store.GetExercise(ctx, exerciseID)
	if err != nil {
		return internal(err)
	}
	if exercise == nil {
		return notFound("exercise not found")
	}

	if err := m.store.DeleteExercise(ctx, exerciseID); err != nil {
		return internal(err)
	}
	return nil
}

// Update renames an exercise entry.
func (m *Manager) Update(ctx context.Context, userID, exerciseID, name string) (*db.Exercise, error) {
	userID, err := requireUserID(userID)
	if err != nil {
		return nil, err
	}
	exerciseID, err = requireEntityID(exerciseID, "exercise id is required")
	if err != nil {
		return nil, err
	}
	name, err = requireName(name)
	if err != nil {
		return nil, err
	}

	user, err := m.store.GetUser(ctx, userID)
	if err != nil {
		return nil, internal(err)
	}
	if user == nil {
		return nil, notFound("user not found")
	}
	if !user.IsAdmin {
		return nil, forbidden("admin privileges required")
	}

	exercise, err := m.store.GetExercise(ctx, exerciseID)
	if err != nil {
		return nil, internal(err)
	}
	if exercise == nil {
		return nil, notFound("exercise not found")
	}

	updated, err := m.store.RenameExercise(ctx, exerciseID, name)
	if err != nil {
		return nil, internal(err)
	}
	return updated, nil
}
