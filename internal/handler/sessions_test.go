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
	"github.com/gi8lino/motus/internal/service/sessions"
	"github.com/gi8lino/motus/internal/service/sounds"
)

type fakeSessionStore struct {
	workoutWithStepsFn   func(context.Context, string) (*db.Workout, error)
	sessionHistoryFn     func(context.Context, string, int) ([]db.SessionLog, error)
	sessionStepTimingsFn func(context.Context, string) ([]db.SessionStepLog, error)
	recordSessionFn      func(context.Context, db.SessionLog, []db.SessionStepLog) error
}

func (f *fakeSessionStore) WorkoutWithSteps(ctx context.Context, id string) (*db.Workout, error) {
	if f.workoutWithStepsFn == nil {
		return nil, nil
	}
	return f.workoutWithStepsFn(ctx, id)
}

func (f *fakeSessionStore) SessionHistory(ctx context.Context, userID string, limit int) ([]db.SessionLog, error) {
	if f.sessionHistoryFn == nil {
		return nil, nil
	}
	return f.sessionHistoryFn(ctx, userID, limit)
}

func (f *fakeSessionStore) SessionStepTimings(ctx context.Context, sessionID string) ([]db.SessionStepLog, error) {
	if f.sessionStepTimingsFn == nil {
		return nil, nil
	}
	return f.sessionStepTimingsFn(ctx, sessionID)
}

func (f *fakeSessionStore) RecordSession(ctx context.Context, log db.SessionLog, steps []db.SessionStepLog) error {
	if f.recordSessionFn == nil {
		return nil
	}
	return f.recordSessionFn(ctx, log, steps)
}

func TestSessionsHandlers(t *testing.T) {
	t.Run("Create session", func(t *testing.T) {
		store := &fakeSessionStore{workoutWithStepsFn: func(context.Context, string) (*db.Workout, error) {
			return &db.Workout{
				ID:     "w1",
				UserID: "user@example.com",
				Name:   "Workout",
				Steps: []db.WorkoutStep{
					{ID: "s1", Type: "set", Name: "Step"},
				},
			}, nil
		}}
		api := &API{Sessions: sessions.New(store, sounds.URLByKey)}
		h := api.CreateSession()
		body := strings.NewReader(`{"workoutId":"w1"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/sessions", body)
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		require.Equal(t, http.StatusCreated, rec.Code)
		var payload struct {
			SessionID string `json:"sessionId"`
		}
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&payload))
		assert.NotEmpty(t, payload.SessionID)
	})

	t.Run("List session history", func(t *testing.T) {
		store := &fakeSessionStore{
			sessionHistoryFn: func(context.Context, string, int) ([]db.SessionLog, error) {
				return []db.SessionLog{{ID: "s1", WorkoutID: "w1", UserID: "user@example.com"}}, nil
			},
			sessionStepTimingsFn: func(context.Context, string) ([]db.SessionStepLog, error) {
				return []db.SessionStepLog{{ID: "s1-0", SessionID: "s1", StepOrder: 0}}, nil
			},
		}
		api := &API{Sessions: sessions.New(store, sounds.URLByKey)}
		h := api.ListSessionHistory()
		req := httptest.NewRequest(http.MethodGet, "/api/users/user@example.com/sessions", nil)
		req.SetPathValue("id", "user@example.com")
		req.Header.Set("X-User-ID", "user@example.com")
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		var payload []map[string]any
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&payload))
		require.Len(t, payload, 1)
	})

	t.Run("Session steps", func(t *testing.T) {
		store := &fakeSessionStore{sessionStepTimingsFn: func(context.Context, string) ([]db.SessionStepLog, error) {
			return []db.SessionStepLog{{ID: "s1-0", SessionID: "s1", StepOrder: 0}}, nil
		}}
		api := &API{Sessions: sessions.New(store, sounds.URLByKey)}
		h := api.SessionSteps()
		req := httptest.NewRequest(http.MethodGet, "/api/sessions/s1/steps", nil)
		req.SetPathValue("id", "s1")
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		var payload []db.SessionStepLog
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&payload))
		require.Len(t, payload, 1)
		assert.Equal(t, "s1-0", payload[0].ID)
	})

	t.Run("Complete session", func(t *testing.T) {
		store := &fakeSessionStore{recordSessionFn: func(context.Context, db.SessionLog, []db.SessionStepLog) error { return nil }}
		api := &API{Sessions: sessions.New(store, sounds.URLByKey)}
		h := api.CompleteSession()
		body := strings.NewReader(`{"sessionId":"s1","workoutId":"w1","workoutName":"Workout","userId":"user@example.com","startedAt":"2024-01-01T00:00:00Z","completedAt":"2024-01-01T00:00:10Z","steps":[{"id":"s1","name":"Step","type":"set","elapsedMillis":1000}]}`)
		req := httptest.NewRequest(http.MethodPost, "/api/sessions/complete", body)
		req.Header.Set("X-User-ID", "user@example.com")
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		require.Equal(t, http.StatusCreated, rec.Code)
		var payload db.SessionLog
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&payload))
		assert.Equal(t, "s1", payload.ID)
		assert.WithinDuration(t, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), payload.StartedAt, time.Second)
	})
}
