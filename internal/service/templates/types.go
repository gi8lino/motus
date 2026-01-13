// Package templates provides service access to shared workout templates.
package templates

import domain "github.com/gi8lino/motus/internal/domain/templates"

// Store exposes persistence required by the templates service.
type Store = domain.Store

// Workout is the domain-level DTO for templates.
type Workout = domain.Workout

// Service orchestrates template use cases.
type Service struct {
	manager *domain.Manager
}

// New builds a template service.
func New(store Store) *Service {
	return &Service{manager: domain.NewManager(store)}
}
