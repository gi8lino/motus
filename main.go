package main

import (
	"context"
	"embed"
	"os"

	"github.com/gi8lino/motus/internal/app"
)

var (
	Version = "dev"
	Commit  = "none"
)

//go:embed web/dist
var webFS embed.FS

// main boots the Motus application.
func main() {
	ctx := context.Background()
	if err := app.Run(ctx, webFS, Version, Commit, os.Args[1:], os.Stdout, os.Stderr); err != nil {
		os.Exit(1)
	}
}
