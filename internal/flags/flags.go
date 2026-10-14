// Package flags parses deployment-level configuration for the Motus server.
package flags

import (
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

// Config contains deployment-level runtime configuration for Motus.
type Config struct {
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
func Parse(args []string, version string) (Config, error) {
	cfg := Config{}

	tf := tinyflags.NewFlagSet("motus", tinyflags.ContinueOnError)
	tf.EnvPrefix("MOTUS_")
	tf.Version(version)
	tf.HideEnvs()
	tf.Note("\nFlags can also be set through environment variables using the MOTUS__ prefix. " +
		"Flag names are uppercased and hyphens are replaced with underscores; for example, " +
		"--listen-address becomes MOTUS__LISTEN_ADDRESS.")
	tf.SectionOrder(
		sectionServer,
		sectionDatabase,
		sectionAuthentication,
		sectionOIDC,
		sectionCatalog,
		sectionLoggingDiagnostics,
	)

	finalizeServer := appkitconfig.RegisterServer(tf, &cfg.Server, appkitconfig.ServerOptions{})
	appkitconfig.RegisterPostgres(tf, &cfg.Database, appkitconfig.PostgresOptions{})
	finalizeAuthentication := appkitconfig.RegisterAuthentication(
		tf,
		&cfg.Authentication,
		appkitconfig.AuthenticationOptions{},
	)
	appkitconfig.RegisterProxy(tf, &cfg.Proxy)
	appkitconfig.RegisterOIDC(tf, &cfg.OIDC)
	finalizeLogging := appkitconfig.RegisterLogging(tf, &cfg.Logging)
	registerCatalogFlags(tf, &cfg)

	if err := tf.Parse(args); err != nil {
		return Config{}, err
	}

	finalizeServer()
	finalizeAuthentication()
	finalizeLogging()
	cfg.Overrides = tf.Overrides()

	cfg.Overrides = tf.Overrides()
	return cfg, nil
}

// registerCatalogFlags registers exercise-catalog deployment overrides.
func registerCatalogFlags(tf *tinyflags.FlagSet, cfg *Config) {
	tf.StringVar(&cfg.CoreExercisesFile, "core-exercises-file", "", "YAML file overriding the embedded core exercise catalog").
		Section(sectionCatalog).
		Placeholder("FILE").
		Value()
}
