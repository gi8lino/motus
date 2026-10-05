package endpoint

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gi8lino/motus/internal/db"
	httpauth "github.com/gi8lino/motus/internal/http/auth"
	"github.com/gi8lino/motus/internal/service/workouts"
)

type ownershipAuthStore struct{ user *db.User }

func (s *ownershipAuthStore) GetUser(context.Context, string) (*db.User, error) { return s.user, nil }
func (s *ownershipAuthStore) CreateUser(context.Context, string, string, string) (*db.User, error) {
	return s.user, nil
}

func TestWorkoutOwnership(t *testing.T) {
	store := &fakeWorkoutStore{workoutWithStepsFn: func(context.Context, string) (*db.Workout, error) {
		return &db.Workout{ID: "workout", UserID: "owner@example.com", Name: "Private"}, nil
	}}
	api := &testEndpoints{Logger: slog.Default(),
		Workouts: workouts.New(store),
	}

	for _, test := range []struct {
		name    string
		method  string
		handler http.Handler
	}{
		{"get", http.MethodGet, GetWorkout(api.Workouts, api.Logger)},
		{"export", http.MethodGet, ExportWorkout(api.Workouts, api.Logger)},
		{"delete", http.MethodDelete, DeleteWorkout(api.Workouts, api.Logger)},
	} {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(test.method, "/api/workouts/workout", nil)
			req = httpauth.WithUser(req, db.User{ID: "other@example.com"})
			req.SetPathValue("id", "workout")
			rec := httptest.NewRecorder()

			test.handler.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusNotFound, rec.Code)
		})
	}
}
