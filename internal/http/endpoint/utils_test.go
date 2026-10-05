package endpoint

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecode(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}

	t.Run("Decodes JSON request body", func(t *testing.T) {
		t.Parallel()

		r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"name":"squat"}`))
		got, err := decode[payload](r)
		require.NoError(t, err)
		assert.Equal(t, "squat", got.Name)
	})

	t.Run("Returns error on invalid JSON", func(t *testing.T) {
		t.Parallel()

		r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"name":`)) // truncated JSON
		_, err := decode[payload](r)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "decode json")
	})
}
