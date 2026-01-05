package handler

import "net/http"

// VersionInfo returns version metadata for the SPA.
func (a *API) VersionInfo(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"version": a.Version, "commit": a.Commit})
}
