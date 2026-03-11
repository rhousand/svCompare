//go:build !dev

package main

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

// frontendFS holds the built Vue SPA. The frontend/dist directory is populated
// during the production build (Dockerfile stage 2 / Nix preBuild step).
//
//go:embed all:frontend/dist
var frontendFS embed.FS

func spaHandler() http.HandlerFunc {
	dist, err := fs.Sub(frontendFS, "frontend/dist")
	if err != nil {
		panic("frontend/dist not embedded: " + err.Error())
	}
	fileServer := http.FileServer(http.FS(dist))

	return func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/")
		if name == "" {
			name = "index.html"
		}

		// If the file exists in the embedded FS, serve it directly.
		if f, err := dist.Open(name); err == nil {
			f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}

		// SPA fallback: serve index.html so Vue Router handles the path.
		r2 := r.Clone(r.Context())
		r2.URL.Path = "/"
		fileServer.ServeHTTP(w, r2)
	}
}
