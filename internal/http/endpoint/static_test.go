package endpoint

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/containeroo/httpkit/assets"
	"github.com/stretchr/testify/assert"
)

func TestStaticServiceWorkerIsNotCached(t *testing.T) {
	handler := assets.Static(assets.StaticConfig{Files: fstest.MapFS{"sw.js": {Data: []byte(`self.addEventListener("fetch", () => {});`)}}, NoCacheFiles: []string{"sw.js"}})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/sw.js", nil))
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "no-cache", recorder.Header().Get("Cache-Control"))
}
