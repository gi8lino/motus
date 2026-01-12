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
	"github.com/gi8lino/motus/internal/service/templates"
)

type fakeTemplateStore struct {
	listTemplatesFn           func(context.Context) ([]db.Workout, error)
	createTemplateFn          func(context.Context, string, string) (*db.Workout, error)
	createWorkoutFromTemplate func(context.Context, string, string, string) (*db.Workout, error)
	workoutWithStepsFn        func(context.Context, string) (*db.Workout, error)
}

func (f *fakeTemplateStore) ListTemplates(ctx context.Context) ([]db.Workout, error) {
	if f.listTemplatesFn == nil {
		return nil, nil
	}
	return f.listTemplatesFn(ctx)
}

func (f *fakeTemplateStore) CreateTemplateFromWorkout(ctx context.Context, workoutID, name string) (*db.Workout, error) {
	if f.createTemplateFn == nil {
		return nil, nil
	}
	return f.createTemplateFn(ctx, workoutID, name)
}

func (f *fakeTemplateStore) CreateWorkoutFromTemplate(ctx context.Context, templateID, userID, name string) (*db.Workout, error) {
	if f.createWorkoutFromTemplate == nil {
		return nil, nil
	}
	return f.createWorkoutFromTemplate(ctx, templateID, userID, name)
}

func (f *fakeTemplateStore) WorkoutWithSteps(ctx context.Context, id string) (*db.Workout, error) {
	if f.workoutWithStepsFn == nil {
		return nil, nil
	}
	return f.workoutWithStepsFn(ctx, id)
}

func TestTemplatesHandlers(t *testing.T) {
	t.Run("List templates", func(t *testing.T) {
		store := &fakeTemplateStore{listTemplatesFn: func(context.Context) ([]db.Workout, error) {
			return []db.Workout{{ID: "t1", Name: "Template"}}, nil
		}}
		api := &API{Templates: templates.New(store)}
		h := api.ListTemplates()
		req := httptest.NewRequest(http.MethodGet, "/api/templates", nil)
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		var payload []db.Workout
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&payload))
		require.Len(t, payload, 1)
		assert.Equal(t, "t1", payload[0].ID)
	})

	t.Run("Create template", func(t *testing.T) {
		store := &fakeTemplateStore{createTemplateFn: func(context.Context, string, string) (*db.Workout, error) {
			return &db.Workout{ID: "t1", Name: "Template"}, nil
		}}
		api := &API{Templates: templates.New(store)}
		h := api.CreateTemplate()
		body := strings.NewReader(`{"workoutId":"w1","name":"Template"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/templates", body)
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		require.Equal(t, http.StatusCreated, rec.Code)
		var payload db.Workout
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&payload))
		assert.Equal(t, "t1", payload.ID)
	})

	t.Run("Get template", func(t *testing.T) {
		store := &fakeTemplateStore{workoutWithStepsFn: func(context.Context, string) (*db.Workout, error) {
			return &db.Workout{ID: "t1", Name: "Template", IsTemplate: true}, nil
		}}
		api := &API{Templates: templates.New(store)}
		h := api.GetTemplate()
		req := httptest.NewRequest(http.MethodGet, "/api/templates/t1", nil)
		req.SetPathValue("id", "t1")
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		var payload db.Workout
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&payload))
		assert.Equal(t, "t1", payload.ID)
	})

	t.Run("Apply template", func(t *testing.T) {
		store := &fakeTemplateStore{createWorkoutFromTemplate: func(context.Context, string, string, string) (*db.Workout, error) {
			return &db.Workout{ID: "w1", Name: "Copy"}, nil
		}}
		api := &API{Templates: templates.New(store)}
		h := api.ApplyTemplate()
		body := strings.NewReader(`{"userId":"user@example.com","name":"Copy"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/templates/t1/apply", body)
		req.Header.Set("X-User-ID", "user@example.com")
		req.SetPathValue("id", "t1")
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		require.Equal(t, http.StatusCreated, rec.Code)
		var payload db.Workout
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&payload))
		assert.Equal(t, "w1", payload.ID)
	})
}
