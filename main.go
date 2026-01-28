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
	if err := app.Run(
		context.Background(),
		webFS,
		Version,
		Commit,
		os.Args[1:],
		os.Stdout,
	); err != nil {
		os.Exit(1)
	}
}
