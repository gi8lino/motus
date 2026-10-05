// Package flags parses deployment-level configuration for the Motus server.
package flags

import (
	"net"
	"strings"
	"time"

	httpprefix "github.com/containeroo/httpkit/prefix"
	"github.com/containeroo/tinyflags"
	httpauth "github.com/gi8lino/motus/internal/http/auth"
	"github.com/gi8lino/motus/internal/logging"
	"github.com/gi8lino/motus/internal/secrets"
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
	// RoutePrefix is the normalized deployment path, empty at root.
	RoutePrefix string
	// ApplicationListenAddress is the TCP address used by the public application server.
	ApplicationListenAddress string
	// ManagementListenAddress is the TCP address used by the operational management server.
	ManagementListenAddress string
	// PublicURL is the externally visible base URL, including RoutePrefix when configured.
	PublicURL string
	// DatabaseURL is the PostgreSQL connection URL.
	DatabaseURL string
	// AuthMode selects exactly one HTTP authentication strategy.
	AuthMode httpauth.Mode
	// TrustedProxyUserHeader is the identity header set by the trusted reverse proxy.
	TrustedProxyUserHeader string
	// AllowRegistration allows unknown users to create local or OIDC-backed accounts.
	AllowRegistration bool
	// SessionLifetime controls local and OIDC session expiry.
	SessionLifetime time.Duration
	// OIDCIssuer is the OpenID Connect issuer URL.
	OIDCIssuer string
	// OIDCClientID identifies Motus to the OpenID Connect provider.
	OIDCClientID string
	// OIDCClientSecret authenticates Motus to the OpenID Connect provider.
	OIDCClientSecret string
	// OIDCSessionSecret signs short-lived OpenID Connect login state.
	OIDCSessionSecret string
	// EncryptionKey encrypts sensitive persisted settings.
	EncryptionKey string
	// CoreExercisesFile optionally overrides the embedded exercise catalog.
	CoreExercisesFile string
	// LogFormat selects structured text or JSON logging.
	LogFormat logging.LogFormat
	// Debug enables verbose diagnostic logging.
	Debug bool
	// AccessLog enables HTTP request logging on the application listener.
	AccessLog bool
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

	applyServerFlags := registerServerFlags(tf, &cfg)
	registerAuthenticationFlags(tf, &cfg)
	registerOIDCFlags(tf, &cfg)
	registerCatalogFlags(tf, &cfg)
	applyLoggingFlags := registerLoggingFlags(tf, &cfg)

	if err := tf.Parse(args); err != nil {
		return Config{}, err
	}

	applyServerFlags()
	applyLoggingFlags()

	cfg.Overrides = tf.Overrides()
	return cfg, nil
}

// registerServerFlags registers public, management, and database listener configuration.
func registerServerFlags(tf *tinyflags.FlagSet, cfg *Config) func() {
	tf.StringVar(&cfg.RoutePrefix, "route-prefix", "", "URL path prefix under which Motus is served").
		Section(sectionServer).
		Validate(httpprefix.ValidateRoutePrefix).
		Finalize(httpprefix.NormalizeRoutePrefix).
		Value()

	application := tf.TCPAddr(
		"listen-address",
		&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080},
		"Address on which the application server listens",
	).
		Section(sectionServer).
		Short("a").
		Placeholder("ADDR").
		Value()

	management := tf.TCPAddr(
		"management-listen-address",
		&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8081},
		"Address on which health and readiness endpoints listen",
	).
		Section(sectionServer).
		Short("m").
		Placeholder("ADDR").
		Value()

	tf.StringVar(&cfg.PublicURL, "public-url", "http://localhost:8080", "Externally visible base URL").
		Section(sectionServer).
		Placeholder("URL").
		Finalize(func(value string) string { return strings.TrimRight(value, "/") }).
		Value()

	tf.StringVar(&cfg.DatabaseURL, "database-url", "", "PostgreSQL connection URL").
		Section(sectionDatabase).
		Required().
		Placeholder("URL").
		OverriddenValueMaskFn(tinyflags.MaskPostgresURL).
		Value()

	return func() {
		cfg.ApplicationListenAddress = (*application).String()
		cfg.ManagementListenAddress = (*management).String()
		if cfg.RoutePrefix != "" && !strings.HasSuffix(cfg.PublicURL, cfg.RoutePrefix) {
			cfg.PublicURL += cfg.RoutePrefix
		}
	}
}

// registerAuthenticationFlags registers local and trusted-header authentication configuration.
func registerAuthenticationFlags(tf *tinyflags.FlagSet, cfg *Config) {
	tf.StringVar((*string)(&cfg.AuthMode), "auth-mode", string(httpauth.ModeLocal), "Authentication mode").
		Section(sectionAuthentication).
		Choices(string(httpauth.ModeLocal), string(httpauth.ModeTrustedProxy), string(httpauth.ModeOIDC), string(httpauth.ModeNone)).
		Value()

	tf.StringVar(&cfg.TrustedProxyUserHeader, "trusted-proxy-user-header", "", "Trusted reverse-proxy identity header").
		Section(sectionAuthentication).
		Placeholder("HEADER").
		Value()

	tf.BoolVar(&cfg.AllowRegistration, "allow-registration", false, "Allow user self-registration").
		Section(sectionAuthentication).
		Value()

	tf.DurationVar(&cfg.SessionLifetime, "session-lifetime", 30*24*time.Hour, "Authentication session lifetime").
		Section(sectionAuthentication).
		Value()
}

// registerOIDCFlags registers OpenID Connect provider configuration.
func registerOIDCFlags(tf *tinyflags.FlagSet, cfg *Config) {
	tf.StringVar(&cfg.OIDCIssuer, "oidc-issuer", "", "OpenID Connect issuer URL").
		Section(sectionOIDC).
		Placeholder("URL").
		Value()

	tf.StringVar(&cfg.OIDCClientID, "oidc-client-id", "", "OpenID Connect client ID").
		Section(sectionOIDC).
		Placeholder("ID").
		Value()

	tf.StringVar(&cfg.OIDCClientSecret, "oidc-client-secret", "", "OpenID Connect client secret").
		Section(sectionOIDC).
		Placeholder("SECRET").
		OverriddenValueMaskFn(tinyflags.MaskFirstLast).
		Value()

	tf.StringVar(&cfg.OIDCSessionSecret, "oidc-session-secret", "", "Secret used to sign OIDC login state").
		Section(sectionOIDC).
		Placeholder("SECRET").
		OverriddenValueMaskFn(tinyflags.MaskFirstLast).
		Value()

	tf.StringVar(&cfg.EncryptionKey, "encryption-key", "", "Base64 256-bit key for sensitive persisted settings").
		Section(sectionAuthentication).
		Placeholder("KEY").
		Validate(secrets.ValidateKey).
		OverriddenValueMaskFn(tinyflags.MaskFirstLast).
		Value()
}

// registerCatalogFlags registers exercise-catalog deployment overrides.
func registerCatalogFlags(tf *tinyflags.FlagSet, cfg *Config) {
	tf.StringVar(&cfg.CoreExercisesFile, "core-exercises-file", "", "YAML file overriding the embedded core exercise catalog").
		Section(sectionCatalog).
		Placeholder("FILE").
		Value()
}

// registerLoggingFlags registers process logging configuration and returns its finalizer.
func registerLoggingFlags(tf *tinyflags.FlagSet, cfg *Config) func() {
	tf.BoolVar(&cfg.Debug, "debug", false, "Enable debug logging").
		Section(sectionLoggingDiagnostics).
		Short("d").
		Value()

	tf.BoolVar(&cfg.AccessLog, "access-log", false, "Enable HTTP access logging").
		Section(sectionLoggingDiagnostics).
		Value()

	logFormat := tf.String("log-format", string(logging.LogFormatJSON), "Log format").
		Section(sectionLoggingDiagnostics).
		Choices(string(logging.LogFormatText), string(logging.LogFormatJSON)).
		Short("l").
		Value()

	return func() {
		cfg.LogFormat = logging.LogFormat(*logFormat)
	}
}
