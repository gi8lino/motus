package users

// Service provides user operations for handlers.
type Service struct {
	store             Store
	authHeader        string
	allowRegistration bool
}

// New creates a new users service.
func New(store Store, authHeader string, allowRegistration bool) *Service {
	return &Service{store: store, authHeader: authHeader, allowRegistration: allowRegistration}
}
