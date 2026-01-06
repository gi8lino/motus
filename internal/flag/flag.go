package flag

import (
	"net"
	"strings"

	"github.com/containeroo/tinyflags"

	"github.com/gi8lino/motus/internal/logging"
	"github.com/gi8lino/motus/internal/server"
)

// Options holds the application configuration.
type Options struct {
	Debug             bool              // Set LogLevel to Debug
	LogFormat         logging.LogFormat // Specify the log output format
	ListenAddr        string            // Address to listen on
	SiteRoot          string            // Root URL of the site
	RoutePrefix       string            // Route prefix
	AuthHeader        string            // Authentication header
	AllowRegistration bool              // Allow user self-registration
	AutoCreateUsers   bool              // Auto-create users in auth-header mode
	DatabaseURL       string            // Database URL
	OverriddenValues  map[string]any    // Overridden values from environment
	AdminEmail        string            // AdminEmail is the email address of the site admin
	AdminPassword     string            // AdminPassword is the password for the site admin
}

// ParseFlags parses flags and environment variables.
func ParseFlags(args []string, version string) (Options, error) {
	opts := Options{}

	tf := tinyflags.NewFlagSet("motus", tinyflags.ContinueOnError)
	tf.Version(version)
	tf.EnvPrefix("MOTUS")

	listenAddr := tf.TCPAddr("listen-address", &net.TCPAddr{IP: nil, Port: 8080}, "Listen address").
		Short("a").
		Placeholder("ADDR").
		Value()

	tf.StringVar(&opts.DatabaseURL, "database-url", "", "Database URL").
		Required().
		OverriddenValueMaskFn(tinyflags.MaskPostgresURL).
		Placeholder("URL").
		Value()

	tf.StringVar(&opts.RoutePrefix, "route-prefix", "", "Path prefix to mount the app (e.g., /heartbeats). Empty = root.").
		Finalize(func(input string) string {
			return server.NormalizeRoutePrefix(input) // canonical "" or "/motus"
		}).
		Placeholder("PATH").
		Value()

	tf.StringVar(&opts.SiteRoot, "site-root", "http://localhost:8080", "Site root URL").
		Finalize(func(input string) string {
			return strings.TrimRight(input, "/")
		}).
		Short("r").
		Placeholder("URL").
		Value()

	tf.StringVar(&opts.AuthHeader, "auth-header", "", "Authentication header").
		Placeholder("HEADER").
		Value()

	tf.BoolVar(&opts.AllowRegistration, "allow-registration", false, "Allow user self-registration").
		Value()

	tf.BoolVar(&opts.AutoCreateUsers, "auto-create-users", false, "Auto-create users when auth-header is enabled").
		Value()

	// If set, the admin email and password are used to create a user on startup.
	tf.StringVar(&opts.AdminEmail, "admin-email", "", "Site admin email").
		Placeholder("EMAIL").
		OverriddenValueMaskFn(tinyflags.MaskFirstLast).
		Value()
	tf.StringVar(&opts.AdminPassword, "admin-password", "", "Site admin password").
		Placeholder("PASSWORD").
		OverriddenValueMaskFn(tinyflags.MaskFirstLast).
		Value()

	tf.BoolVar(&opts.Debug, "debug", false, "Enable debug logging").Short("d").Value()

	logFormat := tf.String("log-format", "json", "Log format").
		Choices(string(logging.LogFormatText), string(logging.LogFormatJSON)).
		Short("l").
		Value()

	if err := tf.Parse(args); err != nil {
		return Options{}, err
	}

	opts.ListenAddr = (*listenAddr).String()
	opts.LogFormat = logging.LogFormat(*logFormat)
	opts.OverriddenValues = tf.OverriddenValues()
	if opts.RoutePrefix != "" && !strings.HasSuffix(opts.SiteRoot, opts.RoutePrefix) {
		opts.SiteRoot += opts.RoutePrefix
	}

	return opts, nil
}
