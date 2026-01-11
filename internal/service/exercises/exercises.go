package exercises

import (
	domain "github.com/gi8lino/motus/internal/domain/exercises"
	servicepkg "github.com/gi8lino/motus/internal/service"
)

type Store = domain.Store

// Service coordinates exercise catalog operations.
type Service struct {
	manager *domain.Manager
}

// New creates a new exercises service.
func New(store Store) *Service {
	return &Service{manager: domain.NewManager(store)}
}

// mapError translates domain errors into service errors for templates.
func (s *Service) mapError(err error) error {
	return servicepkg.MapDomainError(err, func(kind int) (servicepkg.ErrorKind, bool) {
		switch domain.ErrorKind(kind) {
		case domain.KindValidation:
			return servicepkg.ErrorValidation, true
		case domain.KindForbidden:
			return servicepkg.ErrorForbidden, true
		case domain.KindNotFound:
			return servicepkg.ErrorNotFound, true
		case domain.KindInternal:
			return servicepkg.ErrorInternal, true
		default:
			return servicepkg.ErrorInternal, false
		}
	})
}
