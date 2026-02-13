package db

import "errors"

// ErrWorkoutNotFound indicates that the referenced workout does not exist.
var ErrWorkoutNotFound = errors.New("workout not found")
