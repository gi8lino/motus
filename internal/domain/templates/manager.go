package templates

import "context"

// Manager contains the template business rules.
type Manager struct {
	store Store
}

// NewManager builds a template Manager.
func NewManager(store Store) *Manager {
	return &Manager{store: store}
}

// Create records a workout as a template.
func (m *Manager) Create(ctx context.Context, workoutID, name string) (*Workout, error) {
	wid, err := requireID(workoutID, "workoutId is required")
	if err != nil {
		return nil, err
	}

	template, err := m.store.CreateTemplateFromWorkout(ctx, wid, name)
	if err != nil {
		return nil, internal(err)
	}
	return template, nil
}

// List returns all available templates.
func (m *Manager) List(ctx context.Context) ([]Workout, error) {
	list, err := m.store.ListTemplates(ctx)
	if err != nil {
		return nil, internal(err)
	}
	return list, nil
}

// Get fetches a template by id.
func (m *Manager) Get(ctx context.Context, id string) (*Workout, error) {
	tid, err := requireID(id, "template id is required")
	if err != nil {
		return nil, err
	}

	template, err := m.store.WorkoutWithSteps(ctx, tid)
	if err != nil {
		return nil, internal(err)
	}
	if template == nil || !template.IsTemplate {
		return nil, notFound("template not found")
	}
	return template, nil
}

// Apply clones a template to a workout for the user.
func (m *Manager) Apply(ctx context.Context, templateID, userID, name string) (*Workout, error) {
	tid, err := requireID(templateID, "template id is required")
	if err != nil {
		return nil, err
	}
	uid, err := requireID(userID, "userId is required")
	if err != nil {
		return nil, err
	}

	workout, err := m.store.CreateWorkoutFromTemplate(ctx, tid, uid, name)
	if err != nil {
		return nil, internal(err)
	}
	return workout, nil
}
