//go:build dev

package main

import "net/http"

// In dev mode (go build -tags dev), the Vite dev server at :5173 handles
// the frontend. The Go server only handles /api/* routes.
func spaHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w,
			"Frontend is served by the Vite dev server at http://localhost:5173",
			http.StatusNotFound,
		)
	}
}
