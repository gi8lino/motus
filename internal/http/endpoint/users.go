package endpoint

import (
	"errors"
	"github.com/gi8lino/motus/internal/db"
	"github.com/gi8lino/motus/internal/http/response"
	"log/slog"
	"net/http"

	httpauth "github.com/gi8lino/motus/internal/http/auth"
	"github.com/gi8lino/motus/internal/service/users"
)

// GetUsers lists all users.
func GetUsers(service *users.Service, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users, err := service.List(r.Context())
		if err != nil {
			logger.Error("list users failed", "event", "list_users_failed", "err", err)
			response.InternalServerError(logger, w, err)
			return
		}

		response.Respond(w, http.StatusOK, users)
	}
}

// SetUserLocalPassword sets or removes an administrator-managed local credential.
func SetUserLocalPassword(service *users.Service, logger *slog.Logger) http.HandlerFunc {
	type request struct {
		Password string `json:"password"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.PathValue("id")
		req, err := decode[request](r)
		if err != nil {
			response.Problem(w, http.StatusBadRequest, err.Error())
			return
		}
		if req.Password == "" {
			err = service.DisableLocalPassword(r.Context(), userID)
		} else {
			err = service.SetLocalPassword(r.Context(), userID, req.Password)
		}
		if err != nil {
			response.ServiceError(logger, w, err)
			return
		}
		response.Respond(w, http.StatusNoContent, nil)
	}
}

// LinkUserIdentity connects a provider identity to a selected native user.
func LinkUserIdentity(service *users.Service, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, err := decode[db.AuthIdentity](r)
		if err != nil {
			response.Problem(w, http.StatusBadRequest, err.Error())
			return
		}
		if err := service.LinkIdentity(r.Context(), r.PathValue("id"), identity); err != nil {
			response.ServiceError(logger, w, err)
			return
		}
		response.Respond(w, http.StatusNoContent, nil)
	}
}

// UnlinkUserIdentity removes a provider identity from a selected native user.
func UnlinkUserIdentity(service *users.Service, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := service.UnlinkIdentity(r.Context(), r.PathValue("id"), r.PathValue("provider"), r.PathValue("subject")); err != nil {
			response.ServiceError(logger, w, err)
			return
		}
		response.Respond(w, http.StatusNoContent, nil)
	}
}

// RevokeUserSessions invalidates every local browser session for a selected user.
func RevokeUserSessions(service *users.Service, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := service.RevokeSessions(r.Context(), r.PathValue("id")); err != nil {
			response.ServiceError(logger, w, err)
			return
		}
		response.Respond(w, http.StatusNoContent, nil)
	}
}

// CreateUser registers a new local user.
func CreateUser(
	service *users.Service,
	local *httpauth.Local,
	logger *slog.Logger,
) http.HandlerFunc {
	type createUserRequest struct {
		Email     string `json:"email"`
		AvatarURL string `json:"avatarUrl"`
		Password  string `json:"password"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := decode[createUserRequest](r)
		if err != nil {
			logger.Error("decode request failed", "event", "decode_request_failed", "err", err)
			response.Problem(w, http.StatusBadRequest, err.Error())
			return
		}

		user, err := service.Create(r.Context(), req.Email, req.AvatarURL, req.Password)
		if err != nil {
			logger.Error("create user failed", "event", "create_user_failed", "err", err)
			response.ServiceError(logger, w, err)
			return
		}
		if local != nil {
			token, err := local.StartSession(r.Context(), *user)
			if err != nil {
				response.InternalServerError(logger, w, errors.New("create session failed"))
				return
			}
			local.WriteSessionCookie(w, r, token)
		}

		logger.Info("user created",
			"event", "user_created",
			"resource", "user",
			"resource_id", user.ID,
			"user_id", user.ID,
		)
		response.Respond(w, http.StatusCreated, user)
	}
}

// UpdateUserRole toggles admin access.
func UpdateUserRole(service *users.Service, logger *slog.Logger) http.HandlerFunc {
	type updateUserRoleRequest struct {
		IsAdmin bool `json:"isAdmin"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		req, err := decode[updateUserRoleRequest](r)
		if err != nil {
			logger.Error("decode request failed", "event", "decode_request_failed", "err", err)
			response.Problem(w, http.StatusBadRequest, err.Error())
			return
		}

		if err := service.UpdateRole(r.Context(), id, req.IsAdmin); err != nil {
			logger.Error("update user role failed", "event", "update_user_role_failed", "err", err)
			response.ServiceError(logger, w, err)
			return
		}

		logger.Info("user role updated",
			"event", "user_role_updated",
			"resource", "user",
			"resource_id", id,
			"user_id", id,
			"is_admin", req.IsAdmin,
		)
		response.Respond(w, http.StatusNoContent, statusResponse{Status: "ok"})
	}
}

// Login validates credentials when using local authentication.
func Login(local *httpauth.Local, logger *slog.Logger) http.HandlerFunc {
	type loginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := decode[loginRequest](r)
		if err != nil {
			logger.Error("decode request failed", "event", "decode_request_failed", "err", err)
			response.Problem(w, http.StatusBadRequest, err.Error())
			return
		}

		user, token, err := local.SignIn(r.Context(), req.Email, req.Password)
		if err != nil {
			logger.Error("login user failed", "event", "login_user_failed", "err", err)
			response.ServiceError(logger, w, err)
			return
		}
		local.WriteSessionCookie(w, r, token)

		logger.Info("user login",
			"event", "user_login",
			"resource", "user",
			"resource_id", user.ID,
			"user_id", user.ID,
		)
		response.Respond(w, http.StatusOK, user)
	}
}

// Logout revokes the current local session.
func Logout(local *httpauth.Local, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		local.ClearSession(w, r)
		response.Respond(w, http.StatusNoContent, nil)
	}
}

// ChangePassword updates the password for the current user.
func ChangePassword(local *httpauth.Local, logger *slog.Logger) http.HandlerFunc {
	type changePasswordRequest struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := currentUser(r, "")
		if err != nil {
			logger.Error("resolve user id failed", "event", "resolve_user_id_failed", "err", err)
			response.Problem(w, http.StatusBadRequest, err.Error())
			return
		}

		req, err := decode[changePasswordRequest](r)
		if err != nil {
			logger.Error("decode request failed", "event", "decode_request_failed", "err", err)
			response.Problem(w, http.StatusBadRequest, err.Error())
			return
		}

		if err := local.ChangePassword(r.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
			logger.Error("change password failed", "event", "change_password_failed", "err", err)
			response.ServiceError(logger, w, err)
			return
		}

		logger.Info("user password changed",
			"event", "user_password_changed",
			"resource", "user",
			"resource_id", userID,
			"user_id", userID,
		)
		response.Respond(w, http.StatusNoContent, statusResponse{Status: "ok"})
	}
}

// UpdateUserName updates the current user's display name.
func UpdateUserName(service *users.Service, logger *slog.Logger) http.HandlerFunc {
	type updateUserNameRequest struct {
		Name string `json:"name"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := currentUser(r, "")
		if err != nil {
			logger.Error("resolve user id failed", "event", "resolve_user_id_failed", "err", err)
			response.Problem(w, http.StatusBadRequest, err.Error())
			return
		}

		req, err := decode[updateUserNameRequest](r)
		if err != nil {
			logger.Error("decode request failed", "event", "decode_request_failed", "err", err)
			response.Problem(w, http.StatusBadRequest, err.Error())
			return
		}

		if err := service.UpdateName(r.Context(), userID, req.Name); err != nil {
			logger.Error("user name update", "event", "user_name_update_failed", "err", err)
			response.ServiceError(logger, w, err)
			return
		}

		logger.Info("user name updated",
			"event", "user_name_updated",
			"resource", "user",
			"resource_id", userID,
			"user_id", userID,
		)
		response.Respond(w, http.StatusNoContent, statusResponse{Status: "ok"})
	}
}
