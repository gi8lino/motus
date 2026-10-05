package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestManagementRejectsUnsupportedMethods(t *testing.T) {
	response := httptest.NewRecorder()
	newManagement(Config{}).ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/healthz", nil))

	assert.Equal(t, http.StatusMethodNotAllowed, response.Code)
	assert.Equal(t, "GET, HEAD", response.Header().Get("Allow"))
}
