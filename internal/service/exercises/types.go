// Package exercises provides service access to exercise catalog operations.
package exercises

import domain "github.com/gi8lino/motus/internal/domain/exercises"

// Store exposes persistence required by the exercises service.
type Store = domain.Store

// Exercise is the domain-level DTO for catalog exercises.
type Exercise = domain.Exercise

// Service coordinates exercise catalog operations.
type Service struct {
	manager *domain.Manager
}

// New creates a new exercises service.
func New(store Store) *Service {
	return &Service{manager: domain.NewManager(store)}
}
