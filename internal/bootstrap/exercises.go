package bootstrap

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/gi8lino/motus/internal/db"
	"github.com/gi8lino/motus/internal/logging"
)

// exerciseStore is an interface for storing exercises.
type exerciseStore interface {
	CreateExercise(ctx context.Context, name, ownerUserID string, isCore bool) (*db.Exercise, error)
}

// coreExercisesFile mirrors the YAML layout expected in the seed file.
type coreExercisesFile struct {
	Exercises []string `yaml:"exercises"`
}

// SeedCoreExercises loads core exercises from a YAML file and inserts them if they don't exist.
func SeedCoreExercises(ctx context.Context, store exerciseStore, logger *slog.Logger, filePath string) error {
	if filePath == "" {
		return nil
	}
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open core exercises file: %w", err)
	}
	defer f.Close() // nolint:errcheck

	data, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("read core exercises file: %w", err)
	}

	var exercises coreExercisesFile
	if err := yaml.Unmarshal(data, &exercises); err != nil {
		return err
	}

	for _, name := range exercises.Exercises {
		if _, err := store.CreateExercise(ctx, name, "", true); err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
				continue
			}
			return fmt.Errorf("create core exercise %q: %w", name, err)
		}
		logging.SystemLogger(logger, ctx).Info(
			"seeded core exercise",
			"event", "bootstrap_exercise_seeded",
			"resource", "exercise",
			"name", name,
		)
	}

	return nil
}
