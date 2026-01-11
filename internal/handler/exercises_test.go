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

type fakeExercisesStore struct {
	getUserFn                func(context.Context, string) (*db.User, error)
	updateUserNameFn         func(context.Context, string, string) error
	createUserFn             func(context.Context, string, string, string) (*db.User, error)
	listExercisesFn          func(context.Context, string) ([]db.Exercise, error)
	createExerciseFn         func(context.Context, string, string, bool) (*db.Exercise, error)
	getExerciseFn            func(context.Context, string) (*db.Exercise, error)
	renameExerciseFn         func(context.Context, string, string) (*db.Exercise, error)
	replaceExerciseForUserFn func(context.Context, string, string, string, string) error
	deleteExerciseFn         func(context.Context, string) error
	backfillCoreExercisesFn  func(context.Context) error
}

func (f *fakeExercisesStore) GetUser(ctx context.Context, id string) (*db.User, error) {
	if f.getUserFn == nil {
		return nil, nil
	}
	return f.getUserFn(ctx, id)
}

func (f *fakeExercisesStore) UpdateUserName(ctx context.Context, id, name string) error {
	if f.updateUserNameFn == nil {
		return nil
	}
	return f.updateUserNameFn(ctx, id, name)
}

func (f *fakeExercisesStore) CreateUser(ctx context.Context, email, avatarURL, passwordHash string) (*db.User, error) {
	if f.createUserFn == nil {
		return nil, nil
	}
	return f.createUserFn(ctx, email, avatarURL, passwordHash)
}

func (f *fakeExercisesStore) ListExercises(ctx context.Context, userID string) ([]db.Exercise, error) {
	if f.listExercisesFn == nil {
		return nil, nil
	}
	return f.listExercisesFn(ctx, userID)
}

func (f *fakeExercisesStore) CreateExercise(ctx context.Context, name, userID string, isCore bool) (*db.Exercise, error) {
	if f.createExerciseFn == nil {
		return nil, nil
	}
	return f.createExerciseFn(ctx, name, userID, isCore)
}

func (f *fakeExercisesStore) GetExercise(ctx context.Context, id string) (*db.Exercise, error) {
	if f.getExerciseFn == nil {
		return nil, nil
	}
	return f.getExerciseFn(ctx, id)
}

func (f *fakeExercisesStore) RenameExercise(ctx context.Context, id, name string) (*db.Exercise, error) {
	if f.renameExerciseFn == nil {
		return nil, nil
	}
	return f.renameExerciseFn(ctx, id, name)
}

func (f *fakeExercisesStore) ReplaceExerciseForUser(ctx context.Context, userID, oldID, newID, newName string) error {
	if f.replaceExerciseForUserFn == nil {
		return nil
	}
	return f.replaceExerciseForUserFn(ctx, userID, oldID, newID, newName)
}

func (f *fakeExercisesStore) DeleteExercise(ctx context.Context, id string) error {
	if f.deleteExerciseFn == nil {
		return nil
	}
	return f.deleteExerciseFn(ctx, id)
}

func (f *fakeExercisesStore) BackfillCoreExercises(ctx context.Context) error {
	if f.backfillCoreExercisesFn == nil {
		return nil
	}
	return f.backfillCoreExercisesFn(ctx)
}

func TestExercisesHandlers(t *testing.T) {
	t.Run("List exercises", func(t *testing.T) {
		store := &fakeExercisesStore{listExercisesFn: func(_ context.Context, userID string) ([]db.Exercise, error) {
			return []db.Exercise{{ID: "ex1", Name: "Burpee"}}, nil
		}}
		api := &API{ExercisesStore: store}
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
		store := &fakeExercisesStore{
			getUserFn: func(context.Context, string) (*db.User, error) {
				return &db.User{ID: "user@example.com"}, nil
			},
			createExerciseFn: func(context.Context, string, string, bool) (*db.Exercise, error) {
				return &db.Exercise{ID: "ex1", Name: "Burpee"}, nil
			},
		}
		api := &API{ExercisesStore: store}
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
		store := &fakeExercisesStore{
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
		api := &API{ExercisesStore: store}
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
		store := &fakeExercisesStore{
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
		api := &API{ExercisesStore: store}
		h := api.DeleteExercise()
		req := httptest.NewRequest(http.MethodDelete, "/api/exercises/ex1", nil)
		req.SetPathValue("id", "ex1")
		req.Header.Set("X-User-ID", "user@example.com")
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		require.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("Backfill exercises", func(t *testing.T) {
		store := &fakeExercisesStore{backfillCoreExercisesFn: func(context.Context) error { return nil }}
		api := &API{ExercisesStore: store}
		h := api.BackfillExercises()
		req := httptest.NewRequest(http.MethodPost, "/api/exercises/backfill", nil)
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
	})
}
