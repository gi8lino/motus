package app

import (
	"context"
	"embed"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/containeroo/httpgrace/server"
	"github.com/containeroo/tinyflags"

	"github.com/gi8lino/motus/internal/bootstrap"
	"github.com/gi8lino/motus/internal/db"
	"github.com/gi8lino/motus/internal/flag"
	"github.com/gi8lino/motus/internal/handler"
	"github.com/gi8lino/motus/internal/logging"
	"github.com/gi8lino/motus/internal/routes"
)

// Run is the entry point for the application lifecycle.
func Run(
	ctx context.Context,
	assets embed.FS,
	version, commit string,
	args []string,
	w io.Writer,
) error {
	// Create a cancellable context that listens for termination signals.
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Create another context to listen for reload signals
	reloadCh := make(chan os.Signal, 1)
	signal.Notify(reloadCh, syscall.SIGHUP)

	// Parse CLI flags and handle help/version requests.
	opts, err := flag.ParseFlags(args, version)
	// Setup logger immediately so startup errors are correctly logged.
	logger := logging.SetupLogger(opts.LogFormat, opts.Debug, w)
	sysLogger := logging.SystemLogger(logger, nil)
	if err != nil {
		if tinyflags.IsHelpRequested(err) || tinyflags.IsVersionRequested(err) {
			fmt.Fprint(w, err.Error()) // nolint:errcheck
			return nil
		}
		sysLogger.Error(
			"application failed",
			"event", "app_failed",
			"stage", "parse_flags",
			"err", err,
		)
		return fmt.Errorf("CLI flags error: %w", err)
	}

	// Configure the logger early so startup errors are visible.
	sysLogger.Info(
		"starting Motus",
		"event", "app_starting",
		"version", version,
		"commit", commit,
	)

	// Record any CLI overrides to aid debugging.
	if len(opts.OverriddenValues) > 0 {
		sysLogger.Info(
			"CLI Overrides",
			"event", "cli_overrides",
			"overrides", opts.OverriddenValues,
		)
	}

	// Connect to the database.
	store, err := db.New(ctx, opts.DatabaseURL)
	if err != nil {
		sysLogger.Error(
			"application failed",
			"event", "app_failed",
			"stage", "connect_db",
			"err", err,
		)
		return fmt.Errorf("connect db: %w", err)
	}
	defer store.Close()

	// Ensure database schema is up to date before serving requests.
	if err := store.EnsureSchema(ctx, sysLogger); err != nil {
		sysLogger.Error(
			"application failed",
			"event", "app_failed",
			"stage", "ensure_schema",
			"err", err,
		)
		return fmt.Errorf("ensure schema: %w", err)
	}

	// Bootstrap an admin user if credentials were configured.
	if err := bootstrap.EnsureAdminUser(ctx, store, logger, opts.AdminEmail, opts.AdminPassword); err != nil {
		sysLogger.Error(
			"application failed",
			"event", "app_failed",
			"stage", "ensure_admin_user",
			"err", err,
		)
		return fmt.Errorf("ensure admin user: %w", err)
	}

	// Load extra core exercises if the CLI flag was set.
	if opts.CoreExercisesFile != "" {
		if err := bootstrap.SeedCoreExercises(ctx, store, sysLogger, opts.CoreExercisesFile); err != nil {
			sysLogger.Error(
				"application failed",
				"event", "app_failed",
				"stage", "seed_core_exercises",
				"err", err,
			)
			return fmt.Errorf("load core exercises: %w", err)
		}
	}

	// Build the API handler with runtime configuration.
	api := handler.NewAPI(
		store,
		logger,
		opts.AuthHeader,
		opts.SiteRoot,
		version,
		commit,
		opts.AllowRegistration,
		opts.AutoCreateUsers,
	)

	// Configure the HTTP router and SPA asset handler.
	router, err := routes.NewRouter(assets, opts.RoutePrefix, sysLogger, api, opts.Debug)
	if err != nil {
		sysLogger.Error(
			"application failed",
			"event", "app_failed",
			"stage", "configure_router",
			"err", err,
		)
		return fmt.Errorf("configure router: %w", err)
	}

	// Start the HTTP server and block until shutdown.
	if err := server.Run(ctx, opts.ListenAddr, router, sysLogger); err != nil {
		sysLogger.Error(
			"application failed",
			"event", "app_failed",
			"stage", "run_server",
			"err", err,
		)
		return fmt.Errorf("run server: %w", err)
	}

	return nil
}
