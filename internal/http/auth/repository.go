package httpauth

import (
	"context"
	"time"

	"github.com/gi8lino/motus/internal/db"
)

type localRepository interface {
	CreateSession(context.Context, string, string, time.Time) error
	CreateSessionHash(context.Context, string, string, time.Time) error
	GetSessionUserByHash(context.Context, string, time.Time) (*db.User, error)
	DeleteSessionHash(context.Context, string) error
	GetSessionUser(context.Context, string, time.Time) (*db.User, error)
	DeleteSession(context.Context, string) error
	DeleteUserSessions(context.Context, string) error
	GetUserWithPassword(context.Context, string) (*db.User, string, error)
	UpdateUserPassword(context.Context, string, string) error
	CreateInitialAdministrator(context.Context, string, string) (*db.User, error)
}

type noneRepository interface {
	EnsureAdministrator(context.Context) (*db.User, error)
}
type identityRepository interface {
	GetIdentityUser(context.Context, string, string) (*db.User, error)
	ResolveExternalIdentity(context.Context, string, string, string, string, bool) (*db.User, error)
}

// externalLoginService resolves a verified external identity through application policy.
type oidcLoginService interface {
	Login(context.Context, string, string, string, string, string) (*db.User, error)
}
type trustedProxyLoginService interface {
	Login(context.Context, string, string, string, bool, bool) (*db.User, error)
}
type settingsRepository interface {
	AuthenticationSettings(context.Context) (db.AuthenticationSettings, error)
}

// browserRepository is composition-only wiring; handlers never receive it.
type browserRepository interface {
	localRepository
	noneRepository
	identityRepository
	settingsRepository
}
