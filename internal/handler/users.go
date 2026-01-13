package handler

import "net/http"

// GetUsers lists all users.
func (a *API) GetUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users, err := a.Users.List(r.Context())
		if err != nil {
			if err := encode(w, r, http.StatusInternalServerError, apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("user list", "err", err)
			}
			return
		}

		if err := encode(w, r, http.StatusOK, users); err != nil {
			a.Logger.Error("user list encode", "err", err)
		}
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
			if err := encode(w, r, http.StatusBadRequest, apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("user create decode", "err", err)
			}
			return
		}

		user, err := a.Users.Create(r.Context(), req.Email, req.AvatarURL, req.Password)
		if err != nil {
			if err := encode(w, r, serviceStatus(err), apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("user create", "err", err)
			}
			return
		}

		if err := encode(w, r, http.StatusCreated, user); err != nil {
			a.Logger.Error("user create encode", "err", err)
		}
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
			if err := encode(w, r, http.StatusBadRequest, apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("user role decode", "err", err)
			}
			return
		}

		if err := a.Users.UpdateRole(r.Context(), id, req.IsAdmin); err != nil {
			if err := encode(w, r, serviceStatus(err), apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("user role", "err", err)
			}
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
			if err := encode(w, r, http.StatusBadRequest, apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("login decode", "err", err)
			}
			return
		}

		user, err := a.Users.Login(r.Context(), req.Email, req.Password)
		if err != nil {
			if err := encode(w, r, serviceStatus(err), apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("login", "err", err)
			}
			return
		}

		if err := encode(w, r, http.StatusOK, user); err != nil {
			a.Logger.Error("login encode", "err", err)
		}
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
			if err := encode(w, r, http.StatusBadRequest, apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("password user", "err", err)
			}
			return
		}

		req, err := decode[changePasswordRequest](r)
		if err != nil {
			if err := encode(w, r, http.StatusBadRequest, apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("password decode", "err", err)
			}
			return
		}

		if err := a.Users.ChangePassword(r.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
			if err := encode(w, r, serviceStatus(err), apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("password change", "err", err)
			}
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
			if err := encode(w, r, http.StatusBadRequest, apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("user name user", "err", err)
			}
			return
		}

		req, err := decode[updateUserNameRequest](r)
		if err != nil {
			if err := encode(w, r, http.StatusBadRequest, apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("user name decode", "err", err)
			}
			return
		}

		if err := a.Users.UpdateName(r.Context(), userID, req.Name); err != nil {
			if err := encode(w, r, serviceStatus(err), apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("user name update", "err", err)
			}
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
