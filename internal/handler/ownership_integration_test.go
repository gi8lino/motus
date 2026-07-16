package handler

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gi8lino/motus/internal/auth"
	"github.com/gi8lino/motus/internal/db"
	"github.com/gi8lino/motus/internal/utils"
)

func TestPostgresRejectsCrossUserWorkoutAccess(t *testing.T) {
	databaseURL := os.Getenv("MOTUS_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set MOTUS_TEST_DATABASE_URL to run PostgreSQL authorization integration tests")
	}
	ctx := context.Background()
	store, err := db.New(ctx, databaseURL)
	require.NoError(t, err)
	t.Cleanup(store.Close)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	require.NoError(t, store.EnsureSchema(ctx, logger))

	suffix := utils.NewID()
	owner, err := store.CreateUser(ctx, "owner-"+suffix+"@example.com", "", "hash")
	require.NoError(t, err)
	other, err := store.CreateUser(ctx, "other-"+suffix+"@example.com", "", "hash")
	require.NoError(t, err)
	workout, err := store.CreateWorkout(ctx, &db.Workout{UserID: owner.ID, Name: "Private"})
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.DeleteWorkout(context.Background(), workout.ID) })

	token := utils.NewID()
	require.NoError(t, store.CreateSession(ctx, token, other.ID, time.Now().Add(time.Hour)))
	api := NewAPI(store, logger, "", "", "test", "test", false, false, time.Hour)
	req := httptest.NewRequest(http.MethodGet, "/api/workouts/"+workout.ID, nil)
	req.SetPathValue("id", workout.ID)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: token})
	rec := httptest.NewRecorder()

	api.GetWorkout().ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}
