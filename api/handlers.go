package api

import (
	"encoding/json"
	"net/http"

	"supafirehose/load"
	"supafirehose/metrics"
	"supafirehose/schema"
)

// Handlers holds the HTTP handler dependencies
type Handlers struct {
	controller *load.Controller
	collector  *metrics.Collector
}

// NewHandlers creates a new handlers instance
func NewHandlers(controller *load.Controller, collector *metrics.Collector) *Handlers {
	return &Handlers{
		controller: controller,
		collector:  collector,
	}
}

// StatusResponse is the response for GET /api/status
type StatusResponse struct {
	Running       bool        `json:"running"`
	Config        load.Config `json:"config"`
	UptimeSeconds float64     `json:"uptime_seconds"`
}

// HandleStatus returns the current system status
func (h *Handlers) HandleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	resp := StatusResponse{
		Running:       h.controller.IsRunning(),
		Config:        h.controller.GetConfig(),
		UptimeSeconds: h.collector.Uptime().Seconds(),
	}

	writeJSON(w, resp)
}

// ConfigRequest is the request body for POST /api/config
type ConfigRequest struct {
	Connections int    `json:"connections"`
	ReadQPS     int    `json:"read_qps"`
	WriteQPS    int    `json:"write_qps"`
	ChurnRate   int    `json:"churn_rate"`
	Scenario    string `json:"scenario,omitempty"`
	CustomTable string `json:"custom_table,omitempty"`
}

// ConfigResponse is the response for POST /api/config
type ConfigResponse struct {
	OK     bool        `json:"ok"`
	Config load.Config `json:"config"`
}

// HandleConfig updates the workload configuration
func (h *Handlers) HandleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get current config for defaults
	currentConfig := h.controller.GetConfig()

	// Use current values if not provided
	scenario := req.Scenario
	if scenario == "" {
		scenario = currentConfig.Scenario
	}
	customTable := req.CustomTable
	if customTable == "" && scenario == currentConfig.Scenario {
		customTable = currentConfig.CustomTable
	}

	cfg := load.Config{
		Connections: req.Connections,
		ReadQPS:     req.ReadQPS,
		WriteQPS:    req.WriteQPS,
		ChurnRate:   req.ChurnRate,
		Scenario:    scenario,
		CustomTable: customTable,
	}

	h.controller.UpdateConfig(cfg)

	resp := ConfigResponse{
		OK:     true,
		Config: h.controller.GetConfig(),
	}

	writeJSON(w, resp)
}

// MessageResponse is a generic response with a message
type MessageResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

// HandleStart starts the load generator
func (h *Handlers) HandleStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.controller.Start()

	resp := MessageResponse{
		OK:      true,
		Message: "Load generator started",
	}

	writeJSON(w, resp)
}

// HandleStop stops the load generator
func (h *Handlers) HandleStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.controller.Stop()

	resp := MessageResponse{
		OK:      true,
		Message: "Load generator stopped",
	}

	writeJSON(w, resp)
}

// HandleReset resets all metrics
func (h *Handlers) HandleReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.collector.Reset()

	resp := MessageResponse{
		OK:      true,
		Message: "Metrics reset",
	}

	writeJSON(w, resp)
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// ScenariosResponse is the response for GET /api/scenarios
type ScenariosResponse struct {
	Scenarios []schema.ScenarioInfo `json:"scenarios"`
}

// HandleScenarios returns the list of available scenarios
func (h *Handlers) HandleScenarios(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	scenarios := h.controller.GetRegistry().List()

	// Add custom scenario option
	scenarios = append(scenarios, schema.ScenarioInfo{
		Name:        "custom",
		Description: "Custom table (optionally specify table name)",
		TableName:   "",
	})

	resp := ScenariosResponse{
		Scenarios: scenarios,
	}

	writeJSON(w, resp)
}
