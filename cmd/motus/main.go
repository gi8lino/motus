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

// main runs the Motus server and exits non-zero on failure.
func main() {
	if err := app.Run(
		context.Background(),
		os.Args[1:],
		web.Assets,
		Version,
		Commit,
		os.Stdout,
		os.Stderr,
	); err != nil {
		os.Exit(1)
	}
}
