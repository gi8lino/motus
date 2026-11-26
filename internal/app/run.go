package app

import (
	"context"
	"fmt"
	"io"
	"io/fs"

	appkithttpserver "github.com/containeroo/appkit/httpserver"
	appkitlogging "github.com/containeroo/appkit/logging"
	server "github.com/containeroo/httpkit/grace"
	"github.com/containeroo/tinyflags"

	"github.com/gi8lino/motus/internal/bootstrap"
	"github.com/gi8lino/motus/internal/cli"
	httpserver "github.com/gi8lino/motus/internal/http/server"
)

// Run composes and runs the Motus process.
func Run(
	ctx context.Context,
	args []string,
	assets fs.FS,
	version, commit string,
	stdout, stderr io.Writer,
) error {
	// Parse deployment configuration before constructing runtime dependencies.
	flags, err := cli.Parse(args, version)
	if err != nil {
		if tinyflags.IsHelpRequested(err) || tinyflags.IsVersionRequested(err) {
			fmt.Fprint(stdout, err.Error()) // nolint:errcheck
			return nil
		}
		fmt.Fprintln(stderr, err) // nolint:errcheck
		return err
	}

	// Configure process logging and record the effective application identity.
	logger := appkitlogging.Setup(flags.Logging.Format, flags.Logging.Debug, stdout)
	setupLogger := logger.With("component", "setup")
	setupLogger.Info(
		"starting Motus",
		"event", "app_starting",
		"version", version,
		"commit", commit,
	)
	if len(flags.Overrides) > 0 {
		setupLogger.Info("CLI Overrides", "event", "cli_overrides", "overrides", flags.Overrides.Values())
	}

	// Bind the process lifetime to operating-system shutdown signals.
	ctx, stop := server.SignalContext(ctx)
	defer stop()

	// Construct process infrastructure and own its lifetime here.
	infrastructure, err := bootstrap.NewInfrastructure(ctx, flags, logger)
	if err != nil {
		setupLogger.Error("initialize infrastructure", "event", "infrastructure_init_failed", "error", err)
		return err
	}
	defer infrastructure.Close()

	// Compose application capabilities around the initialized infrastructure.
	application, err := bootstrap.NewApplication(ctx, flags, infrastructure, logger.With("component", "bootstrap"))
	if err != nil {
		setupLogger.Error("initialize application", "event", "application_init_failed", "error", err)
		return err
	}

	// Build the passive HTTP adapter after the application graph is complete.
	serverConfig := bootstrap.NewHTTPConfig(assets, flags, infrastructure, application, logger.With("component", "server"), version, commit)

	// Construct both HTTP handlers and run their listeners until shutdown.
	handlers := httpserver.New(serverConfig)
	if err := appkithttpserver.Run(
		ctx,
		flags.Server.ListenAddress,
		flags.Server.ManagementListenAddress,
		handlers,
		logger.With("server", "application"),
		logger.With("server", "management"),
	); err != nil {
		setupLogger.Error("run server", "event", "server_run_failed", "error", err)
		return err
	}

	return nil
}
