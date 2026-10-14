package bootstrap

import (
	"io/fs"
	"log/slog"

	"github.com/gi8lino/motus/internal/flags"
	httpserver "github.com/gi8lino/motus/internal/http/server"
)

// NewHTTPConfig constructs the passive HTTP adapter from the completed application graph.
func NewHTTPConfig(
	assets fs.FS,
	cfg flags.Config,
	infrastructure *Infrastructure,
	application *Application,
	logger *slog.Logger,
	version, commit string,
) httpserver.Config {
	return httpserver.Config{
		InfrastructureConfig: httpserver.InfrastructureConfig{
			RoutePrefix: cfg.RoutePrefix,
			Assets:      assets,
			Logger:      logger,
			AccessLog:   cfg.AccessLog,
			HealthStore: infrastructure.store,
			Version:     version,
			Commit:      commit,
		},
		AuthenticationConfig: httpserver.AuthenticationConfig{
			Browser:           application.browser,
			AllowRegistration: cfg.AllowRegistration,
		},
		ApplicationConfig: httpserver.ApplicationConfig{
			Authentication: application.authentication,
			Users:          application.users,
			Exercises:      application.exercises,
			Workouts:       application.workouts,
			Trainings:      application.trainings,
		},
	}
}
