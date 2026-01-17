package exercises

// Service coordinates exercise catalog operations for handlers.
type Service struct {
	store Store
}

// New creates a new exercises service.
func New(store Store) *Service {
	return &Service{store: store}
}
