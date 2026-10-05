package response

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"

	errpkg "github.com/gi8lino/motus/internal/service/errors"
)

// RequestWriter retains diagnostic context for a request and its access log.
type RequestWriter struct {
	http.ResponseWriter
	method         string
	path           string
	ErrorReference string
	Status         int
}

// NewRequestWriter wraps a response writer with request and status diagnostics.
func NewRequestWriter(w http.ResponseWriter, r *http.Request) *RequestWriter {
	return &RequestWriter{ResponseWriter: w, method: r.Method, path: r.URL.Path}
}

// Unwrap lets http.ResponseController reach optional transport capabilities.
func (w *RequestWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

// WriteHeader records the final status without treating informational headers as committed.
func (w *RequestWriter) WriteHeader(status int) {
	if w.Status != 0 {
		return
	}
	if status >= http.StatusOK {
		w.Status = status
	}
	w.ResponseWriter.WriteHeader(status)
}

// Write records an implicit successful status before writing response data.
func (w *RequestWriter) Write(data []byte) (int, error) {
	if w.Status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(data)
}

// FlushError flushes the underlying response while preserving recorded status.
func (w *RequestWriter) FlushError() error {
	if w.Status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	return http.NewResponseController(w.ResponseWriter).Flush()
}

// ServiceError translates typed application errors at the HTTP boundary.
func ServiceError(logger *slog.Logger, w http.ResponseWriter, err error) {
	switch {
	case errpkg.IsKind(err, errpkg.ErrorValidation):
		Problem(w, http.StatusBadRequest, err.Error())
	case errpkg.IsKind(err, errpkg.ErrorForbidden):
		Problem(w, http.StatusForbidden, "Forbidden.")
	case errpkg.IsKind(err, errpkg.ErrorNotFound):
		Problem(w, http.StatusNotFound, "Not found.")
	case errpkg.IsKind(err, errpkg.ErrorUnauthorized):
		Problem(w, http.StatusUnauthorized, "Authentication required.")
	default:
		InternalServerError(logger, w, err)
	}
}

// InternalServerError logs a failure and writes a safe HTTP 500 response with a reference ID.
func InternalServerError(logger *slog.Logger, w http.ResponseWriter, err error) {
	if logger == nil {
		logger = slog.Default()
	}
	if request, ok := w.(*RequestWriter); ok {
		logger = logger.With("method", request.method, "path", request.path)
	}
	reference := errorReference()
	if request, ok := w.(*RequestWriter); ok {
		request.ErrorReference = reference
	}
	logger.Error("request failed", "event", "request_failed", "error_reference", reference, "error", err)
	if request, ok := w.(*RequestWriter); ok && request.Status != 0 {
		return
	}
	Problem(w, http.StatusInternalServerError, "The request could not be processed. Reference: "+reference)
}

func errorReference() string {
	var data [16]byte
	if _, err := rand.Read(data[:]); err != nil {
		return "unavailable"
	}
	return hex.EncodeToString(data[:])
}
