package middleware

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAccessLog(t *testing.T) {
	t.Parallel()

	t.Run("logs the standard request completion attributes", func(t *testing.T) {
		t.Parallel()

		var logs strings.Builder
		logger := slog.New(slog.NewTextHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug}))

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(10 * time.Millisecond) // simulate latency
			w.WriteHeader(http.StatusTeapot)
			w.Write([]byte("short and stout")) // nolint:errcheck
		})

		req := httptest.NewRequest(http.MethodPost, "/bump", nil)
		req.Header.Set("User-Agent", "unit-test")
		req.RemoteAddr = "1.2.3.4:5678"

		rec := httptest.NewRecorder()
		logged := AccessLog(logger)(handler)

		logged.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusTeapot, rec.Code)
		assert.Contains(t, logs.String(), "method=POST")
		assert.Contains(t, logs.String(), "path=/bump")
		assert.Contains(t, logs.String(), "event=request_complete")
		assert.Contains(t, logs.String(), "duration=")
		assert.Contains(t, logs.String(), "msg=request")
	})

	t.Run("records requests that use implicit success", func(t *testing.T) {
		t.Parallel()

		var logs strings.Builder
		logger := slog.New(slog.NewTextHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug}))

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("ok")) // nolint:errcheck
		})

		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		rec := httptest.NewRecorder()
		logged := AccessLog(logger)(handler)

		logged.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, logs.String(), "event=request_complete")
	})
}
