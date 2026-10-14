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
	"github.com/gi8lino/motus/internal/utils"
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
	// AllowUserRegistrationOverride overrides the persisted external-user registration policy when non-nil.
	AllowUserRegistrationOverride *bool
	// AuthModeOverride forces a browser authentication mode for recovery when non-empty.
	AuthModeOverride httpauth.Mode
	// LocalLogin enables local password sign-in alongside external authentication.
	LocalLogin bool
	// SessionLifetime controls local and OIDC session expiry.
	SessionLifetime time.Duration
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
	applyAuthFlags := registerAuthFlags(tf, &cfg)
	registerOIDCFlags(tf, &cfg)
	registerCatalogFlags(tf, &cfg)
	applyLoggingFlags := registerLoggingFlags(tf, &cfg)

	if err := tf.Parse(args); err != nil {
		return Config{}, err
	}

	applyServerFlags()
	applyAuthFlags()
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

// registerAuthFlags registers cross-mode authentication and recovery flags.
func registerAuthFlags(tf *tinyflags.FlagSet, cfg *Config) func() {
	authModeFlag := tinyflags.Enum(
		tf,
		"auth-mode",
		httpauth.ModeNone,
		"Emergency override for the database-managed authentication mode",
		httpauth.ModeNone,
		httpauth.ModeLocal,
		httpauth.ModeTrustedProxy,
		httpauth.ModeOIDC,
	).
		Section(sectionAuthentication).
		Placeholder("MODE")

	allowUserRegistrationFlag := tf.BoolVar(
		utils.ToPtr(false),
		"allow-user-registration",
		false,
		"Deployment override for whether unknown OIDC or trusted-proxy identities may create accounts",
	).
		Section(sectionAuthentication).
		Strict()

	tf.BoolVar(&cfg.LocalLogin, "local-login", false, "Enable local recovery login alongside the configured authentication mode").
		Section(sectionAuthentication).
		Value()

	tf.StringVar(&cfg.EncryptionKey, "encryption-key", "", "Base64-encoded 32-byte key used to encrypt sensitive application settings").
		Section(sectionAuthentication).
		OverriddenValueMaskFn(tinyflags.MaskFirstLast).
		Validate(secrets.ValidateKey).
		Value()

	tf.DurationVar(&cfg.SessionLifetime, "session-lifetime", 30*24*time.Hour, "Authentication session lifetime").
		Section(sectionAuthentication).
		Value()

	return func() {
		if authModeFlag.Changed() {
			cfg.AuthModeOverride = *authModeFlag.Value()
		}
		if allowUserRegistrationFlag.Changed() {
			cfg.AllowUserRegistrationOverride = allowUserRegistrationFlag.Value()
		}
	}
}

// registerOIDCFlags registers OpenID Connect provider configuration.
func registerOIDCFlags(tf *tinyflags.FlagSet, cfg *Config) {
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
