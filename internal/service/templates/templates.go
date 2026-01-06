package templates

import (
	"context"
	"strings"

	"github.com/gi8lino/motus/internal/db"
	"github.com/gi8lino/motus/internal/service"
)

// store defines the persistence methods needed by the templates service.
type store interface {
	ListTemplates(ctx context.Context) ([]db.Workout, error)
	CreateTemplateFromWorkout(ctx context.Context, workoutID, name string) (*db.Workout, error)
	WorkoutWithSteps(ctx context.Context, id string) (*db.Workout, error)
	CreateWorkoutFromTemplate(ctx context.Context, templateID, userID, name string) (*db.Workout, error)
}

// Service coordinates template operations.
type Service struct {
	Store store
}

// New creates a new templates service.
func New(store store) *Service {
	return &Service{Store: store}
}

// List returns all shared templates.
func (s *Service) List(ctx context.Context) ([]db.Workout, error) {
	templates, err := s.Store.ListTemplates(ctx)
	if err != nil {
		return nil, service.NewError(service.ErrorInternal, err.Error())
	}
	return templates, nil
}

// Create marks a workout as a template.
func (s *Service) Create(ctx context.Context, workoutID, name string) (*db.Workout, error) {
	workoutID = strings.TrimSpace(workoutID)
	if workoutID == "" {
		return nil, service.NewError(service.ErrorValidation, "workoutId is required")
	}
	template, err := s.Store.CreateTemplateFromWorkout(ctx, workoutID, name)
	if err != nil {
		return nil, service.NewError(service.ErrorValidation, err.Error())
	}
	return template, nil
}

// Get returns a template by id.
func (s *Service) Get(ctx context.Context, id string) (*db.Workout, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, service.NewError(service.ErrorValidation, "template id is required")
	}
	template, err := s.Store.WorkoutWithSteps(ctx, id)
	if err != nil || !template.IsTemplate {
		if err == nil {
			return nil, service.NewError(service.ErrorNotFound, "template not found")
		}
		return nil, service.NewError(service.ErrorNotFound, err.Error())
	}
	return template, nil
}

// Apply clones a template into a new workout.
func (s *Service) Apply(ctx context.Context, templateID, userID, name string) (*db.Workout, error) {
	templateID = strings.TrimSpace(templateID)
	if templateID == "" {
		return nil, service.NewError(service.ErrorValidation, "template id is required")
	}
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, service.NewError(service.ErrorValidation, "userId is required")
	}
	name = strings.TrimSpace(name)
	workout, err := s.Store.CreateWorkoutFromTemplate(ctx, templateID, userID, name)
	if err != nil {
		return nil, service.NewError(service.ErrorValidation, err.Error())
	}
	return workout, nil
}
