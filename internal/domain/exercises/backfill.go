package exercises

import "context"

// Backfill reloads core exercises from workout data.
func (m *Manager) Backfill(ctx context.Context) error {
	if err := m.store.BackfillCoreExercises(ctx); err != nil {
		return internal(err)
	}
	return nil
}
