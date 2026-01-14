package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gi8lino/motus/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncode(t *testing.T) {
	t.Run("Writes JSON response", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		require.NoError(t, encode(rec, http.StatusCreated, map[string]string{"status": "ok"}))

		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

		var payload map[string]string
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&payload))
		assert.Equal(t, "ok", payload["status"])
	})

	t.Run("Returns error for non-encodable value and still writes headers/status", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()

		// channels are not JSON-encodable, json will error with "unsupported type"
		err := encode(rec, http.StatusOK, make(chan int))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "encode json")

		// Because encode writes headers/status before encoding,
		// we still expect these to be set even on error.
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	})
}

func TestDecode(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}

	t.Run("Decodes JSON request body", func(t *testing.T) {
		t.Parallel()

		r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"name":"squat"}`))
		got, err := decode[payload](r)
		require.NoError(t, err)
		assert.Equal(t, "squat", got.Name)
	})

	t.Run("Returns error on invalid JSON", func(t *testing.T) {
		t.Parallel()

		r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"name":`)) // truncated JSON
		_, err := decode[payload](r)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "decode json")
	})
}

func TestServiceStatus(t *testing.T) {
	t.Parallel()

	t.Run("Validation -> 400", func(t *testing.T) {
		t.Parallel()
		err := &service.Error{Kind: service.ErrorValidation, Err: errors.New("validation error")}
		assert.Equal(t, http.StatusBadRequest, serviceStatus(err))
	})

	t.Run("Forbidden -> 403", func(t *testing.T) {
		t.Parallel()
		err := &service.Error{Kind: service.ErrorForbidden, Err: errors.New("forbidden error")}
		assert.Equal(t, http.StatusForbidden, serviceStatus(err))
	})

	t.Run("NotFound -> 404", func(t *testing.T) {
		t.Parallel()
		err := &service.Error{Kind: service.ErrorNotFound, Err: errors.New("not found error")}
		assert.Equal(t, http.StatusNotFound, serviceStatus(err))
	})

	t.Run("Unauthorized -> 401", func(t *testing.T) {
		t.Parallel()
		err := &service.Error{Kind: service.ErrorUnauthorized, Err: errors.New("unauthorized error")}
		assert.Equal(t, http.StatusUnauthorized, serviceStatus(err))
	})

	t.Run("Internal -> 500", func(t *testing.T) {
		t.Parallel()
		err := &service.Error{Kind: service.ErrorInternal, Err: errors.New("internal error")}
		assert.Equal(t, http.StatusInternalServerError, serviceStatus(err))
	})
}
