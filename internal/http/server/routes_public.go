package server

import (
	"net/http"

	appkitweb "github.com/containeroo/appkit/web"
	"github.com/containeroo/httpkit/assets"
	"github.com/gi8lino/motus/internal/http/endpoint"
)

// registerPublicRoutes registers unauthenticated assets, setup, and login routes.
func registerPublicRoutes(mux *http.ServeMux, config Config) {
	setupRequired := func() bool { return false }
	if config.System != nil {
		setupRequired = config.System.SetupRequired
	}
	staticAssets := assets.Static(assets.StaticConfig{
		Files:           config.Assets,
		ImmutablePrefix: "/assets/",
		NoCacheFiles:    []string{"sw.js"},
	})
	mux.Handle("GET /assets/", staticAssets)
	mux.Handle("GET /fonts/", staticAssets)
	mux.Handle("GET /sounds/", staticAssets)
	// Root-level PWA and brand files must be served as files, not by the SPA fallback.
	mux.Handle("GET /brand.svg", staticAssets)
	mux.Handle("GET /motus.svg", staticAssets)
	mux.Handle("GET /sw.js", staticAssets)
	mux.Handle("GET /manifest.webmanifest", staticAssets)
	mux.Handle("GET /favicon-16x16.png", staticAssets)
	mux.Handle("GET /favicon-32x32.png", staticAssets)
	mux.Handle("GET /favicon-48x48.png", staticAssets)
	mux.Handle("GET /favicon-64x64.png", staticAssets)
	mux.Handle("GET /apple-touch-icon.png", staticAssets)
	mux.Handle("GET /index.html", (appkitweb.SPA{Assets: config.Assets, BasePath: config.RoutePrefix}).Handler())
	mux.Handle("GET /auth/local", (appkitweb.SPA{Assets: config.Assets, BasePath: config.RoutePrefix}).Handler())
	// The SPA must discover the selected login mode before it can authenticate.
	mux.Handle("GET /api/config", endpoint.ClientConfiguration{
		Current:       config.Browser.Current,
		SetupRequired: setupRequired,
		Version:       config.Version,
		Commit:        config.Commit,
	}.Handler())
	registerSetupRoutes(mux, config)
	mux.Handle("POST /api/login", endpoint.Login(config.Browser.Local, func() bool {
		return config.Browser.Current().LocalLoginEnabled
	}, config.Logger))
	mux.Handle("POST /api/users", endpoint.CreateUser(config.Users, config.Browser.Local, config.Logger))
	if config.Browser.Login != nil && config.Browser.Callback != nil {
		mux.Handle("GET /auth/oidc/login", config.Browser.Login)
		mux.Handle("GET /auth/oidc/callback", config.Browser.Callback)
	}
}

// registerSetupRoutes exposes bootstrap only when Motus started before setup completed.
func registerSetupRoutes(mux *http.ServeMux, config Config) {
	if config.System == nil || !config.System.SetupRequired() {
		return
	}
	mux.Handle("POST /api/setup", endpoint.Setup(config.System.SetupRequired, config.Browser.Local))
}
