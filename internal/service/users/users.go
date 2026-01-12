package users

import (
	domain "github.com/gi8lino/motus/internal/domain/users"
	servicepkg "github.com/gi8lino/motus/internal/service"
)

type Store = domain.Store

// Service coordinates user operations.
type Service struct {
	manager *domain.Manager
}

// New creates a new users service.
func New(store Store, authHeader string, allowRegistration bool) *Service {
	return &Service{manager: domain.NewManager(store, authHeader, allowRegistration)}
}

// mapError translates domain errors into service errors for templates.
func (s *Service) mapError(err error) error {
	return servicepkg.MapDomainError(err, func(kind int) (servicepkg.ErrorKind, bool) {
		switch domain.ErrorKind(kind) {
		case domain.KindValidation:
			return servicepkg.ErrorValidation, true
		case domain.KindForbidden:
			return servicepkg.ErrorForbidden, true
		case domain.KindUnauthorized:
			return servicepkg.ErrorUnauthorized, true
		case domain.KindInternal:
			return servicepkg.ErrorInternal, true
		default:
			return servicepkg.ErrorInternal, false
		}
	})
}
