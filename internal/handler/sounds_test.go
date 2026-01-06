package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListSounds(t *testing.T) {
	t.Run("Returns sounds", func(t *testing.T) {
		api := &API{}
		h := api.ListSounds()
		req := httptest.NewRequest(http.MethodGet, "/api/sounds", nil)
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		var payload []map[string]any
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&payload))
		require.NotEmpty(t, payload)
	})
}
