package workouts

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeRepeatRest(t *testing.T) {
	t.Parallel()

	t.Run("Defaults", func(t *testing.T) {
		t.Parallel()
		seconds, auto, after, _ := normalizeRepeatRest(3, 10, true, false, "ok")
		assert.Equal(t, 10, seconds)
		assert.True(t, auto)
		assert.False(t, after)
	})

	t.Run("Auto-advance", func(t *testing.T) {
		t.Parallel()
		seconds, auto, after, _ := normalizeRepeatRest(1, 10, true, false, "ok")
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
