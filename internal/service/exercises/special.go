package exercises

import "context"

// Backfill rebuilds core exercises from workout data.
func (s *Service) Backfill(ctx context.Context) error {
	if err := s.manager.Backfill(ctx); err != nil {
		return s.mapError(err)
	}
	return nil
}
