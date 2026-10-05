package endpoint

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

type fakeHealthStore struct {
	pingFn func(context.Context) error
}

func (f *fakeHealthStore) Ping(ctx context.Context) error {
	if f.pingFn == nil {
		return nil
	}
	return f.pingFn(ctx)
}

func TestHealth(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	Health().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	require.Equal(t, http.StatusOK, recorder.Code)
}

func TestReady(t *testing.T) {
	t.Parallel()
	t.Run("Returns ok", func(t *testing.T) {
		t.Parallel()
		store := &fakeHealthStore{pingFn: func(context.Context) error { return nil }}
		h := Ready(store)
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("Reports unavailable dependency", func(t *testing.T) {
		t.Parallel()
		store := &fakeHealthStore{pingFn: func(context.Context) error { return errors.New("unavailable") }}
		recorder := httptest.NewRecorder()

		Ready(store).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/readyz", nil))

		require.Equal(t, http.StatusServiceUnavailable, recorder.Code)
	})
}
