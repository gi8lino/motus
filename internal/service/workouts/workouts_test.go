package workouts

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gi8lino/motus/internal/db"
	domainworkouts "github.com/gi8lino/motus/internal/domain/workouts"
	"github.com/gi8lino/motus/internal/utils"
)

func TestNormalizeSteps(t *testing.T) {
	t.Parallel()

	t.Run("Normalizes steps", func(t *testing.T) {
		t.Parallel()

		inputs := []domainworkouts.StepInput{
			{
				Type: " set ",
				Name: "  Squats ",
				Subsets: []domainworkouts.SubsetInput{
					{
						Name:     "Main",
						Duration: "15s",
						SoundKey: "beep",
						Exercises: []domainworkouts.ExerciseInput{
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
				Subsets: []domainworkouts.SubsetInput{
					{
						Name:     "Jog",
						Duration: "20s",
						Exercises: []domainworkouts.ExerciseInput{
							{Name: "Jog", Duration: "20s", Type: "stopwatch"},
						},
					},
				},
			},
		}

		steps, err := domainworkouts.NormalizeSteps(inputs, func(key string) bool {
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

		warmupStep := steps[2]
		assert.Equal(t, string(utils.StepTypeSet), warmupStep.Type)
		require.Len(t, warmupStep.Subsets, 1)
		assert.Equal(t, "Jog", warmupStep.Subsets[0].Name)
		require.Len(t, warmupStep.Subsets[0].Exercises, 1)
		assert.Equal(t, "stopwatch", warmupStep.Subsets[0].Exercises[0].Type)
	})

	t.Run("Rejects invalid sound", func(t *testing.T) {
		t.Parallel()

		_, err := domainworkouts.NormalizeSteps([]domainworkouts.StepInput{{
			Type:     "set",
			Name:     "Lift",
			SoundKey: "nope",
			Subsets: []domainworkouts.SubsetInput{
				{
					Name: "Lift",
					Exercises: []domainworkouts.ExerciseInput{
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

		_, err := domainworkouts.NormalizeSteps([]domainworkouts.StepInput{{
			Type:     "set",
			Name:     "Lift",
			Duration: "nope",
			Subsets: []domainworkouts.SubsetInput{
				{
					Name: "Lift",
					Exercises: []domainworkouts.ExerciseInput{
						{Name: "Lift", Reps: "5"},
					},
				},
			},
		}}, func(key string) bool { return true })
		require.Error(t, err)
	})
}
