package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestHealthHandler() (*HealthHandler, *http.ServeMux) {
	handler := NewHealthHandler("1.0.0-test")
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	return handler, mux
}

func TestHealthHandler_Health(t *testing.T) {
	_, mux := newTestHealthHandler()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var status HealthStatus
	if err := json.NewDecoder(rec.Body).Decode(&status); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if status.Status != "healthy" {
		t.Errorf("Expected status 'healthy', got %q", status.Status)
	}
	if status.Version != "1.0.0-test" {
		t.Errorf("Expected version '1.0.0-test', got %q", status.Version)
	}
}

func TestHealthHandler_Health_WithCheckers(t *testing.T) {
	handler, mux := newTestHealthHandler()

	// Add healthy checker
	handler.RegisterChecker("database", func() error {
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var status HealthStatus
	json.NewDecoder(rec.Body).Decode(&status)

	if status.Checks["database"] != "healthy" {
		t.Errorf("Expected database check 'healthy', got %q", status.Checks["database"])
	}
}

func TestHealthHandler_Health_UnhealthyChecker(t *testing.T) {
	handler, mux := newTestHealthHandler()

	// Add unhealthy checker
	handler.RegisterChecker("database", func() error {
		return errors.New("connection refused")
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status %d, got %d", http.StatusServiceUnavailable, rec.Code)
	}

	var status HealthStatus
	json.NewDecoder(rec.Body).Decode(&status)

	if status.Status != "unhealthy" {
		t.Errorf("Expected status 'unhealthy', got %q", status.Status)
	}
	if !strings.Contains(status.Checks["database"], "connection refused") {
		t.Errorf("Expected error message in check, got %q", status.Checks["database"])
	}
}

func TestHealthHandler_Liveness(t *testing.T) {
	_, mux := newTestHealthHandler()

	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	body := rec.Body.String()
	if body != "OK" {
		t.Errorf("Expected body 'OK', got %q", body)
	}
}

func TestHealthHandler_Readiness_Ready(t *testing.T) {
	handler, mux := newTestHealthHandler()
	handler.SetReady(true)

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestHealthHandler_Readiness_NotReady(t *testing.T) {
	handler, mux := newTestHealthHandler()
	handler.SetReady(false)

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status %d, got %d", http.StatusServiceUnavailable, rec.Code)
	}
}

func TestHealthHandler_Info(t *testing.T) {
	_, mux := newTestHealthHandler()

	req := httptest.NewRequest(http.MethodGet, "/health/info", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var info map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&info); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	requiredFields := []string{"version", "go_version", "go_os", "go_arch", "cpus", "goroutines", "uptime"}
	for _, field := range requiredFields {
		if _, ok := info[field]; !ok {
			t.Errorf("Expected field %q in info response", field)
		}
	}
}

func TestHealthHandler_MethodNotAllowed(t *testing.T) {
	_, mux := newTestHealthHandler()

	endpoints := []string{"/health", "/health/live", "/health/ready", "/health/info"}

	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, endpoint, nil)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			if rec.Code != http.StatusMethodNotAllowed {
				t.Errorf("Expected status %d for POST %s, got %d", http.StatusMethodNotAllowed, endpoint, rec.Code)
			}
		})
	}
}

func TestHealthHandler_Uptime(t *testing.T) {
	_, mux := newTestHealthHandler()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	var status HealthStatus
	json.NewDecoder(rec.Body).Decode(&status)

	if status.Uptime == "" {
		t.Error("Expected uptime to be set")
	}
}
