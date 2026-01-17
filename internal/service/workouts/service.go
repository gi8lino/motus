package workouts

// Service coordinates workout operations for handlers.
type Service struct {
	store Store
}

// New creates a new workouts service.
func New(store Store) *Service {
	return &Service{store: store}
}
