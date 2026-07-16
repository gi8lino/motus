package bootstrap

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/gi8lino/motus/internal/db"
)

// exerciseStore is an interface for storing exercises.
type exerciseStore interface {
	UpsertCoreExercise(ctx context.Context, name string, hasSides bool) (*db.Exercise, bool, error)
}

// coreExercisesFile mirrors the YAML layout expected in the seed file.
type coreExercisesFile struct {
	Exercises []coreExercise `yaml:"exercises"`
}

type coreExercise struct {
	Name     string `yaml:"name"`
	HasSides bool   `yaml:"hasSides"`
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

	for _, item := range exercises.Exercises {
		_, created, err := store.UpsertCoreExercise(ctx, item.Name, item.HasSides)
		if err != nil {
			return fmt.Errorf("reconcile core exercise %q: %w", item.Name, err)
		}
		logger.Info(
			"reconciled core exercise",
			"event", "bootstrap_exercise_reconciled",
			"resource", "exercise",
			"name", item.Name,
			"created", created,
		)
	}

	return nil
}
