package bootstrap

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gi8lino/motus/internal/db"
)

type fakeExerciseStore struct {
	exercises map[string]bool
}

func (f *fakeExerciseStore) UpsertCoreExercise(_ context.Context, name string, hasSides bool) (*db.Exercise, bool, error) {
	if f.exercises == nil {
		f.exercises = make(map[string]bool)
	}
	_, existed := f.exercises[name]
	f.exercises[name] = hasSides
	return &db.Exercise{Name: name, HasSides: hasSides, IsCore: true}, !existed, nil
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestSeedCoreExercisesUsesEmbeddedCatalogByDefault(t *testing.T) {
	t.Parallel()

	store := &fakeExerciseStore{}
	err := SeedCoreExercises(context.Background(), store, discardLogger(), "")
	require.NoError(t, err)
	assert.Greater(t, len(store.exercises), 50)
	assert.Contains(t, store.exercises, "Push-up")
	assert.True(t, store.exercises["Side Plank"])
	assert.False(t, store.exercises["Squat"])
}

func TestSeedCoreExercisesUsesConfiguredOverride(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "exercises.yaml")
	require.NoError(t, os.WriteFile(path, []byte("version: 1\nexercises:\n  - { name: Custom Carry, hasSides: true }\n"), 0o600))

	store := &fakeExerciseStore{}
	err := SeedCoreExercises(context.Background(), store, discardLogger(), path)
	require.NoError(t, err)
	assert.Equal(t, map[string]bool{"Custom Carry": true}, store.exercises)
}

func TestSeedCoreExercisesRejectsEmptyOverride(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "exercises.yaml")
	require.NoError(t, os.WriteFile(path, []byte("version: 1\nexercises: []\n"), 0o600))

	err := SeedCoreExercises(context.Background(), &fakeExerciseStore{}, discardLogger(), path)
	require.ErrorContains(t, err, "is empty")
}

func TestSeedCoreExercisesRejectsUnknownVersion(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "exercises.yaml")
	require.NoError(t, os.WriteFile(path, []byte("version: 2\nexercises:\n  - { name: Future Move }\n"), 0o600))

	err := SeedCoreExercises(context.Background(), &fakeExerciseStore{}, discardLogger(), path)
	require.ErrorContains(t, err, "unsupported version 2")
}
