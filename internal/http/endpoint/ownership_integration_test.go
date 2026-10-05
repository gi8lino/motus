package endpoint

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gi8lino/motus/internal/db"
	httpauth "github.com/gi8lino/motus/internal/http/auth"
	"github.com/gi8lino/motus/internal/service/workouts"
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

	api := &testEndpoints{Workouts: workouts.New(store), Logger: logger}
	req := httptest.NewRequest(http.MethodGet, "/api/workouts/"+workout.ID, nil)
	req = httpauth.WithUser(req, *other)
	req.SetPathValue("id", workout.ID)
	rec := httptest.NewRecorder()

	GetWorkout(api.Workouts, api.Logger).ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}
