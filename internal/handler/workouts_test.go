package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gi8lino/motus/internal/db"
)

func TestWorkoutsHandlers(t *testing.T) {
	t.Run("List workouts", func(t *testing.T) {
		store := &fakeStore{workoutsByUserFn: func(context.Context, string) ([]db.Workout, error) {
			return []db.Workout{{ID: "w1", Name: "Workout"}}, nil
		}}
		api := &API{Store: store}
		h := api.GetWorkouts()
		req := httptest.NewRequest(http.MethodGet, "/api/workouts", nil)
		req.SetPathValue("id", "user@example.com")
		req.Header.Set("X-User-ID", "user@example.com")
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		var payload []db.Workout
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&payload))
		require.Len(t, payload, 1)
		assert.Equal(t, "w1", payload[0].ID)
	})

	t.Run("Create workout", func(t *testing.T) {
		store := &fakeStore{createWorkoutFn: func(context.Context, *db.Workout) (*db.Workout, error) {
			return &db.Workout{ID: "w1", Name: "Workout"}, nil
		}}
		api := &API{Store: store}
		h := api.CreateWorkout()
		body := strings.NewReader(`{"name":"Workout","steps":[{"type":"set","name":"Step"}]}`)
		req := httptest.NewRequest(http.MethodPost, "/api/workouts", body)
		req.SetPathValue("id", "user@example.com")
		req.Header.Set("X-User-ID", "user@example.com")
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		require.Equal(t, http.StatusCreated, rec.Code)
		var payload db.Workout
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&payload))
		assert.Equal(t, "w1", payload.ID)
	})

	t.Run("Get workout", func(t *testing.T) {
		store := &fakeStore{workoutWithStepsFn: func(context.Context, string) (*db.Workout, error) {
			return &db.Workout{ID: "w1", Name: "Workout"}, nil
		}}
		api := &API{Store: store}
		h := api.GetWorkout()
		req := httptest.NewRequest(http.MethodGet, "/api/workouts/w1", nil)
		req.SetPathValue("id", "w1")
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		var payload db.Workout
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&payload))
		assert.Equal(t, "w1", payload.ID)
	})

	t.Run("Export workout", func(t *testing.T) {
		store := &fakeStore{workoutWithStepsFn: func(context.Context, string) (*db.Workout, error) {
			return &db.Workout{ID: "w1", Name: "Workout"}, nil
		}}
		api := &API{Store: store}
		h := api.ExportWorkout()
		req := httptest.NewRequest(http.MethodGet, "/api/workouts/w1/export", nil)
		req.SetPathValue("id", "w1")
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		var payload db.Workout
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&payload))
		assert.Equal(t, "w1", payload.ID)
	})

	t.Run("Import workout", func(t *testing.T) {
		store := &fakeStore{createWorkoutFn: func(context.Context, *db.Workout) (*db.Workout, error) {
			return &db.Workout{ID: "w1", Name: "Imported"}, nil
		}}
		api := &API{Store: store}
		h := api.ImportWorkout()
		body := strings.NewReader(`{"userId":"user@example.com","workout":{"name":"Imported","steps":[{"type":"set","name":"Step"}]}}`)
		req := httptest.NewRequest(http.MethodPost, "/api/workouts/import", body)
		req.Header.Set("X-User-ID", "user@example.com")
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		require.Equal(t, http.StatusCreated, rec.Code)
		var payload db.Workout
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&payload))
		assert.Equal(t, "w1", payload.ID)
	})

	t.Run("Update workout", func(t *testing.T) {
		store := &fakeStore{updateWorkoutFn: func(context.Context, *db.Workout) (*db.Workout, error) {
			return &db.Workout{ID: "w1", Name: "Updated"}, nil
		}}
		api := &API{Store: store}
		h := api.UpdateWorkout()
		body := strings.NewReader(`{"userId":"user@example.com","name":"Updated","steps":[{"type":"set","name":"Step"}]}`)
		req := httptest.NewRequest(http.MethodPut, "/api/workouts/w1", body)
		req.SetPathValue("id", "w1")
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		var payload db.Workout
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&payload))
		assert.Equal(t, "Updated", payload.Name)
	})

	t.Run("Delete workout", func(t *testing.T) {
		store := &fakeStore{deleteWorkoutFn: func(context.Context, string) error { return nil }}
		api := &API{Store: store}
		h := api.DeleteWorkout()
		req := httptest.NewRequest(http.MethodDelete, "/api/workouts/w1", nil)
		req.SetPathValue("id", "w1")
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		require.Equal(t, http.StatusNoContent, rec.Code)
	})
}
