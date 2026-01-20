package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gi8lino/motus/internal/logging"
	"github.com/stretchr/testify/require"
)

func TestRequestIDMiddleware(t *testing.T) {
	t.Parallel()

	t.Run("generates a request id when missing", func(t *testing.T) {
		t.Parallel()

		var seenID string
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			seenID = logging.RequestID(r.Context())
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		rec := httptest.NewRecorder()

		RequestIDMiddleware()(handler).ServeHTTP(rec, req)

		require.NotEmpty(t, seenID)
		require.Equal(t, seenID, rec.Header().Get(logging.RequestIDHeader))
	})

	t.Run("preserves existing request id", func(t *testing.T) {
		t.Parallel()

		const requestID = "req-123"
		var seenID string
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			seenID = logging.RequestID(r.Context())
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		req.Header.Set(logging.RequestIDHeader, requestID)
		rec := httptest.NewRecorder()

		RequestIDMiddleware()(handler).ServeHTTP(rec, req)

		require.Equal(t, requestID, seenID)
		require.Equal(t, requestID, rec.Header().Get(logging.RequestIDHeader))
	})
}
