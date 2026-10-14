package httpauth

import (
	"net/http"

	"github.com/gi8lino/motus/internal/db"
)

var _ Authenticator = (*None)(nil)

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
