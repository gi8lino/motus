package routes

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
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
		"web/dist/index.html": &fstest.MapFile{Data: []byte(`<!doctype html><base href="{{ .BaseHref }}"><meta name="routePrefix" content="{{ .RoutePrefix }}">`)},
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	api := &handler.API{
		Logger:            logger,
		AuthHeader:        "X-User-Email",
		AllowRegistration: true,
		Version:           "v1.2.3",
		Commit:            "abc123",
	}

	router, err := NewRouter(webFS, "/motus", api, false)
	require.NoError(t, err)

	t.Run("GET /motus/", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/motus/", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `base href="/motus/"`)
		assert.Contains(t, rec.Body.String(), `content="/motus"`)
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
