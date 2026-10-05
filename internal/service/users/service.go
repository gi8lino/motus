package users

// Service provides user operations for handlers.
type Service struct {
	store             Store
	allowRegistration bool
}

// New creates a new users service.
func New(store Store, allowRegistration bool) *Service {
	return &Service{store: store, allowRegistration: allowRegistration}
}
