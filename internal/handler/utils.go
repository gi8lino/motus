package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	errpkg "github.com/gi8lino/motus/internal/service/errors"
)

// encode encodes a value to JSON and writes it to the response.
func encode[T any](w http.ResponseWriter, status int, v T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}

// decode decodes a value from JSON and returns it.
func decode[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("decode json: %w", err)
	}
	return v, nil
}

// serviceStatus maps a service error to an HTTP status code.
func serviceStatus(err error) int {
	switch {
	case errpkg.IsKind(err, errpkg.ErrorValidation):
		return http.StatusBadRequest
	case errpkg.IsKind(err, errpkg.ErrorForbidden):
		return http.StatusForbidden
	case errpkg.IsKind(err, errpkg.ErrorNotFound):
		return http.StatusNotFound
	case errpkg.IsKind(err, errpkg.ErrorUnauthorized):
		return http.StatusUnauthorized
	case errpkg.IsKind(err, errpkg.ErrorInternal):
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
