package response

import (
	"log/slog"
	"net/http"

	httpkitresponse "github.com/containeroo/httpkit/response"
	errpkg "github.com/gi8lino/motus/internal/service/errors"
)

// RequestWriter is the shared request diagnostic writer.
type RequestWriter = httpkitresponse.RequestWriter

// NewRequestWriter wraps the writer with method, path and status diagnostics.
func NewRequestWriter(w http.ResponseWriter, r *http.Request) *RequestWriter {
	return httpkitresponse.NewRequestWriter(w, r)
}

// ServiceError translates domain failures at the HTTP boundary.
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

// InternalServerError records the private failure and sends a safe correlated JSON response.
func InternalServerError(logger *slog.Logger, w http.ResponseWriter, err error) {
	reference := httpkitresponse.LogFailure(logger, w, err)
	if request, ok := w.(*httpkitresponse.RequestWriter); ok && request.Status != 0 {
		return
	}
	Problem(w, http.StatusInternalServerError, "The request could not be processed. Reference: "+reference)
}
