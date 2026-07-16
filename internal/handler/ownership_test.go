package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/gi8lino/motus/internal/auth"
	"github.com/gi8lino/motus/internal/db"
	"github.com/gi8lino/motus/internal/service/workouts"
)

type ownershipAuthStore struct{ user *db.User }

func (s *ownershipAuthStore) GetUser(context.Context, string) (*db.User, error) { return s.user, nil }
func (s *ownershipAuthStore) CreateUser(context.Context, string, string, string) (*db.User, error) {
	return s.user, nil
}
func (s *ownershipAuthStore) CreateSession(context.Context, string, string, time.Time) error {
	return nil
}
func (s *ownershipAuthStore) GetSessionUser(context.Context, string, time.Time) (*db.User, error) {
	return s.user, nil
}

func TestWorkoutOwnership(t *testing.T) {
	store := &fakeWorkoutStore{workoutWithStepsFn: func(context.Context, string) (*db.Workout, error) {
		return &db.Workout{ID: "workout", UserID: "owner@example.com", Name: "Private"}, nil
	}}
	api := &API{
		AuthStore: &ownershipAuthStore{user: &db.User{ID: "other@example.com"}},
		Workouts:  workouts.New(store),
	}

	for _, test := range []struct {
		name    string
		method  string
		handler http.Handler
	}{
		{"get", http.MethodGet, api.GetWorkout()},
		{"export", http.MethodGet, api.ExportWorkout()},
		{"delete", http.MethodDelete, api.DeleteWorkout()},
	} {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(test.method, "/api/workouts/workout", nil)
			req.SetPathValue("id", "workout")
			req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "session"})
			rec := httptest.NewRecorder()

			test.handler.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusNotFound, rec.Code)
		})
	}
}
