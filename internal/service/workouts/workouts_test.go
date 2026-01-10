package workouts

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gi8lino/motus/internal/db"
	"github.com/gi8lino/motus/internal/utils"
)

func TestNormalizeSteps(t *testing.T) {
	t.Parallel()

	t.Run("Normalizes steps", func(t *testing.T) {
		t.Parallel()

		inputs := []StepInput{
			{
				Type: " set ",
				Name: "  Squats ",
				Subsets: []SubsetInput{
					{
						Name:     "Main",
						Duration: "15s",
						SoundKey: "beep",
						Exercises: []ExerciseInput{
							{Name: "Squat", Reps: "10", Weight: "", Type: "rep"},
						},
					},
				},
			},
			{
				Type:         "pause",
				Name:         "Rest",
				Duration:     "30s",
				PauseOptions: db.PauseOptions{AutoAdvance: true},
			},
			{
				Type: "set",
				Name: "Warmup",
				Subsets: []SubsetInput{
					{
						Name:     "Jog",
						Duration: "20s",
						Exercises: []ExerciseInput{
							{Name: "Jog", Duration: "20s", Type: "timed"},
						},
					},
				},
			},
		}

		steps, err := NormalizeSteps(inputs, func(key string) bool {
			return key == "beep" || key == ""
		})
		require.NoError(t, err)
		require.Len(t, steps, 3)

		setStep := steps[0]
		require.Len(t, setStep.Subsets, 1)
		assert.Equal(t, string(utils.StepTypeSet), setStep.Type)
		assert.Equal(t, "Squats", setStep.Name)
		sub := setStep.Subsets[0]
		assert.Equal(t, "Main", sub.Name)
		require.Len(t, sub.Exercises, 1)
		assert.Equal(t, "Squat", sub.Exercises[0].Name)
		assert.Equal(t, "10", sub.Exercises[0].Reps)

		pauseStep := steps[1]
		assert.Equal(t, string(utils.StepTypePause), pauseStep.Type)
		assert.True(t, pauseStep.PauseOptions.AutoAdvance)
		assert.Empty(t, pauseStep.Subsets)

		timedStep := steps[2]
		assert.Equal(t, string(utils.StepTypeSet), timedStep.Type)
		require.Len(t, timedStep.Subsets, 1)
		assert.Equal(t, "Jog", timedStep.Subsets[0].Name)
		require.Len(t, timedStep.Subsets[0].Exercises, 1)
		assert.Equal(t, "stopwatch", timedStep.Subsets[0].Exercises[0].Type)
	})

	t.Run("Rejects invalid sound", func(t *testing.T) {
		t.Parallel()

		_, err := NormalizeSteps([]StepInput{{
			Type:     "set",
			Name:     "Lift",
			SoundKey: "nope",
			Subsets: []SubsetInput{
				{
					Name: "Lift",
					Exercises: []ExerciseInput{
						{Name: "Lift", Reps: "5"},
					},
				},
			},
		}}, func(key string) bool {
			return key == "beep"
		})
		require.Error(t, err)
	})

	t.Run("Rejects invalid duration", func(t *testing.T) {
		t.Parallel()

		_, err := NormalizeSteps([]StepInput{{
			Type:     "set",
			Name:     "Lift",
			Duration: "nope",
			Subsets: []SubsetInput{
				{
					Name: "Lift",
					Exercises: []ExerciseInput{
						{Name: "Lift", Reps: "5"},
					},
				},
			},
		}}, func(key string) bool { return true })
		require.Error(t, err)
	})
}
