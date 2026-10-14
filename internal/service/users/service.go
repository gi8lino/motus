package users

// Service provides user operations for handlers.
type Service struct {
	store               Store
	registrationAllowed func() bool
}

// New creates a new users service.
func New(store Store, registrationAllowed func() bool) *Service {
	return &Service{store: store, registrationAllowed: registrationAllowed}
}

// AllowRegistration reports whether self-registration is currently permitted.
func (s *Service) AllowRegistration() bool {
	return s.registrationAllowed != nil && s.registrationAllowed()
}
