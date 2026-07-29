package routes

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gi8lino/motus/internal/handler"
)

func TestNewRouter(t *testing.T) {
	t.Parallel()

	// In-memory file system with a minimal SPA template.
	webFS := fstest.MapFS{
		"index.html":             &fstest.MapFile{Data: []byte(`<!doctype html><base href="{{ .BaseHref }}"><meta name="routePrefix" content="{{ .RoutePrefix }}">`)},
		"assets/app-deadbeef.js": &fstest.MapFile{Data: []byte(strings.Repeat("const value = 'compress me';\n", 200))},
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	api := &handler.API{
		Logger:            logger,
		AuthHeader:        "X-User-Email",
		AllowRegistration: true,
		Version:           "v1.2.3",
		Commit:            "abc123",
	}

	router, err := NewRouter(webFS, "/motus", logger, api, false)
	require.NoError(t, err)

	t.Run("GET /motus/", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/motus/", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `base href="/motus/"`)
		assert.Contains(t, rec.Body.String(), `content="/motus"`)
		assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))
	})

	t.Run("GET immutable compressed asset", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(
			http.MethodGet,
			"/motus/assets/app-deadbeef.js",
			nil,
		)
		req.Header.Set("Accept-Encoding", "gzip")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, immutableCacheControl, rec.Header().Get("Cache-Control"))
		assert.Equal(t, "gzip", rec.Header().Get("Content-Encoding"))
		assert.Contains(t, rec.Header().Values("Vary"), "Accept-Encoding")

		reader, err := gzip.NewReader(rec.Body)
		require.NoError(t, err)
		body, err := io.ReadAll(reader)
		require.NoError(t, err)
		assert.Equal(
			t,
			strings.Repeat("const value = 'compress me';\n", 200),
			string(body),
		)
	})

	t.Run("GET /motus/api/config", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/motus/api/config", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		var payload map[string]any
		err := json.Unmarshal(rec.Body.Bytes(), &payload)
		require.NoError(t, err)
		assert.Equal(t, true, payload["authHeaderEnabled"])
		assert.Equal(t, true, payload["allowRegistration"])
		assert.Equal(t, "v1.2.3", payload["version"])
		assert.Equal(t, "abc123", payload["commit"])
	})
}
