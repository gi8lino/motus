package endpoint

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionInfo(t *testing.T) {
	t.Run("Returns version payload", func(t *testing.T) {
		api := &testEndpoints{Logger: slog.Default(), Version: "v1", Commit: "c1"}
		req := httptest.NewRequest(http.MethodGet, "/api/version", nil)
		rec := httptest.NewRecorder()

		func(w http.ResponseWriter, r *http.Request) { VersionInfo(api.Version, api.Commit, api.Logger, w, r) }(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		var payload versionResponse
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&payload))
		assert.Equal(t, "v1", payload.Version)
		assert.Equal(t, "c1", payload.Commit)
	})
}
