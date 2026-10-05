package httpauth

import (
	"github.com/gi8lino/motus/internal/db"
	"net/http"
)

// None deliberately disables authentication for a deployment behind a trusted boundary.
type None struct{ repository noneRepository }

func NewNone(repository noneRepository) *None { return &None{repository: repository} }
func (a *None) Authenticate(r *http.Request) (db.User, error) {
	user, err := a.repository.EnsureAdministrator(r.Context())
	if err != nil || user == nil {
		return db.User{}, ErrUnauthenticated
	}
	return *user, nil
}

var _ Authenticator = (*None)(nil)
