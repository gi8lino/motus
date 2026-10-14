package httpauth

import (
	"context"
	"time"

	"github.com/gi8lino/motus/internal/db"
)

type localRepository interface {
	CreateSession(context.Context, string, string, time.Time) error
	GetSessionUser(context.Context, string, time.Time) (*db.User, error)
	DeleteSession(context.Context, string) error
	DeleteUserSessions(context.Context, string) error
	GetUserWithPassword(context.Context, string) (*db.User, string, error)
	UpdateUserPassword(context.Context, string, string) error
	SetupRequired(context.Context) (bool, error)
	CreateInitialAdministrator(context.Context, string, string) (*db.User, error)
}

type noneRepository interface {
	EnsureAdministrator(context.Context) (*db.User, error)
}
type identityRepository interface {
	ResolveExternalIdentity(context.Context, string, string, string, string, bool) (*db.User, error)
}

// browserRepository is composition-only wiring; handlers never receive it.
type browserRepository interface {
	localRepository
	noneRepository
	identityRepository
}
