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

func TestApplicationServesFaviconSVG(t *testing.T) {
	const icon = `<svg xmlns="http://www.w3.org/2000/svg"></svg>`
	handler := newApplication(Config{
		InfrastructureConfig: InfrastructureConfig{
			RoutePrefix: "/motus",
			Assets: fstest.MapFS{
				"index.html": {Data: []byte(`<base href="{{.BaseHref}}">`)},
				"motus.svg":  {Data: []byte(icon)},
			},
		},
	})

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/motus/motus.svg", nil))

	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "image/svg+xml", response.Header().Get("Content-Type"))
	assert.Equal(t, icon, response.Body.String())
}
