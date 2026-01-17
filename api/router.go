package api

import (
	"io/fs"
	"net/http"
	"strings"
)

// NewRouter creates the HTTP router with all routes configured
func NewRouter(handlers *Handlers, wsHub *WebSocketHub, staticFS fs.FS) http.Handler {
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/status", handlers.HandleStatus)
	mux.HandleFunc("/api/config", handlers.HandleConfig)
	mux.HandleFunc("/api/start", handlers.HandleStart)
	mux.HandleFunc("/api/stop", handlers.HandleStop)
	mux.HandleFunc("/api/reset", handlers.HandleReset)
	mux.HandleFunc("/api/scenarios", handlers.HandleScenarios)

	// WebSocket route
	mux.HandleFunc("/ws/metrics", wsHub.HandleWebSocket)

	// Static files with SPA fallback
	if staticFS != nil {
		fileServer := http.FileServer(http.FS(staticFS))
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			// Try to serve the file directly
			path := r.URL.Path
			if path == "/" {
				path = "/index.html"
			}

			// Check if file exists
			f, err := staticFS.Open(strings.TrimPrefix(path, "/"))
			if err != nil {
				// File not found, serve index.html for SPA routing
				r.URL.Path = "/"
				fileServer.ServeHTTP(w, r)
				return
			}
			f.Close()

			fileServer.ServeHTTP(w, r)
		})
	}

	return mux
}
