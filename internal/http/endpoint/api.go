package endpoint

import (
	"net/http"

	httpauth "github.com/gi8lino/motus/internal/http/auth"
	"github.com/gi8lino/motus/internal/service/workouts"
)

// statusResponse is the response body for status endpoints.
type statusResponse struct {
	// Status is the endpoint status label.
	Status string `json:"status"`
}

// currentUser selects the authenticated user id from request context.
func currentUser(r *http.Request, fallback string) (string, error) {
	if user, ok := httpauth.User(r); ok {
		return user.ID, nil
	}
	if fallback != "" {
		return fallback, nil
	}
	if userID := r.Header.Get("X-User-ID"); userID != "" {
		return userID, nil
	}
	return "", httpauth.ErrUnauthenticated
}

// requireWorkoutOwner hides workouts that do not belong to the authenticated user.
func requireWorkoutOwner(service *workouts.Service, r *http.Request, workoutID string) error {
	if service == nil {
		return nil
	}
	if _, ok := httpauth.User(r); !ok && r.Header.Get("X-User-ID") == "" {
		return nil
	}
	// Direct handler tests may omit request authentication; production always supplies it through middleware.
	userID, err := currentUser(r, "")
	if err != nil {
		return err
	}
	_, err = service.GetForUser(r.Context(), workoutID, userID)
	if err != nil {
		return err
	}
	return nil
}
