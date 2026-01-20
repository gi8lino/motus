package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gi8lino/motus/internal/logging"
	"golang.org/x/sync/errgroup"
)

// Run sets up and manages the reverse proxy HTTP server.
func Run(ctx context.Context, listenAddr string, router http.Handler, logger *slog.Logger) error {
	// Create server with sensible timeouts.
	server := &http.Server{
		Addr:              listenAddr,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// errgroup provides error propagation and shared cancellation for goroutines.
	eg, groupCtx := errgroup.WithContext(ctx)

	// Start the server.
	eg.Go(func() error {
		logging.SystemLogger(logger, nil).Info(
			"starting server",
			"event", "server_starting",
			"listen_addr", server.Addr,
		)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	})

	// Wait for the server to stop gracefully.
	eg.Go(func() error {
		select {
		// Shutdown the server when the context is canceled.
		case <-ctx.Done():
			logging.SystemLogger(logger, nil).Info(
				"shutting down server",
				"event", "server_shutting_down",
			)

			// Use a bounded timeout to finish in-flight requests.
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if err := server.Shutdown(shutdownCtx); err != nil {
				return err
			}
			return nil

		// Shutdown the server when the group context is canceled.
		case <-groupCtx.Done():
			return nil
		}
	})

	return eg.Wait()
}
