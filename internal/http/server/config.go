package server

import (
	"io/fs"
	"log/slog"

	appauth "github.com/gi8lino/motus/internal/application/authentication"
	appsystem "github.com/gi8lino/motus/internal/application/system"
	"github.com/gi8lino/motus/internal/db"
	httpauth "github.com/gi8lino/motus/internal/http/auth"
	"github.com/gi8lino/motus/internal/service/exercises"
	"github.com/gi8lino/motus/internal/service/trainings"
	"github.com/gi8lino/motus/internal/service/users"
	"github.com/gi8lino/motus/internal/service/workouts"
)

// InfrastructureConfig contains the HTTP transport and presentation dependencies.
type InfrastructureConfig struct {
	RoutePrefix     string
	Assets          fs.FS
	Logger          *slog.Logger
	AccessLog       bool
	HealthStore     db.HealthChecker
	Version, Commit string
}

// AuthenticationConfig contains the configured browser authentication boundary.
type AuthenticationConfig struct {
	Browser httpauth.BrowserAuth
}

// ApplicationConfig contains Motus use cases exposed by HTTP routes.
type ApplicationConfig struct {
	Authentication *appauth.Configuration
	System         *appsystem.System
	Users          *users.Service
	Exercises      *exercises.Service
	Workouts       *workouts.Service
	Trainings      *trainings.Service
}

// Config groups capabilities used to construct HTTP routes.
type Config struct {
	InfrastructureConfig
	AuthenticationConfig
	ApplicationConfig
}
