package endpoint

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	appkitweb "github.com/containeroo/appkit/web"
	"github.com/stretchr/testify/require"
)

func TestSPA(t *testing.T) {
	handler := (appkitweb.SPA{Assets: fstest.MapFS{"index.html": {Data: []byte(`<html><head><base href="{{ .BaseHref }}"></head></html>`)}}, BasePath: "/motus"}).Handler()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code)
	require.Contains(t, response.Body.String(), `base href="/motus/"`)
}
