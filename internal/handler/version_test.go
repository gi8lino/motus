package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionInfo(t *testing.T) {
	t.Run("Returns version payload", func(t *testing.T) {
		api := &API{Version: "v1", Commit: "c1"}
		req := httptest.NewRequest(http.MethodGet, "/api/version", nil)
		rec := httptest.NewRecorder()

		api.VersionInfo(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		var payload versionResponse
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&payload))
		assert.Equal(t, "v1", payload.Version)
		assert.Equal(t, "c1", payload.Commit)
	})
}
