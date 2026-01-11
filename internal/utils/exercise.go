package utils

// Exercise type constants.
const (
	ExerciseTypeRep       = "rep"
	ExerciseTypeStopwatch = "stopwatch"
	ExerciseTypeCountdown = "countdown"
)

// NormalizeExerciseType converts the provided value to a canonical exercise type.
func NormalizeExerciseType(value string) string {
	switch NormalizeToken(value) {
	case ExerciseTypeCountdown:
		return ExerciseTypeCountdown
	case ExerciseTypeStopwatch:
		return ExerciseTypeStopwatch
	default:
		return ExerciseTypeRep
	}
}
