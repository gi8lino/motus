package trainings

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUpdateFeedbackValidatesEffort(t *testing.T) {
	service := New(&fakeStore{}, nil)
	effort := 11
	err := service.UpdateFeedback(context.Background(), "training", "user", "hard", &effort)
	require.Error(t, err)
}
