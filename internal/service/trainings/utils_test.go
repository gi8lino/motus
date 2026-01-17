package trainings

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gi8lino/motus/internal/utils"
)

func TestMapExercise(t *testing.T) {
	t.Parallel()

	t.Run("Defaults", func(t *testing.T) {
		t.Parallel()

		input := SubsetExercise{
			Name:     "Push",
			Type:     utils.ExerciseTypeRep,
			Reps:     "10",
			Weight:   "10kg",
			Duration: "10s",
			SoundKey: "beep",
		}
		result := mapExercise(input)
		assert.Equal(t, "Push", result.Name)
		assert.Equal(t, utils.ExerciseTypeRep, result.Type)
		assert.Equal(t, "10", result.Reps)
		assert.Equal(t, "10kg", result.Weight)
		assert.Equal(t, "10s", result.Duration)
		assert.Equal(t, "beep", result.SoundKey)
	})
}

func TestDeriveExerciseDuration(t *testing.T) {
	t.Parallel()

	t.Run("Defaults", func(t *testing.T) {
		t.Parallel()
		subset := WorkoutSubset{EstimatedSeconds: 30}
		countdown := SubsetExercise{Type: utils.ExerciseTypeCountdown, Duration: "15s"}
		stopwatch := SubsetExercise{Type: utils.ExerciseTypeStopwatch}
		rep := SubsetExercise{Type: utils.ExerciseTypeRep}

		seconds, auto := deriveExerciseDuration(countdown, subset)
		assert.Equal(t, 15, seconds)
		assert.True(t, auto)

		seconds, auto = deriveExerciseDuration(stopwatch, subset)
		assert.Equal(t, 30, seconds)
		assert.False(t, auto)

		subset.Exercises = []SubsetExercise{rep}
		seconds, auto = deriveExerciseDuration(rep, subset)
		assert.Equal(t, 30, seconds)
		assert.False(t, auto)
	})
}

func TestParseDurationSeconds(t *testing.T) {
	t.Parallel()

	t.Run("Defaults", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, 5, parseDurationSeconds("5s"))
		assert.Equal(t, 0, parseDurationSeconds("-1s"))
		assert.Equal(t, 0, parseDurationSeconds("invalid"))
		assert.Equal(t, 0, parseDurationSeconds(""))
	})
}
