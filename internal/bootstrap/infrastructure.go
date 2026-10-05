// Package bootstrap composes Motus process-level runtime dependencies.
package bootstrap

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gi8lino/motus/internal/db"
	"github.com/gi8lino/motus/internal/flags"
)

// Infrastructure owns process-level dependencies shared by application services.
type Infrastructure struct {
	// store owns the PostgreSQL-backed persistence store for the process.
	store *db.Store
}

// NewInfrastructure initializes process-level dependencies in startup order.
func NewInfrastructure(ctx context.Context, cfg flags.Config, logger *slog.Logger) (*Infrastructure, error) {
	store, err := db.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect db: %w", err)
	}
	if err := store.EnsureSchema(ctx, logger); err != nil {
		store.Close()
		return nil, fmt.Errorf("ensure schema: %w", err)
	}

	return &Infrastructure{store: store}, nil
}

// Close releases infrastructure owned by the process.
func (i *Infrastructure) Close() {
	if i == nil || i.store == nil {
		return
	}
	i.store.Close()
}
