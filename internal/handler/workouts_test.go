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
	"github.com/gi8lino/motus/internal/service/workouts"
)

type fakeWorkoutStore struct {
	listTemplatesFn           func(context.Context) ([]db.Workout, error)
	createTemplateFn          func(context.Context, string, string) (*db.Workout, error)
	createWorkoutFromTemplate func(context.Context, string, string, string) (*db.Workout, error)
	workoutWithStepsFn        func(context.Context, string) (*db.Workout, error)
	createWorkoutFn           func(context.Context, *db.Workout) (*db.Workout, error)
	updateWorkoutFn           func(context.Context, *db.Workout) (*db.Workout, error)
	workoutsByUserFn          func(context.Context, string) ([]db.Workout, error)
	deleteWorkoutFn           func(context.Context, string) error
}

func (f *fakeWorkoutStore) ListTemplates(ctx context.Context) ([]db.Workout, error) {
	if f.listTemplatesFn == nil {
		return nil, nil
	}
	return f.listTemplatesFn(ctx)
}

func (f *fakeWorkoutStore) CreateTemplateFromWorkout(ctx context.Context, workoutID, name string) (*db.Workout, error) {
	if f.createTemplateFn == nil {
		return nil, nil
	}
	return f.createTemplateFn(ctx, workoutID, name)
}

func (f *fakeWorkoutStore) CreateWorkoutFromTemplate(ctx context.Context, templateID, userID, name string) (*db.Workout, error) {
	if f.createWorkoutFromTemplate == nil {
		return nil, nil
	}
	return f.createWorkoutFromTemplate(ctx, templateID, userID, name)
}

func (f *fakeWorkoutStore) WorkoutWithSteps(ctx context.Context, id string) (*db.Workout, error) {
	if f.workoutWithStepsFn == nil {
		return nil, nil
	}
	return f.workoutWithStepsFn(ctx, id)
}

func (f *fakeWorkoutStore) CreateWorkout(ctx context.Context, workout *db.Workout) (*db.Workout, error) {
	if f.createWorkoutFn == nil {
		return workout, nil
	}
	return f.createWorkoutFn(ctx, workout)
}

func (f *fakeWorkoutStore) UpdateWorkout(ctx context.Context, workout *db.Workout) (*db.Workout, error) {
	if f.updateWorkoutFn == nil {
		return workout, nil
	}
	return f.updateWorkoutFn(ctx, workout)
}

func (f *fakeWorkoutStore) DeleteWorkout(ctx context.Context, id string) error {
	if f.deleteWorkoutFn == nil {
		return nil
	}
	return f.deleteWorkoutFn(ctx, id)
}

func (f *fakeWorkoutStore) WorkoutsByUser(ctx context.Context, id string) ([]db.Workout, error) {
	if f.workoutsByUserFn == nil {
		return nil, nil
	}
	return f.workoutsByUserFn(ctx, id)
}

func TestWorkoutsHandlers(t *testing.T) {
	t.Run("List workouts", func(t *testing.T) {
		store := &fakeWorkoutStore{workoutsByUserFn: func(context.Context, string) ([]db.Workout, error) {
			return []db.Workout{{ID: "w1", Name: "Workout"}}, nil
		}}
		api := &API{Workouts: workouts.New(store)}
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
		store := &fakeWorkoutStore{createWorkoutFn: func(context.Context, *db.Workout) (*db.Workout, error) {
			return &db.Workout{ID: "w1", Name: "Workout"}, nil
		}}
		api := &API{Workouts: workouts.New(store)}
		h := api.CreateWorkout()
		body := strings.NewReader(`{"name":"Workout","steps":[{"type":"set","name":"Step","subsets":[{"name":"Main","exercises":[{"name":"Lift","reps":"5"}]}]}]}`)
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
		store := &fakeWorkoutStore{workoutWithStepsFn: func(context.Context, string) (*db.Workout, error) {
			return &db.Workout{ID: "w1", Name: "Workout"}, nil
		}}
		api := &API{Workouts: workouts.New(store)}
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
		store := &fakeWorkoutStore{workoutWithStepsFn: func(context.Context, string) (*db.Workout, error) {
			return &db.Workout{ID: "w1", Name: "Workout"}, nil
		}}
		api := &API{Workouts: workouts.New(store)}
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
		store := &fakeWorkoutStore{createWorkoutFn: func(context.Context, *db.Workout) (*db.Workout, error) {
			return &db.Workout{ID: "w1", Name: "Imported"}, nil
		}}
		api := &API{Workouts: workouts.New(store)}
		h := api.ImportWorkout()
		body := strings.NewReader(`{"userId":"user@example.com","workout":{"name":"Imported","steps":[{"type":"set","name":"Step","subsets":[{"name":"Main","exercises":[{"name":"Lift","reps":"5"}]}]}]}}`)
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
		store := &fakeWorkoutStore{updateWorkoutFn: func(context.Context, *db.Workout) (*db.Workout, error) {
			return &db.Workout{ID: "w1", Name: "Updated"}, nil
		}}
		api := &API{Workouts: workouts.New(store)}
		h := api.UpdateWorkout()
		body := strings.NewReader(`{"userId":"user@example.com","name":"Updated","steps":[{"type":"set","name":"Step","subsets":[{"name":"Main","exercises":[{"name":"Lift","reps":"5"}]}]}]}`)
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
		store := &fakeWorkoutStore{deleteWorkoutFn: func(context.Context, string) error { return nil }}
		api := &API{Workouts: workouts.New(store)}
		h := api.DeleteWorkout()
		req := httptest.NewRequest(http.MethodDelete, "/api/workouts/w1", nil)
		req.SetPathValue("id", "w1")
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		require.Equal(t, http.StatusNoContent, rec.Code)
	})
}
