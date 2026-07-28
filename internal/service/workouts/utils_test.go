package workouts

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gi8lino/motus/internal/utils"
)

func validSound(key string) bool {
	return key == "ok"
}

func TestNormalizeRepeatRest(t *testing.T) {
	t.Parallel()

	t.Run("Defaults", func(t *testing.T) {
		t.Parallel()
		seconds, auto, after, _, _ := normalizeRepeatRest(3, 10, true, false, "ok", "Rest")
		assert.Equal(t, 10, seconds)
		assert.True(t, auto)
		assert.False(t, after)
	})

	t.Run("Auto-advance", func(t *testing.T) {
		t.Parallel()
		seconds, auto, after, _, _ := normalizeRepeatRest(1, 10, true, false, "ok", "Rest")
		assert.Equal(t, 0, seconds)
		assert.False(t, auto)
		assert.False(t, after)
	})
}

func TestParseDurationField(t *testing.T) {
	t.Parallel()

	t.Run("parses seconds", func(t *testing.T) {
		seconds, err := parseDurationField("5s", 0)
		assert.NoError(t, err)
		assert.Equal(t, 5, seconds)
	})

	t.Run("parses minutes", func(t *testing.T) {
		t.Parallel()
		seconds, err := parseDurationField("", 7)
		assert.NoError(t, err)
		assert.Equal(t, 7, seconds)
	})

	t.Run("parses hours", func(t *testing.T) {
		t.Parallel()
		seconds, err := parseDurationField("-1s", 0)
		assert.NoError(t, err)
		assert.Equal(t, 0, seconds)
	})
}

func TestIsEmptyRepExercise(t *testing.T) {
	t.Parallel()

	t.Run("returns true when empty", func(t *testing.T) {
		t.Parallel()
		assert.True(t, isEmptyRepExercise(ExerciseInput{}))
	})

	t.Run("returns false when not empty", func(t *testing.T) {
		t.Parallel()
		assert.False(t, isEmptyRepExercise(ExerciseInput{Name: "Push"}))
	})
}

func TestNormalizeSteps(t *testing.T) {
	t.Parallel()

	t.Run("Pause and set", func(t *testing.T) {
		t.Parallel()
		steps, err := NormalizeSteps([]StepInput{
			{Type: utils.StepTypePause.String(), Name: "Rest", Duration: "5s"},
			{Type: utils.StepTypeSet.String(), Name: "Set", Subsets: []SubsetInput{
				{Name: "First", SoundKey: "ok", Exercises: []ExerciseInput{
					{Name: "Push", Type: utils.ExerciseTypeRep},
				}},
			}},
		}, validSound)
		require.NoError(t, err)
		assert.Len(t, steps, 2)
		assert.Equal(t, utils.StepTypePause.String(), steps[0].Type)
		assert.Equal(t, "First", steps[1].Subsets[0].Name)
	})

	t.Run("No subset", func(t *testing.T) {
		t.Parallel()

		_, err := NormalizeSteps([]StepInput{
			{Type: utils.StepTypeSet.String(), Name: "Set"},
		}, validSound)
		require.Error(t, err)
	})

	t.Run("Invalid sound", func(t *testing.T) {
		t.Parallel()

		_, err := NormalizeSteps([]StepInput{
			{Type: utils.StepTypePause.String(), Name: "Rest", SoundKey: "bad"},
		}, validSound)
		require.Error(t, err)
	})
}

func TestNormalizeSubsetExercises(t *testing.T) {
	t.Parallel()

	t.Run("Valid reps", func(t *testing.T) {
		t.Parallel()
		_, err := normalizeSubsetExercises("test", []ExerciseInput{
			{Name: "Rep", Type: utils.ExerciseTypeRep, Reps: "10-12"},
		}, validSound)
		assert.NoError(t, err)
	})

	t.Run("Invalid reps", func(t *testing.T) {
		t.Parallel()
		_, err := normalizeSubsetExercises("test", []ExerciseInput{
			{Name: "Rep", Type: utils.ExerciseTypeRep, Reps: "x"},
		}, validSound)
		assert.Error(t, err)
	})
}

func TestNormalizeSubsets(t *testing.T) {
	t.Parallel()

	t.Run("No exercises", func(t *testing.T) {
		t.Parallel()
		_, err := normalizeSubsets("Set", []SubsetInput{{
			Name: "A",
		}}, validSound)
		assert.Error(t, err)
	})

	t.Run("Valid", func(t *testing.T) {
		t.Parallel()
		_, err := normalizeSubsets("Set", []SubsetInput{{
			Name: "A",
			Exercises: []ExerciseInput{
				{Name: "Push", Type: utils.ExerciseTypeRep},
			},
		}}, validSound)
		assert.NoError(t, err)
	})

	t.Run("Superset rejects timed exercises", func(t *testing.T) {
		t.Parallel()

		_, err := normalizeSubsets("Set", []SubsetInput{{
			Name:     "Circuit",
			Superset: true,
			Exercises: []ExerciseInput{
				{Name: "Pull-ups", Type: utils.ExerciseTypeRep, Reps: "10"},
				{Name: "Mountain climbers", Type: utils.ExerciseTypeStopwatch, Duration: "20s"},
			},
		}}, validSound)

		require.EqualError(
			t,
			err,
			"superset Circuit supports rep exercises only; use a regular subset for timed exercises",
		)
	})

	t.Run("Superset uses only its block sound", func(t *testing.T) {
		t.Parallel()

		subsets, err := normalizeSubsets("Set", []SubsetInput{{
			Name:     "Circuit",
			SoundKey: "ok",
			Superset: true,
			Exercises: []ExerciseInput{
				{Name: "Pull-ups", Type: utils.ExerciseTypeRep, Reps: "10", SoundKey: "ok"},
			},
		}}, validSound)

		require.NoError(t, err)
		require.Len(t, subsets, 1)
		require.Len(t, subsets[0].Exercises, 1)
		assert.Equal(t, "ok", subsets[0].SoundKey)
		assert.Empty(t, subsets[0].Exercises[0].SoundKey)
	})
}
