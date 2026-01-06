package handler

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
)

func TestSPA(t *testing.T) {
	t.Run("Renders base href", func(t *testing.T) {
		spaFS := fstest.MapFS{
			"index.html": {Data: []byte(`<html><head><base href="{{ .BaseHref }}"></head></html>`)},
		}
		h := SPA(fs.FS(spaFS), "/motus")
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		require.Contains(t, rec.Body.String(), `base href="/motus/"`)
	})
}
