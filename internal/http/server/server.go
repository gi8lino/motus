// Package server provides Motus's passive application and management HTTP adapters.
package server

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/containeroo/httpkit/grace"
	httpkitserver "github.com/containeroo/httpkit/server"
)

// Handlers contains the application and operational management handlers.
type Handlers struct {
	// Application serves public application traffic.
	Application http.Handler
	// Management serves operational health and metrics traffic.
	Management http.Handler
}

// New constructs both HTTP applications with the same deployment prefix.
func New(config Config) Handlers {
	return Handlers{
		Application: newApplication(config),
		Management:  newManagement(config),
	}
}

// Run serves both handlers and stops the remaining listener when either exits.
func Run(
	ctx context.Context,
	appAddress string,
	managementAddress string,
	handlers Handlers,
	logger *slog.Logger,
) error {
	return httpkitserver.Run(ctx,
		httpkitserver.Listener{
			Address: appAddress,
			Handler: handlers.Application,
			Logger:  logger.With("server", "application"),
			Options: []grace.Option{grace.WithMaxHeaderValueCount(100)},
		},
		httpkitserver.Listener{
			Address: managementAddress,
			Handler: handlers.Management,
			Logger:  logger.With("server", "management"),
			Options: []grace.Option{grace.WithMaxHeaderValueCount(100)},
		},
	)
}
