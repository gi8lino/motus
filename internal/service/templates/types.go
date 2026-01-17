// Package templates provides domain logic for workout templates.
package templates

import "github.com/gi8lino/motus/internal/db"

// Workout is the domain-level DTO for templates.
type Workout = db.Workout

// errorScope is the service error scope for templates.
const errorScope = "templates"
