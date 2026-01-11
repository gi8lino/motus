package workouts

import (
	domain "github.com/gi8lino/motus/internal/domain/workouts"
	servicepkg "github.com/gi8lino/motus/internal/service"
)

// Store exposes the persistence requirements for workout services.
type Store = domain.Store

// WorkoutRequest captures the payload for creating or updating workouts.
type WorkoutRequest = domain.WorkoutRequest

// Service coordinates workout operations.
type Service struct {
	manager *domain.Manager
}

// New creates a new workouts service.
func New(store Store) *Service {
	return &Service{manager: domain.NewManager(store)}
}

// mapError translates domain errors into service errors for workouts.
func (s *Service) mapError(err error) error {
	return servicepkg.MapDomainError(err, func(kind int) (servicepkg.ErrorKind, bool) {
		switch domain.ErrorKind(kind) {
		case domain.KindValidation:
			return servicepkg.ErrorValidation, true
		case domain.KindNotFound:
			return servicepkg.ErrorNotFound, true
		case domain.KindInternal:
			return servicepkg.ErrorInternal, true
		default:
			return servicepkg.ErrorInternal, false
		}
	})
}
