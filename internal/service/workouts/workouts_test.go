package workouts

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gi8lino/motus/internal/db"
)

func TestNormalizeSteps(t *testing.T) {
	t.Parallel()

	t.Run("Normalizes steps", func(t *testing.T) {
		t.Parallel()

		inputs := []StepInput{
			{
				Type:     " set ",
				Name:     "  Squats ",
				Duration: "15s",
				SoundKey: "beep",
				Exercises: []ExerciseInput{
					{Name: "Squat", Amount: "10", Weight: ""},
				},
			},
			{
				Type:         "pause",
				Name:         "Rest",
				Duration:     "30s",
				PauseOptions: db.PauseOptions{AutoAdvance: true},
				Weight:       "",
			},
			{
				Type:             "timed",
				Name:             "Warmup",
				Duration:         "20s",
				EstimatedSeconds: 99,
			},
		}

		steps, err := NormalizeSteps(inputs, func(key string) bool {
			return key == "beep" || key == ""
		})
		require.NoError(t, err)
		require.Len(t, steps, 3)

		setStep := steps[0]
		assert.Equal(t, "set", setStep.Type)
		assert.Equal(t, "Squats", setStep.Name)
		assert.Equal(t, "Squat", setStep.Exercise)
		assert.Equal(t, "10", setStep.Amount)

		pauseStep := steps[1]
		assert.Equal(t, "pause", pauseStep.Type)
		assert.True(t, pauseStep.PauseOptions.AutoAdvance)
		assert.Equal(t, "__auto__", pauseStep.Weight)
		assert.Empty(t, pauseStep.Exercise)
		assert.Empty(t, pauseStep.Amount)
		assert.Empty(t, pauseStep.Exercises)

		timedStep := steps[2]
		assert.Equal(t, "timed", timedStep.Type)
		assert.Equal(t, 0, timedStep.EstimatedSeconds)
	})

	t.Run("Rejects invalid sound", func(t *testing.T) {
		t.Parallel()

		_, err := NormalizeSteps([]StepInput{{Type: "set", Name: "Lift", SoundKey: "nope"}}, func(key string) bool {
			return key == "beep"
		})
		require.Error(t, err)
	})

	t.Run("Rejects invalid duration", func(t *testing.T) {
		t.Parallel()

		_, err := NormalizeSteps([]StepInput{{Type: "set", Name: "Lift", Duration: "nope"}}, func(key string) bool { return true })
		require.Error(t, err)
	})
}
