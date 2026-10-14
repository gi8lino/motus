// Package server provides Motus's passive application and management HTTP adapters.
package server

import appkithttpserver "github.com/containeroo/appkit/httpserver"

// Handlers contains the application and operational management handlers.
type Handlers = appkithttpserver.Handlers

// New constructs both HTTP applications with the same deployment prefix.
func New(config Config) Handlers {
	return appkithttpserver.New(appkithttpserver.Config{
		Application: newApplication(config),
		Management:  newManagement(config),
		RoutePrefix: config.RoutePrefix,
	})
}
