package handler

import (
	"encoding/json"
	"net/http"
	"runtime"
	"sync/atomic"
	"time"
)

// HealthStatus represents the health check response.
type HealthStatus struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Version   string            `json:"version"`
	Uptime    string            `json:"uptime"`
	Checks    map[string]string `json:"checks,omitempty"`
}

// HealthHandler handles health check requests.
type HealthHandler struct {
	startTime time.Time
	version   string
	ready     atomic.Bool
	checkers  map[string]HealthChecker
}

// HealthChecker defines a health check function.
type HealthChecker func() error

// NewHealthHandler creates a new health handler.
func NewHealthHandler(version string) *HealthHandler {
	h := &HealthHandler{
		startTime: time.Now(),
		version:   version,
		checkers:  make(map[string]HealthChecker),
	}
	h.ready.Store(true)
	return h
}

// RegisterChecker registers a named health checker.
func (h *HealthHandler) RegisterChecker(name string, checker HealthChecker) {
	h.checkers[name] = checker
}

// SetReady sets the readiness state.
func (h *HealthHandler) SetReady(ready bool) {
	h.ready.Store(ready)
}

// RegisterRoutes registers health routes on the given mux.
func (h *HealthHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.handleHealth)
	mux.HandleFunc("/health/live", h.handleLiveness)
	mux.HandleFunc("/health/ready", h.handleReadiness)
	mux.HandleFunc("/health/info", h.handleInfo)
}

// handleHealth is the main health check endpoint.
func (h *HealthHandler) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now().UTC(),
		Version:   h.version,
		Uptime:    time.Since(h.startTime).Round(time.Second).String(),
		Checks:    make(map[string]string),
	}

	allHealthy := true
	for name, checker := range h.checkers {
		if err := checker(); err != nil {
			status.Checks[name] = "unhealthy: " + err.Error()
			allHealthy = false
		} else {
			status.Checks[name] = "healthy"
		}
	}

	if !allHealthy {
		status.Status = "unhealthy"
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	respondHealthJSON(w, status)
}

// handleLiveness is the Kubernetes liveness probe endpoint.
// Returns 200 OK if the process is alive and can handle requests.
func (h *HealthHandler) handleLiveness(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// handleReadiness is the Kubernetes readiness probe endpoint.
func (h *HealthHandler) handleReadiness(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !h.ready.Load() {
		http.Error(w, "Not Ready", http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Ready"))
}

// handleInfo returns detailed runtime information.
func (h *HealthHandler) handleInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	info := map[string]interface{}{
		"version":     h.version,
		"go_version":  runtime.Version(),
		"go_os":       runtime.GOOS,
		"go_arch":     runtime.GOARCH,
		"cpus":        runtime.NumCPU(),
		"goroutines":  runtime.NumGoroutine(),
		"uptime":      time.Since(h.startTime).Round(time.Second).String(),
		"start_time":  h.startTime.UTC().Format(time.RFC3339),
		"memory_alloc_mb": float64(mem.Alloc) / 1024 / 1024,
		"memory_sys_mb":   float64(mem.Sys) / 1024 / 1024,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

func respondHealthJSON(w http.ResponseWriter, status HealthStatus) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}
