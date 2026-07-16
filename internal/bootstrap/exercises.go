package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"gopkg.in/yaml.v3"

	corecatalog "github.com/gi8lino/motus/examples"
	"github.com/gi8lino/motus/internal/db"
)

// exerciseStore is an interface for storing exercises.
type exerciseStore interface {
	UpsertCoreExercise(ctx context.Context, name string, hasSides bool) (*db.Exercise, bool, error)
}

// coreExercisesFile mirrors the YAML layout expected in the seed file.
type coreExercisesFile struct {
	Version   int            `yaml:"version"`
	Exercises []coreExercise `yaml:"exercises"`
}

type coreExercise struct {
	Name     string `yaml:"name"`
	HasSides bool   `yaml:"hasSides"`
}

// SeedCoreExercises reconciles the embedded catalog, or a configured override file.
func SeedCoreExercises(ctx context.Context, store exerciseStore, logger *slog.Logger, filePath string) error {
	data := corecatalog.CoreExercisesYAML
	source := "embedded"
	if filePath != "" {
		var err error
		data, err = os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("read core exercises file: %w", err)
		}
		source = filePath
	}

	var exercises coreExercisesFile
	if err := yaml.Unmarshal(data, &exercises); err != nil {
		return fmt.Errorf("decode core exercises catalog: %w", err)
	}
	if exercises.Version != 1 {
		return fmt.Errorf("core exercises catalog %q has unsupported version %d", source, exercises.Version)
	}
	if len(exercises.Exercises) == 0 {
		return fmt.Errorf("core exercises catalog %q is empty", source)
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
	logger.Info(
		"reconciled core exercise catalog",
		"event", "bootstrap_exercise_catalog_reconciled",
		"source", source,
		"version", exercises.Version,
		"count", len(exercises.Exercises),
	)

	return nil
}
