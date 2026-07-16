package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExerciseMetadataNormalization(t *testing.T) {
	t.Parallel()

	assert.Equal(t, []string{"swing", "kettlebell"}, normalizedNames([]string{" Swing ", "KETTLEBELL"}))
	assert.Equal(t, []string{"kettlebell", "hinge"}, normalizedLabels([]string{" Kettlebell ", "hinge", "KETTLEBELL", ""}))
}
