package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"supafirehose/api"
	"supafirehose/config"
	"supafirehose/db"
	"supafirehose/load"
	"supafirehose/metrics"
)

//go:embed frontend/dist/*
var frontendFS embed.FS

func main() {
	// Parse flags
	devMode := flag.Bool("dev", false, "Development mode (proxy frontend to Vite)")
	flag.Parse()

	// Load configuration
	cfg := config.Load()

	log.Printf("Starting SupaFirehose on port %d", cfg.HTTPPort)

	// Create connection manager (no pool - direct connections)
	connMgr := db.NewConnectionManager(cfg.DatabaseURL)

	// Verify database connectivity
	ctx := context.Background()
	if err := connMgr.Ping(ctx); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Printf("Connected to database")

	// Create metrics collector with connection stats function
	collector := metrics.NewCollector(func() metrics.PoolStats {
		// Get database size (ignore errors, just return 0 if it fails)
		dbSize, _ := connMgr.GetDatabaseSize(ctx)
		return metrics.PoolStats{
			ActiveConnections: connMgr.ActiveConnections(),
			IdleConnections:   0,
			WaitingRequests:   0,
			DatabaseSizeBytes: dbSize,
		}
	})

	// Create load controller
	controller := load.NewController(connMgr, collector)

	// Set initial scenario
	controller.SetScenario(cfg.DefaultScenario, cfg.CustomTable)

	// Set initial config
	controller.SetConfig(load.Config{
		Connections: cfg.DefaultConnections,
		ReadQPS:     cfg.DefaultReadQPS,
		WriteQPS:    cfg.DefaultWriteQPS,
		ChurnRate:   0,
		Scenario:    cfg.DefaultScenario,
		CustomTable: cfg.CustomTable,
	})

	// Create API handlers
	handlers := api.NewHandlers(controller, collector)

	// Create WebSocket hub
	wsHub := api.NewWebSocketHub(collector, cfg.MetricsInterval)
	go wsHub.StartBroadcast()

	// Set up router
	var staticFS fs.FS
	if *devMode {
		log.Printf("Development mode: proxying frontend to http://localhost:5173")
	} else {
		// Use embedded frontend
		var err error
		staticFS, err = fs.Sub(frontendFS, "frontend/dist")
		if err != nil {
			log.Printf("Warning: No embedded frontend found: %v", err)
		}
	}

	router := api.NewRouter(handlers, wsHub, staticFS)

	// In dev mode, proxy non-API requests to Vite
	var handler http.Handler = router
	if *devMode {
		handler = devModeHandler(router)
	}

	// Create server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler: handler,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down...")
		controller.Stop()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()

	// Start server
	log.Printf("Server listening on http://localhost:%d", cfg.HTTPPort)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}

// devModeHandler proxies non-API requests to the Vite dev server
func devModeHandler(apiRouter http.Handler) http.Handler {
	viteURL, _ := url.Parse("http://localhost:5173")
	proxy := httputil.NewSingleHostReverseProxy(viteURL)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Route API and WebSocket requests to our handlers
		if len(r.URL.Path) >= 4 && r.URL.Path[:4] == "/api" {
			apiRouter.ServeHTTP(w, r)
			return
		}
		if len(r.URL.Path) >= 3 && r.URL.Path[:3] == "/ws" {
			apiRouter.ServeHTTP(w, r)
			return
		}

		// Proxy everything else to Vite
		proxy.ServeHTTP(w, r)
	})
}
