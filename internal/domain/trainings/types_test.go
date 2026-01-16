package trainings

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrainingStepState(t *testing.T) {
	t.Parallel()

	t.Run("ZeroValues", func(t *testing.T) {
		t.Parallel()
		step := TrainingStepState{}
		assert.False(t, step.Running)
		assert.False(t, step.Completed)
		assert.Zero(t, step.EstimatedSeconds)
		assert.Nil(t, step.Exercises)
	})
}
