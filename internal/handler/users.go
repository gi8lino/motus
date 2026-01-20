package handler

import "net/http"

// GetUsers lists all users.
func (a *API) GetUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users, err := a.Users.List(r.Context())
		if err != nil {
			a.logRequestError(r, "list_users_failed", "list users failed", err)
			a.respondJSON(w, http.StatusInternalServerError, apiError{Error: err.Error()})
			return
		}

		a.respondJSON(w, http.StatusOK, users)
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
			a.logRequestError(r, "decode_request_failed", "decode request failed", err)
			a.respondJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}

		user, err := a.Users.Create(r.Context(), req.Email, req.AvatarURL, req.Password)
		if err != nil {
			a.Logger.Debug("create user failed", "err", err)
			a.respondJSON(w, serviceStatus(err), apiError{Error: err.Error()})
			return
		}

		a.businessLogger(r).Info("user created",
			"event", "user_created",
			"resource", "user",
			"resource_id", user.ID,
			"user_id", user.ID,
		)
		a.respondJSON(w, http.StatusCreated, user)
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
			a.logRequestError(r, "decode_request_failed", "decode request failed", err)
			a.respondJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}

		if err := a.Users.UpdateRole(r.Context(), id, req.IsAdmin); err != nil {
			a.logRequestError(r, "update_user_role_failed", "update user role failed", err)
			a.respondJSON(w, serviceStatus(err), apiError{Error: err.Error()})
			return
		}

		a.businessLogger(r).Info("user role updated",
			"event", "user_role_updated",
			"resource", "user",
			"resource_id", id,
			"user_id", id,
			"is_admin", req.IsAdmin,
		)
		a.respondJSON(w, http.StatusNoContent, statusResponse{Status: "ok"})
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
			a.logRequestError(r, "decode_request_failed", "decode request failed", err)
			a.respondJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}

		user, err := a.Users.Login(r.Context(), req.Email, req.Password)
		if err != nil {
			a.logRequestError(r, "login_user_failed", "login user failed", err)
			a.respondJSON(w, serviceStatus(err), apiError{Error: err.Error()})
			return
		}

		a.businessLogger(r).Info("user login",
			"event", "user_login",
			"resource", "user",
			"resource_id", user.ID,
			"user_id", user.ID,
		)
		a.respondJSON(w, http.StatusOK, user)
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
			a.logRequestError(r, "resolve_user_id_failed", "resolve user id failed", err)
			a.respondJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
		}

		req, err := decode[changePasswordRequest](r)
		if err != nil {
			a.logRequestError(r, "decode_request_failed", "decode request failed", err)
			a.respondJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}

		if err := a.Users.ChangePassword(r.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
			a.logRequestError(r, "change_password_failed", "change password failed", err)
			a.respondJSON(w, serviceStatus(err), apiError{Error: err.Error()})
			return
		}

		a.businessLogger(r).Info("user password changed",
			"event", "user_password_changed",
			"resource", "user",
			"resource_id", userID,
			"user_id", userID,
		)
		a.respondJSON(w, http.StatusNoContent, statusResponse{Status: "ok"})
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
			a.logRequestError(r, "resolve_user_id_failed", "resolve user id failed", err)
			a.respondJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}

		req, err := decode[updateUserNameRequest](r)
		if err != nil {
			a.logRequestError(r, "decode_request_failed", "decode request failed", err)
			a.respondJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}

		if err := a.Users.UpdateName(r.Context(), userID, req.Name); err != nil {
			a.logRequestError(r, "user_name_update_failed", "user name update", err)
			a.respondJSON(w, serviceStatus(err), apiError{Error: err.Error()})
			return
		}

		a.businessLogger(r).Info("user name updated",
			"event", "user_name_updated",
			"resource", "user",
			"resource_id", userID,
			"user_id", userID,
		)
		a.respondJSON(w, http.StatusNoContent, statusResponse{Status: "ok"})
	}
}
