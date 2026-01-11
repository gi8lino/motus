package templates

import (
	domain "github.com/gi8lino/motus/internal/domain/templates"
	servicepkg "github.com/gi8lino/motus/internal/service"
)

type Store = domain.Store

// Service orchestrates template use cases.
type Service struct {
	manager *domain.Manager
}

// New builds a template service.
func New(store Store) *Service {
	return &Service{manager: domain.NewManager(store)}
}

// mapError translates domain errors into service errors for templates.
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
