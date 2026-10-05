package endpoint

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gi8lino/motus/internal/http/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInternalServerErrorSanitizesDetails(t *testing.T) {
	rec := httptest.NewRecorder()
	response.InternalServerError(nil, rec, io.ErrUnexpectedEOF)
	var payload struct {
		Error string `json:"error"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&payload))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, payload.Error, "The request could not be processed. Reference: ")
	assert.NotContains(t, payload.Error, "unexpected EOF")
}

func TestDecodeRejectsUnknownAndTrailingJSON(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}
	for _, body := range []string{`{"name":"ok","extra":true}`, `{"name":"ok"} {"name":"again"}`} {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		_, err := decode[payload](req)
		require.Error(t, err)
	}
}
