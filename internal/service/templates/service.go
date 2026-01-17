package templates

// Service provides access to workout template operations.
type Service struct {
	store Store
}

// New builds a template service.
func New(store Store) *Service {
	return &Service{store: store}
}
