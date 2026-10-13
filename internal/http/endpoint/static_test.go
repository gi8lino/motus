package endpoint

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
)

func TestStaticServiceWorkerIsNotCached(t *testing.T) {
	handler := Static(fstest.MapFS{
		"sw.js": {Data: []byte(`self.addEventListener("fetch", () => {});`)},
	})

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/sw.js", nil))

	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, noCacheControl, response.Header().Get("Cache-Control"))
}
