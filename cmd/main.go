package main

import (
	"context"
	"os"

	"github.com/gi8lino/motus/internal/app"
	"github.com/gi8lino/motus/web"
)

var (
	Version = "dev"
	Commit  = "none"
)

// main boots the Motus application.
func main() {
	ctx := context.Background()
	if err := app.Run(ctx, web.Assets, Version, Commit, os.Args[1:], os.Stdout, os.Stderr); err != nil {
		os.Exit(1)
	}
}
