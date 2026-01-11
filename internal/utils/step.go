package utils

// StepType defines the available step kinds.
type StepType string

const (
	StepTypeSet   StepType = "set"
	StepTypePause StepType = "pause"
)

// NormalizeStepType converts an arbitrary value to a StepType.
func NormalizeStepType(value string) StepType {
	switch NormalizeToken(value) {
	case string(StepTypePause):
		return StepTypePause
	default:
		return StepTypeSet
	}
}

// String returns the string representation of the StepType.
func (s StepType) String() string {
	return string(s)
}
