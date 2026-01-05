package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// GetUsers lists all users.
func (a *API) GetUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users, err := a.Store.ListUsers(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, apiError{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, users)
	}
}

// CreateUser registers a new local user.
func (a *API) CreateUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Email     string `json:"email"`
			AvatarURL string `json:"avatarUrl"`
			Password  string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}
		email, err := normalizeEmail(req.Email)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}
		req.Password = strings.TrimSpace(req.Password)
		if a.AuthHeader == "" && !a.AllowRegistration {
			writeJSON(w, http.StatusForbidden, apiError{Error: "registration is disabled"})
			return
		}
		if a.AuthHeader == "" && req.Password == "" {
			writeJSON(w, http.StatusBadRequest, apiError{Error: "password is required"})
			return
		}
		passwordHash := ""
		if req.Password != "" {
			hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, apiError{Error: "unable to secure password"})
				return
			}
			passwordHash = string(hash)
		}
		user, err := a.Store.CreateUser(r.Context(), email, req.AvatarURL, passwordHash)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, apiError{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, user)
	}
}

// UpdateUserRole toggles admin access.
func (a *API) UpdateUserRole() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSpace(r.PathValue("id"))
		if id == "" {
			writeJSON(w, http.StatusBadRequest, apiError{Error: "user id is required"})
			return
		}
		var req struct {
			IsAdmin bool `json:"isAdmin"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}
		if err := a.Store.UpdateUserAdmin(r.Context(), id, req.IsAdmin); err != nil {
			writeJSON(w, http.StatusInternalServerError, apiError{Error: err.Error()})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// Login validates credentials when using local authentication.
func (a *API) Login() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if a.AuthHeader != "" {
			writeJSON(w, http.StatusForbidden, apiError{Error: "local login disabled"})
			return
		}
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}
		email, err := normalizeEmail(req.Email)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}
		req.Password = strings.TrimSpace(req.Password)
		if req.Password == "" {
			writeJSON(w, http.StatusBadRequest, apiError{Error: "email and password are required"})
			return
		}
		user, hash, err := a.Store.GetUserWithPassword(r.Context(), email)
		if err != nil || user == nil {
			writeJSON(w, http.StatusUnauthorized, apiError{Error: "invalid credentials"})
			return
		}
		if hash == "" {
			writeJSON(w, http.StatusUnauthorized, apiError{Error: "password not set"})
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
			writeJSON(w, http.StatusUnauthorized, apiError{Error: "invalid credentials"})
			return
		}
		writeJSON(w, http.StatusOK, user)
	}
}

// ChangePassword updates the password for the current user.
func (a *API) ChangePassword() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if a.AuthHeader != "" {
			writeJSON(w, http.StatusForbidden, apiError{Error: "passwords managed by proxy"})
			return
		}
		userID, err := a.resolveUserID(r, "")
		if err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}
		var req struct {
			CurrentPassword string `json:"currentPassword"`
			NewPassword     string `json:"newPassword"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}
		req.CurrentPassword = strings.TrimSpace(req.CurrentPassword)
		req.NewPassword = strings.TrimSpace(req.NewPassword)
		if req.CurrentPassword == "" || req.NewPassword == "" {
			writeJSON(w, http.StatusBadRequest, apiError{Error: "current and new password are required"})
			return
		}
		_, hash, err := a.Store.GetUserWithPassword(r.Context(), userID)
		if err != nil || hash == "" {
			writeJSON(w, http.StatusUnauthorized, apiError{Error: "invalid credentials"})
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.CurrentPassword)); err != nil {
			writeJSON(w, http.StatusUnauthorized, apiError{Error: "invalid credentials"})
			return
		}
		newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, apiError{Error: "unable to secure password"})
			return
		}
		if err := a.Store.UpdateUserPassword(r.Context(), userID, string(newHash)); err != nil {
			writeJSON(w, http.StatusInternalServerError, apiError{Error: err.Error()})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
