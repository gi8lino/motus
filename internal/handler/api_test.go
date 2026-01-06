package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithCORS(t *testing.T) {
	t.Parallel()

	t.Run("Options short circuit", func(t *testing.T) {
		t.Parallel()

		called := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
		})
		h := WithCORS("https://example.com", next)
		req := httptest.NewRequest(http.MethodOptions, "/", nil)
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		require.False(t, called, "expected next not to be called")
		assert.Equal(t, http.StatusNoContent, rec.Code)
		assert.Equal(t, "https://example.com", rec.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("Passthrough", func(t *testing.T) {
		t.Parallel()

		called := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusAccepted)
		})
		h := WithCORS("https://example.com", next)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		require.True(t, called, "expected next to be called")
		assert.Equal(t, http.StatusAccepted, rec.Code)
	})

	t.Run("Empty origin", func(t *testing.T) {
		t.Parallel()

		called := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
		})
		h := WithCORS("", next)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		require.True(t, called, "expected next to be called")
		assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
	})
}

func TestWriteJSON(t *testing.T) {
	t.Run("Writes JSON response", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		writeJSON(rec, http.StatusCreated, map[string]string{"status": "ok"})

		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

		var payload map[string]string
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&payload))
		assert.Equal(t, "ok", payload["status"])
	})
}
