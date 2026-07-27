package web

import (
	"embed"
	"io/fs"
)

//go:embed dist
var embedded embed.FS

// Assets contains the built frontend files rooted at dist.
var Assets = mustSub(embedded, "dist")

func mustSub(source fs.FS, dir string) fs.FS {
	sub, err := fs.Sub(source, dir)
	if err != nil {
		panic(err)
	}
	return sub
}
