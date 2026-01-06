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

func TestExercisesHandlers(t *testing.T) {
	t.Run("List exercises", func(t *testing.T) {
		store := &fakeStore{listExercisesFn: func(_ context.Context, userID string) ([]db.Exercise, error) {
			return []db.Exercise{{ID: "ex1", Name: "Burpee"}}, nil
		}}
		api := &API{Store: store}
		h := api.ListExercises()
		req := httptest.NewRequest(http.MethodGet, "/api/exercises", nil)
		req.Header.Set("X-User-ID", "user@example.com")
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		var payload []db.Exercise
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&payload))
		require.Len(t, payload, 1)
		assert.Equal(t, "ex1", payload[0].ID)
	})

	t.Run("Create exercise", func(t *testing.T) {
		store := &fakeStore{
			getUserFn: func(context.Context, string) (*db.User, error) {
				return &db.User{ID: "user@example.com"}, nil
			},
			createExerciseFn: func(context.Context, string, string, bool) (*db.Exercise, error) {
				return &db.Exercise{ID: "ex1", Name: "Burpee"}, nil
			},
		}
		api := &API{Store: store}
		h := api.CreateExercise()
		body := strings.NewReader(`{"name":"Burpee","isCore":false}`)
		req := httptest.NewRequest(http.MethodPost, "/api/exercises", body)
		req.Header.Set("X-User-ID", "user@example.com")
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		require.Equal(t, http.StatusCreated, rec.Code)
		var payload db.Exercise
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&payload))
		assert.Equal(t, "ex1", payload.ID)
	})

	t.Run("Update exercise", func(t *testing.T) {
		store := &fakeStore{
			getUserFn: func(context.Context, string) (*db.User, error) {
				return &db.User{ID: "user@example.com", IsAdmin: true}, nil
			},
			getExerciseFn: func(context.Context, string) (*db.Exercise, error) {
				return &db.Exercise{ID: "ex1", Name: "Burpee", OwnerUserID: "user@example.com"}, nil
			},
			renameExerciseFn: func(context.Context, string, string) (*db.Exercise, error) {
				return &db.Exercise{ID: "ex1", Name: "Burpee 2"}, nil
			},
		}
		api := &API{Store: store}
		h := api.UpdateExercise()
		body := strings.NewReader(`{"name":"Burpee 2"}`)
		req := httptest.NewRequest(http.MethodPut, "/api/exercises/ex1", body)
		req.SetPathValue("id", "ex1")
		req.Header.Set("X-User-ID", "user@example.com")
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		var payload db.Exercise
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&payload))
		assert.Equal(t, "Burpee 2", payload.Name)
	})

	t.Run("Delete exercise", func(t *testing.T) {
		store := &fakeStore{
			getUserFn: func(context.Context, string) (*db.User, error) {
				return &db.User{ID: "user@example.com", IsAdmin: true}, nil
			},
			getExerciseFn: func(context.Context, string) (*db.Exercise, error) {
				return &db.Exercise{ID: "ex1", Name: "Burpee", OwnerUserID: "user@example.com"}, nil
			},
			deleteExerciseFn: func(context.Context, string) error {
				return nil
			},
		}
		api := &API{Store: store}
		h := api.DeleteExercise()
		req := httptest.NewRequest(http.MethodDelete, "/api/exercises/ex1", nil)
		req.SetPathValue("id", "ex1")
		req.Header.Set("X-User-ID", "user@example.com")
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		require.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("Backfill exercises", func(t *testing.T) {
		store := &fakeStore{backfillCoreExercisesFn: func(context.Context) error { return nil }}
		api := &API{Store: store}
		h := api.BackfillExercises()
		req := httptest.NewRequest(http.MethodPost, "/api/exercises/backfill", nil)
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
	})
}
