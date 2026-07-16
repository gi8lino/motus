package app

import (
	"context"
	"embed"
	"fmt"
	"io"

	"github.com/gi8lino/motus/internal/bootstrap"
	"github.com/gi8lino/motus/internal/db"
	"github.com/gi8lino/motus/internal/flag"
	"github.com/gi8lino/motus/internal/handler"
	"github.com/gi8lino/motus/internal/logging"
	"github.com/gi8lino/motus/internal/routes"

	"github.com/containeroo/httpgrace/server"
	"github.com/containeroo/tinyflags"
)

// Run is the entry point for the application lifecycle.
func Run(
	ctx context.Context,
	assets embed.FS,
	version, commit string,
	args []string,
	stdOut, stdErr io.Writer,
) error {
	flags, err := flag.ParseFlags(args, version)
	if err != nil {
		if tinyflags.IsHelpRequested(err) || tinyflags.IsVersionRequested(err) {
			_, _ = fmt.Fprint(stdOut, err.Error())
			return nil
		}
		_, _ = fmt.Fprintln(stdErr, err)
		return err
	}

	logger := logging.SetupLogger(flags.LogFormat, flags.Debug, stdOut)
	setupLogger := logger.With("component", "setup")
	setupLogger.Info(
		"starting Motus",
		"event", "app_starting",
		"version", version,
		"commit", commit,
	)

	if len(flags.OverriddenValues) > 0 {
		setupLogger.Info(
			"CLI Overrides",
			"event", "cli_overrides",
			"overrides", flags.OverriddenValues,
		)
	}

	// Connect to the database.
	store, err := db.New(ctx, flags.DatabaseURL)
	if err != nil {
		setupLogger.Error(
			"application failed",
			"event", "app_failed",
			"stage", "connect_db",
			"error", err,
		)
		return fmt.Errorf("connect db: %w", err)
	}
	defer store.Close()

	// Ensure database schema is up to date before serving requests.
	if err := store.EnsureSchema(ctx, setupLogger); err != nil {
		setupLogger.Error(
			"application failed",
			"event", "app_failed",
			"stage", "ensure_schema",
			"error", err,
		)
		return fmt.Errorf("ensure schema: %w", err)
	}

	// Bootstrap an admin user if credentials were configured.
	if err := bootstrap.EnsureAdminUser(ctx, store, logger, flags.AdminEmail, flags.AdminPassword); err != nil {
		setupLogger.Error(
			"application failed",
			"event", "app_failed",
			"stage", "ensure_admin_user",
			"error", err,
		)
		return fmt.Errorf("ensure admin user: %w", err)
	}

	// Reconcile the built-in core catalog, or a configured override.
	if err := bootstrap.SeedCoreExercises(ctx, store, setupLogger, flags.CoreExercisesFile); err != nil {
		setupLogger.Error(
			"application failed",
			"event", "app_failed",
			"stage", "seed_core_exercises",
			"error", err,
		)
		return fmt.Errorf("load core exercises: %w", err)
	}

	// Build the API handler with runtime configuration.
	appLogger := logger.With("component", "server")
	api := handler.NewAPI(
		store,
		appLogger,
		flags.AuthHeader,
		flags.SiteRoot,
		version,
		commit,
		flags.AllowRegistration,
		flags.AutoCreateUsers,
		flags.SessionLifetime,
	)

	// Configure the HTTP router and SPA asset handler.
	router, err := routes.NewRouter(assets, flags.RoutePrefix, setupLogger, api, flags.Debug)
	if err != nil {
		setupLogger.Error(
			"application failed",
			"event", "app_failed",
			"stage", "create_router",
			"error", err,
		)
		return fmt.Errorf("configure router: %w", err)
	}

	ctx, stop := server.SignalContext(ctx)
	defer stop()

	// Start the HTTP server and block until shutdown.
	if err := server.Run(ctx, flags.ListenAddr, router, appLogger); err != nil {
		setupLogger.Error(
			"application failed",
			"event", "app_failed",
			"stage", "run_server",
			"error", err,
		)
		return fmt.Errorf("run server: %w", err)
	}

	return nil
}
