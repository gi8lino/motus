package handler

import "net/http"

// GetUsers lists all users.
func (a *API) GetUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users, err := a.Users.List(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, apiError{Error: err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, users)
	}
}

// CreateUser registers a new local user.
func (a *API) CreateUser() http.HandlerFunc {
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

		user, err := a.Users.Create(r.Context(), req.Email, req.AvatarURL, req.Password)
		if err != nil {
			writeJSON(w, serviceStatus(err), apiError{Error: err.Error()})
			return
		}

		writeJSON(w, http.StatusCreated, user)
	}
}

// UpdateUserRole toggles admin access.
func (a *API) UpdateUserRole() http.HandlerFunc {
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

		if err := a.Users.UpdateRole(r.Context(), id, req.IsAdmin); err != nil {
			writeJSON(w, serviceStatus(err), apiError{Error: err.Error()})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// Login validates credentials when using local authentication.
func (a *API) Login() http.HandlerFunc {
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

		user, err := a.Users.Login(r.Context(), req.Email, req.Password)
		if err != nil {
			writeJSON(w, serviceStatus(err), apiError{Error: err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, user)
	}
}

// ChangePassword updates the password for the current user.
func (a *API) ChangePassword() http.HandlerFunc {
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

		if err := a.Users.ChangePassword(r.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
			writeJSON(w, serviceStatus(err), apiError{Error: err.Error()})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// UpdateUserName updates the current user's display name.
func (a *API) UpdateUserName() http.HandlerFunc {
	type updateUserNameRequest struct {
		Name string `json:"name"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := a.resolveUserID(r, "")
		if err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}

		req, err := decode[updateUserNameRequest](r)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}

		if err := a.Users.UpdateName(r.Context(), userID, req.Name); err != nil {
			writeJSON(w, serviceStatus(err), apiError{Error: err.Error()})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
