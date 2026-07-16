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
)

// exerciseStore is an interface for storing exercises.
type exerciseStore interface {
	CreateExercise(ctx context.Context, name, ownerUserID string, isCore bool) (*db.Exercise, error)
	SetExerciseHasSides(ctx context.Context, id string, hasSides bool) (*db.Exercise, error)
}

// coreExercisesFile mirrors the YAML layout expected in the seed file.
type coreExercisesFile struct {
	Exercises []coreExercise `yaml:"exercises"`
}

type coreExercise struct {
	Name     string `yaml:"name"`
	HasSides bool   `yaml:"hasSides"`
}

func (e *coreExercise) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode {
		e.Name = node.Value
		return nil
	}
	type plain coreExercise
	return node.Decode((*plain)(e))
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
		created, err := store.CreateExercise(ctx, item.Name, "", true)
		if err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
				continue
			}
			return fmt.Errorf("create core exercise %q: %w", item.Name, err)
		}
		if item.HasSides {
			if _, err := store.SetExerciseHasSides(ctx, created.ID, true); err != nil {
				return fmt.Errorf("mark core exercise %q unilateral: %w", item.Name, err)
			}
		}
		logger.Info(
			"seeded core exercise",
			"event", "bootstrap_exercise_seeded",
			"resource", "exercise",
			"name", item.Name,
		)
	}

	return nil
}
