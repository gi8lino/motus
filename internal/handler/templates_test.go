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

func TestTemplatesHandlers(t *testing.T) {
	t.Run("List templates", func(t *testing.T) {
		store := &fakeStore{listTemplatesFn: func(context.Context) ([]db.Workout, error) {
			return []db.Workout{{ID: "t1", Name: "Template"}}, nil
		}}
		api := &API{Store: store}
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
		store := &fakeStore{createTemplateFn: func(context.Context, string, string) (*db.Workout, error) {
			return &db.Workout{ID: "t1", Name: "Template"}, nil
		}}
		api := &API{Store: store}
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
		store := &fakeStore{workoutWithStepsFn: func(context.Context, string) (*db.Workout, error) {
			return &db.Workout{ID: "t1", Name: "Template", IsTemplate: true}, nil
		}}
		api := &API{Store: store}
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
		store := &fakeStore{createWorkoutFromTemplate: func(context.Context, string, string, string) (*db.Workout, error) {
			return &db.Workout{ID: "w1", Name: "Copy"}, nil
		}}
		api := &API{Store: store}
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
