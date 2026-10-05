package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProblem(t *testing.T) {
	recorder := httptest.NewRecorder()
	Problem(recorder, http.StatusBadRequest, "Invalid request.", NewFieldProblem("email", "Required."))

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Equal(t, "application/json; charset=utf-8", recorder.Header().Get("Content-Type"))

	var payload struct {
		Error    string            `json:"error"`
		Problems map[string]string `json:"problems"`
	}
	require.NoError(t, json.NewDecoder(recorder.Body).Decode(&payload))
	assert.Equal(t, "Invalid request.", payload.Error)
	assert.Equal(t, map[string]string{"email": "Required."}, payload.Problems)
}

func TestRequestWriterPreservesFirstFinalStatus(t *testing.T) {
	recorder := httptest.NewRecorder()
	writer := NewRequestWriter(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	writer.WriteHeader(http.StatusCreated)
	writer.WriteHeader(http.StatusInternalServerError)

	assert.Equal(t, http.StatusCreated, writer.Status)
	assert.Equal(t, http.StatusCreated, recorder.Code)
}
