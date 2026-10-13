package server

import (
	"net/http"

	httpauth "github.com/gi8lino/motus/internal/http/auth"
	"github.com/gi8lino/motus/internal/http/endpoint"
)

// registerPublicRoutes registers unauthenticated assets, setup, and login routes.
func registerPublicRoutes(mux *http.ServeMux, config Config) {
	assets := endpoint.Static(config.Assets)
	mux.Handle("GET /assets/", assets)
	mux.Handle("GET /fonts/", assets)
	mux.Handle("GET /sounds/", assets)
	// Root-level PWA and brand files must be served as files, not by the SPA fallback.
	mux.Handle("GET /brand.svg", assets)
	mux.Handle("GET /motus.svg", assets)
	mux.Handle("GET /sw.js", assets)
	mux.Handle("GET /manifest.webmanifest", assets)
	mux.Handle("GET /favicon-16x16.png", assets)
	mux.Handle("GET /favicon-32x32.png", assets)
	mux.Handle("GET /favicon-48x48.png", assets)
	mux.Handle("GET /favicon-64x64.png", assets)
	mux.Handle("GET /apple-touch-icon.png", assets)
	mux.Handle("GET /index.html", endpoint.SPA(config.Assets, config.RoutePrefix))
	// The SPA must discover the selected login mode before it can authenticate.
	mux.Handle("GET /api/config", endpoint.ClientConfiguration{
		Mode:              config.Browser.Mode,
		Local:             config.Browser.Local,
		AllowRegistration: config.AllowRegistration,
		Version:           config.Version,
		Commit:            config.Commit,
	}.Handler())
	mux.Handle("POST /api/setup", endpoint.Setup(config.Browser.Local, config.Logger))
	if config.Browser.Mode == httpauth.ModeLocal {
		mux.Handle("POST /api/login", endpoint.Login(config.Browser.Local, config.Logger))
		mux.Handle("POST /api/users", endpoint.CreateUser(config.Users, config.Browser.Local, config.Logger))
	}
	if config.Browser.Callback != nil {
		mux.Handle("GET /auth/oidc/login", config.Browser.Login)
		mux.Handle("GET /auth/oidc/callback", config.Browser.Callback)
	}
}
