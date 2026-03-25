package api

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"supafirehose/metrics"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for demo purposes
	},
}

// WebSocketHub manages WebSocket connections and broadcasts metrics
type WebSocketHub struct {
	mu        sync.RWMutex
	clients   map[*websocket.Conn]bool
	collector *metrics.Collector
	interval  time.Duration
}

// NewWebSocketHub creates a new WebSocket hub
func NewWebSocketHub(collector *metrics.Collector, interval time.Duration) *WebSocketHub {
	return &WebSocketHub{
		clients:   make(map[*websocket.Conn]bool),
		collector: collector,
		interval:  interval,
	}
}

// HandleWebSocket handles WebSocket upgrade and connection
func (hub *WebSocketHub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	hub.mu.Lock()
	hub.clients[conn] = true
	hub.mu.Unlock()

	// Handle client disconnect
	go func() {
		defer func() {
			hub.mu.Lock()
			delete(hub.clients, conn)
			hub.mu.Unlock()
			conn.Close()
		}()

		// Read messages (mainly to detect disconnect)
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}()
}

// StartBroadcast starts the metrics broadcast loop
func (hub *WebSocketHub) StartBroadcast() {
	ticker := time.NewTicker(hub.interval)
	defer ticker.Stop()

	var lastErrorsVersion int64
	for range ticker.C {
		errVersion := hub.collector.ErrorsVersion()
		snapshot := hub.collector.Snapshot(hub.interval, lastErrorsVersion)
		lastErrorsVersion = errVersion
		hub.broadcast(snapshot)
	}
}

func (hub *WebSocketHub) broadcast(snapshot metrics.MetricsSnapshot) {
	data, err := json.Marshal(snapshot)
	if err != nil {
		log.Printf("Failed to marshal metrics: %v", err)
		return
	}

	hub.mu.RLock()
	defer hub.mu.RUnlock()

	for conn := range hub.clients {
		conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		err := conn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Printf("WebSocket write error: %v", err)
			conn.Close()
			go func(c *websocket.Conn) {
				hub.mu.Lock()
				delete(hub.clients, c)
				hub.mu.Unlock()
			}(conn)
		}
	}
}

// ClientCount returns the number of connected clients
func (hub *WebSocketHub) ClientCount() int {
	hub.mu.RLock()
	defer hub.mu.RUnlock()
	return len(hub.clients)
}
