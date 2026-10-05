package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
)

func TestApplicationMountsRoutePrefix(t *testing.T) {
	handler := newApplication(Config{
		InfrastructureConfig: InfrastructureConfig{
			RoutePrefix: "/motus",
			Assets: fstest.MapFS{
				"index.html": {Data: []byte(`<base href="{{.BaseHref}}">`)},
			},
		},
	})

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/motus/", nil))

	assert.Equal(t, http.StatusOK, response.Code)
	assert.Contains(t, response.Body.String(), `href="/motus/"`)
}
