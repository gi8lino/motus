package endpoint

import (
	"github.com/gi8lino/motus/internal/service/exercises"
	"github.com/gi8lino/motus/internal/service/trainings"
	"github.com/gi8lino/motus/internal/service/users"
	"github.com/gi8lino/motus/internal/service/workouts"
	"log/slog"
)

// testEndpoints is a test fixture only. Production handlers receive explicit inputs.
type testEndpoints struct {
	Version, Commit   string
	Users             *users.Service
	Exercises         *exercises.Service
	Workouts          *workouts.Service
	Trainings         *trainings.Service
	Logger            *slog.Logger
	AuthHeader        string
	AllowRegistration bool
	OIDCEnabled       bool
}
