// Package catalog provides Motus's built-in core exercise catalog.
package catalog

import _ "embed"

// CoreExercisesYAML is the default catalog shipped inside every Motus binary.
//
//go:embed core-exercises.yaml
var CoreExercisesYAML []byte
