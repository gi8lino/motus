package trainings

// Service provides training session orchestration for handlers.
type Service struct {
	store         Store
	soundURLByKey func(string) string
}

// New creates a new trainings service.
func New(store Store, soundURLByKey func(string) string) *Service {
	return &Service{store: store, soundURLByKey: soundURLByKey}
}
