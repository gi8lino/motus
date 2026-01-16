package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gi8lino/motus/internal/db"
	"github.com/gi8lino/motus/internal/service/sounds"
	"github.com/gi8lino/motus/internal/service/trainings"
)

type fakeTrainingStore struct {
	workoutWithStepsFn    func(context.Context, string) (*db.Workout, error)
	trainingHistoryFn     func(context.Context, string, int) ([]db.TrainingLog, error)
	trainingStepTimingsFn func(context.Context, string) ([]db.TrainingStepLog, error)
	recordTrainingFn      func(context.Context, db.TrainingLog, []db.TrainingStepLog) error
}

func (f *fakeTrainingStore) WorkoutWithSteps(ctx context.Context, id string) (*db.Workout, error) {
	if f.workoutWithStepsFn == nil {
		return nil, nil
	}
	return f.workoutWithStepsFn(ctx, id)
}

func (f *fakeTrainingStore) TrainingHistory(ctx context.Context, userID string, limit int) ([]db.TrainingLog, error) {
	if f.trainingHistoryFn == nil {
		return nil, nil
	}
	return f.trainingHistoryFn(ctx, userID, limit)
}

func (f *fakeTrainingStore) TrainingStepTimings(ctx context.Context, trainingID string) ([]db.TrainingStepLog, error) {
	if f.trainingStepTimingsFn == nil {
		return nil, nil
	}
	return f.trainingStepTimingsFn(ctx, trainingID)
}

func (f *fakeTrainingStore) RecordTraining(ctx context.Context, log db.TrainingLog, steps []db.TrainingStepLog) error {
	if f.recordTrainingFn == nil {
		return nil
	}
	return f.recordTrainingFn(ctx, log, steps)
}

func TestTrainingsHandlers(t *testing.T) {
	t.Run("Create training", func(t *testing.T) {
		store := &fakeTrainingStore{workoutWithStepsFn: func(context.Context, string) (*db.Workout, error) {
			return &db.Workout{
				ID:     "w1",
				UserID: "user@example.com",
				Name:   "Workout",
				Steps: []db.WorkoutStep{
					{ID: "s1", Type: "set", Name: "Step"},
				},
			}, nil
		}}
		api := &API{Trainings: trainings.New(store, sounds.URLByKey)}
		h := api.CreateTraining()
		body := strings.NewReader(`{"workoutId":"w1"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/trainings", body)
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		require.Equal(t, http.StatusCreated, rec.Code)
		var payload struct {
			TrainingID string `json:"trainingId"`
		}
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&payload))
		assert.NotEmpty(t, payload.TrainingID)
	})

	t.Run("List training history", func(t *testing.T) {
		store := &fakeTrainingStore{
			trainingHistoryFn: func(context.Context, string, int) ([]db.TrainingLog, error) {
				return []db.TrainingLog{{ID: "s1", WorkoutID: "w1", UserID: "user@example.com"}}, nil
			},
			trainingStepTimingsFn: func(context.Context, string) ([]db.TrainingStepLog, error) {
				return []db.TrainingStepLog{{ID: "s1-0", TrainingID: "s1", StepOrder: 0}}, nil
			},
		}
		api := &API{Trainings: trainings.New(store, sounds.URLByKey)}
		h := api.ListTrainingHistory()
		req := httptest.NewRequest(http.MethodGet, "/api/users/user@example.com/trainings", nil)
		req.SetPathValue("id", "user@example.com")
		req.Header.Set("X-User-ID", "user@example.com")
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		var payload []map[string]any
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&payload))
		require.Len(t, payload, 1)
	})

	t.Run("Training steps", func(t *testing.T) {
		store := &fakeTrainingStore{trainingStepTimingsFn: func(context.Context, string) ([]db.TrainingStepLog, error) {
			return []db.TrainingStepLog{{ID: "s1-0", TrainingID: "s1", StepOrder: 0}}, nil
		}}
		api := &API{Trainings: trainings.New(store, sounds.URLByKey)}
		h := api.TrainingSteps()
		req := httptest.NewRequest(http.MethodGet, "/api/trainings/s1/steps", nil)
		req.SetPathValue("id", "s1")
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		var payload []db.TrainingStepLog
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&payload))
		require.Len(t, payload, 1)
		assert.Equal(t, "s1-0", payload[0].ID)
	})

	t.Run("Complete training", func(t *testing.T) {
		store := &fakeTrainingStore{recordTrainingFn: func(context.Context, db.TrainingLog, []db.TrainingStepLog) error { return nil }}
		api := &API{Trainings: trainings.New(store, sounds.URLByKey)}
		h := api.CompleteTraining()
		body := strings.NewReader(`{"trainingId":"s1","workoutId":"w1","workoutName":"Workout","userId":"user@example.com","startedAt":"2024-01-01T00:00:00Z","completedAt":"2024-01-01T00:00:10Z","steps":[{"id":"s1","name":"Step","type":"set","elapsedMillis":1000}]}`)
		req := httptest.NewRequest(http.MethodPost, "/api/trainings/complete", body)
		req.Header.Set("X-User-ID", "user@example.com")
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		require.Equal(t, http.StatusCreated, rec.Code)
		var payload db.TrainingLog
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&payload))
		assert.Equal(t, "s1", payload.ID)
		assert.WithinDuration(t, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), payload.StartedAt, time.Second)
	})
}
