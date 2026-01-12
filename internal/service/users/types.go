// Package users provides service access to user operations.
package users

import domain "github.com/gi8lino/motus/internal/domain/users"

// Store exposes persistence required by the users service.
type Store = domain.Store

// User is the domain-level DTO for users.
type User = domain.User

// Service coordinates user operations.
type Service struct {
	manager *domain.Manager
}

// New creates a new users service.
func New(store Store, authHeader string, allowRegistration bool) *Service {
	return &Service{manager: domain.NewManager(store, authHeader, allowRegistration)}
}
