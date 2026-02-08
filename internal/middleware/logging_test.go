package middleware

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLogging(t *testing.T) {
	// Create a simple handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap with logging middleware
	logged := Logging(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	logged.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestLoggingWithLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Created"))
	})

	logged := LoggingWithLogger(logger)(handler)

	req := httptest.NewRequest(http.MethodPost, "/api/books", nil)
	rec := httptest.NewRecorder()

	logged.ServeHTTP(rec, req)

	logOutput := buf.String()
	if !strings.Contains(logOutput, "POST") {
		t.Errorf("Log should contain method, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "/api/books") {
		t.Errorf("Log should contain path, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "201") {
		t.Errorf("Log should contain status code, got: %s", logOutput)
	}
}

func TestLogging_CapturesStatusCode(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found"))
	})

	logged := LoggingWithLogger(logger)(handler)

	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	rec := httptest.NewRecorder()

	logged.ServeHTTP(rec, req)

	logOutput := buf.String()
	if !strings.Contains(logOutput, "404") {
		t.Errorf("Log should contain 404 status, got: %s", logOutput)
	}
}

func TestLogging_CapturesBytesWritten(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!")) // 13 bytes
	})

	logged := LoggingWithLogger(logger)(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	logged.ServeHTTP(rec, req)

	logOutput := buf.String()
	if !strings.Contains(logOutput, "13 bytes") {
		t.Errorf("Log should contain bytes written, got: %s", logOutput)
	}
}

func TestRequestID(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	withID := RequestID(handler)

	// First request
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	rec1 := httptest.NewRecorder()
	withID.ServeHTTP(rec1, req1)

	reqID1 := rec1.Header().Get("X-Request-ID")
	if reqID1 == "" {
		t.Error("Expected X-Request-ID header")
	}
	if !strings.HasPrefix(reqID1, "req-") {
		t.Errorf("Expected request ID to start with 'req-', got %s", reqID1)
	}

	// Second request should have different ID
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	rec2 := httptest.NewRecorder()
	withID.ServeHTTP(rec2, req2)

	reqID2 := rec2.Header().Get("X-Request-ID")
	if reqID1 == reqID2 {
		t.Error("Expected different request IDs for different requests")
	}
}

func TestResponseWriter_DefaultStatus(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Write without explicitly setting status
		w.Write([]byte("OK"))
	})

	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	logged := LoggingWithLogger(logger)(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	logged.ServeHTTP(rec, req)

	logOutput := buf.String()
	if !strings.Contains(logOutput, "200") {
		t.Errorf("Log should contain default 200 status, got: %s", logOutput)
	}
}

func TestFormatRequestID(t *testing.T) {
	tests := []struct {
		id       uint64
		expected string
	}{
		{0, "req-0"},
		{1, "req-1"},
		{123, "req-123"},
		{999999, "req-999999"},
	}

	for _, tt := range tests {
		result := formatRequestID(tt.id)
		if result != tt.expected {
			t.Errorf("formatRequestID(%d) = %s, expected %s", tt.id, result, tt.expected)
		}
	}
}
