// Package exercises defines domain logic for exercise catalog rules.
package exercises

import "github.com/gi8lino/motus/internal/db"

// Exercise is the domain-level DTO for catalog exercises.
type Exercise = db.Exercise

// User is the domain-level DTO for users used by the exercise domain.
type User = db.User
