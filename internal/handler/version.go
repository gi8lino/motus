package handler

import "net/http"

// versionResponse is the response body for the version endpoint.
type versionResponse struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
}

// VersionInfo returns version metadata for the SPA.
func (a *API) VersionInfo(w http.ResponseWriter, r *http.Request) {
	a.respondJSON(w, http.StatusOK, versionResponse{Version: a.Version, Commit: a.Commit})
}
