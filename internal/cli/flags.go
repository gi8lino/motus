// Package flags parses deployment-level configuration for the Motus server.
package cli

import (
	"time"

	appkitconfig "github.com/containeroo/appkit/config"
	"github.com/containeroo/tinyflags"
)

const (
	sectionServer             = "Server"
	sectionDatabase           = "Database"
	sectionAuthentication     = "Authentication"
	sectionOIDC               = "OIDC"
	sectionCatalog            = "Exercise catalog"
	sectionLoggingDiagnostics = "Logging and diagnostics"
)

// Flags contains deployment-level runtime configuration for Motus.
type Flags struct {
	// Server contains shared HTTP listener configuration.
	Server appkitconfig.Server
	// Database contains shared PostgreSQL configuration.
	Database appkitconfig.Postgres
	// Authentication contains shared authentication configuration.
	Authentication appkitconfig.Authentication
	// Proxy contains shared trusted-proxy configuration.
	Proxy appkitconfig.Proxy
	// OIDC contains shared OpenID Connect configuration.
	OIDC appkitconfig.OIDC
	// Logging contains shared logging configuration.
	Logging appkitconfig.Logging

	CoreExercisesFile string
	// Overrides records configuration values explicitly overridden by flags or environment variables.
	Overrides tinyflags.Overrides
}

// Parse parses command-line arguments into application configuration.
func Parse(args []string, version string) (flags Flags, err error) {
	tf := appkitconfig.NewFlagSet("motus", "MOTUS", version)
	tf.SectionOrder(
		sectionServer,
		sectionDatabase,
		sectionAuthentication,
		appkitconfig.SectionProxy,
		sectionOIDC,
		sectionCatalog,
		sectionLoggingDiagnostics,
	)

	finalizeServer := appkitconfig.RegisterServer(tf, &flags.Server, appkitconfig.ServerOptions{})

	appkitconfig.RegisterPostgres(tf, &flags.Database, appkitconfig.PostgresOptions{})
	finalizeAuthentication := appkitconfig.RegisterAuthentication(
		tf,
		&flags.Authentication,
		appkitconfig.AuthenticationOptions{RegisterSessionLifetime: true, DefaultSessionLifetime: 30 * 24 * time.Hour},
	)

	appkitconfig.RegisterProxy(tf, &flags.Proxy)

	appkitconfig.RegisterOIDC(tf, &flags.OIDC)
	finalizeLogging := appkitconfig.RegisterLogging(tf, &flags.Logging)
	registerCatalogFlags(tf, &flags)

	if err := tf.Parse(args); err != nil {
		return Flags{}, err
	}

	finalizeServer()
	finalizeAuthentication()
	finalizeLogging()
	flags.Overrides = tf.Overrides()
	return flags, nil
}

// registerCatalogFlags registers exercise-catalog deployment overrides.
func registerCatalogFlags(tf *tinyflags.FlagSet, cfg *Flags) {
	tf.StringVar(&cfg.CoreExercisesFile, "core-exercises-file", "", "YAML file overriding the embedded core exercise catalog").
		Section(sectionCatalog).
		Placeholder("FILE").
		Value()
}
