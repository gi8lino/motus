package endpoint

import (
	"github.com/gi8lino/motus/internal/http/response"
	"log/slog"
	"net/http"
)

// versionResponse is the response body for the version endpoint.
type versionResponse struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
}

// VersionInfo returns version metadata for the SPA.
func VersionInfo(version, commit string, logger *slog.Logger, w http.ResponseWriter, r *http.Request) {
	response.Respond(w, http.StatusOK, versionResponse{Version: version, Commit: commit})
}
