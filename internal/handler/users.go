package handler

import (
	"net/http"

	"github.com/gi8lino/motus/internal/service/users"
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
	svc := users.New(a.Store, a.AuthHeader, a.AllowRegistration)
	type createUserRequest struct {
		Email     string `json:"email"`
		AvatarURL string `json:"avatarUrl"`
		Password  string `json:"password"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := decode[createUserRequest](r)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}
		user, err := svc.Create(r.Context(), req.Email, req.AvatarURL, req.Password)
		if err != nil {
			writeJSON(w, serviceStatus(err), apiError{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, user)
	}
}

// UpdateUserRole toggles admin access.
func (a *API) UpdateUserRole() http.HandlerFunc {
	svc := users.New(a.Store, a.AuthHeader, a.AllowRegistration)
	type updateUserRoleRequest struct {
		IsAdmin bool `json:"isAdmin"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		req, err := decode[updateUserRoleRequest](r)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}
		if err := svc.UpdateRole(r.Context(), id, req.IsAdmin); err != nil {
			writeJSON(w, serviceStatus(err), apiError{Error: err.Error()})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// Login validates credentials when using local authentication.
func (a *API) Login() http.HandlerFunc {
	svc := users.New(a.Store, a.AuthHeader, a.AllowRegistration)
	type loginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := decode[loginRequest](r)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}
		user, err := svc.Login(r.Context(), req.Email, req.Password)
		if err != nil {
			writeJSON(w, serviceStatus(err), apiError{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, user)
	}
}

// ChangePassword updates the password for the current user.
func (a *API) ChangePassword() http.HandlerFunc {
	svc := users.New(a.Store, a.AuthHeader, a.AllowRegistration)
	type changePasswordRequest struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := a.resolveUserID(r, "")
		if err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}
		req, err := decode[changePasswordRequest](r)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}
		if err := svc.ChangePassword(r.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
			writeJSON(w, serviceStatus(err), apiError{Error: err.Error()})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
