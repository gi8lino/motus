package workouts

import (
	"testing"

	"github.com/gi8lino/motus/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validSound(key string) bool {
	return key == "ok"
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
		// ensure at least one exercise
		_, err := normalizeSubsets("Set", []SubsetInput{{
			Name: "A",
			Exercises: []ExerciseInput{
				{Name: "Push", Type: utils.ExerciseTypeRep},
			},
		}}, validSound)
		assert.NoError(t, err)
	})
}
