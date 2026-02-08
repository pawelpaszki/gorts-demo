package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pawelpaszki/gorts-demo/internal/handler"
	"github.com/pawelpaszki/gorts-demo/internal/middleware"
	"github.com/pawelpaszki/gorts-demo/internal/model"
	"github.com/pawelpaszki/gorts-demo/internal/repository"
	"github.com/pawelpaszki/gorts-demo/internal/service"
)

// TestServerWithAuth creates a test server with authentication enabled.
type TestServerWithAuth struct {
	Server    *httptest.Server
	UserStore *middleware.InMemoryUserStore
}

// NewTestServerWithAuth creates a test server with auth middleware.
func NewTestServerWithAuth() *TestServerWithAuth {
	// Create repositories
	bookRepo := repository.NewBookRepository()

	// Create services
	bookService := service.NewBookService(bookRepo)

	// Create handlers
	bookHandler := handler.NewBookHandler(bookService)
	healthHandler := handler.NewHealthHandler("1.0.0-test")

	// Create user store with test users
	userStore := middleware.NewInMemoryUserStore()
	userStore.AddUser("admin", "admin123", "admin")
	userStore.AddUser("editor", "editor123", "editor")
	userStore.AddUser("viewer", "viewer123", "viewer")

	// Setup routes
	mux := http.NewServeMux()

	// Public routes (no auth)
	healthHandler.RegisterRoutes(mux)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Write([]byte("Bookshelf API"))
	})

	// Protected routes - wrap book handler with auth
	protectedMux := http.NewServeMux()
	bookHandler.RegisterRoutes(protectedMux)

	// Apply auth middleware to protected routes
	authMiddleware := middleware.BasicAuth(userStore, "Bookshelf API")
	mux.Handle("/api/books", authMiddleware(protectedMux))
	mux.Handle("/api/books/", authMiddleware(protectedMux))

	// Wrap everything with logging
	var h http.Handler = mux
	h = middleware.Logging(h)
	h = middleware.RequestID(h)

	server := httptest.NewServer(h)

	return &TestServerWithAuth{
		Server:    server,
		UserStore: userStore,
	}
}

func (ts *TestServerWithAuth) Close() {
	ts.Server.Close()
}

func (ts *TestServerWithAuth) URL() string {
	return ts.Server.URL
}

func TestE2E_Auth_PublicEndpoints(t *testing.T) {
	ts := NewTestServerWithAuth()
	defer ts.Close()

	client := &http.Client{}

	// Health endpoint should be public
	resp, err := client.Get(ts.URL() + "/health")
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Health: expected %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Root endpoint should be public
	resp, err = client.Get(ts.URL() + "/")
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Root: expected %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestE2E_Auth_ProtectedEndpoints_NoAuth(t *testing.T) {
	ts := NewTestServerWithAuth()
	defer ts.Close()

	client := &http.Client{}

	// Try to access books without auth
	resp, err := client.Get(ts.URL() + "/api/books")
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected %d, got %d", http.StatusUnauthorized, resp.StatusCode)
	}

	// Should have WWW-Authenticate header
	if resp.Header.Get("WWW-Authenticate") == "" {
		t.Error("Expected WWW-Authenticate header")
	}
}

func TestE2E_Auth_ProtectedEndpoints_ValidAuth(t *testing.T) {
	ts := NewTestServerWithAuth()
	defer ts.Close()

	client := &http.Client{}

	// Create request with valid auth
	req, _ := http.NewRequest(http.MethodGet, ts.URL()+"/api/books", nil)
	req.Header.Set("Authorization", middleware.EncodeBasicAuth("admin", "admin123"))

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestE2E_Auth_ProtectedEndpoints_InvalidPassword(t *testing.T) {
	ts := NewTestServerWithAuth()
	defer ts.Close()

	client := &http.Client{}

	req, _ := http.NewRequest(http.MethodGet, ts.URL()+"/api/books", nil)
	req.Header.Set("Authorization", middleware.EncodeBasicAuth("admin", "wrongpassword"))

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected %d, got %d", http.StatusUnauthorized, resp.StatusCode)
	}
}

func TestE2E_Auth_ProtectedEndpoints_UnknownUser(t *testing.T) {
	ts := NewTestServerWithAuth()
	defer ts.Close()

	client := &http.Client{}

	req, _ := http.NewRequest(http.MethodGet, ts.URL()+"/api/books", nil)
	req.Header.Set("Authorization", middleware.EncodeBasicAuth("unknown", "password"))

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected %d, got %d", http.StatusUnauthorized, resp.StatusCode)
	}
}

func TestE2E_Auth_CRUD_WithAuth(t *testing.T) {
	ts := NewTestServerWithAuth()
	defer ts.Close()

	client := &http.Client{}
	authHeader := middleware.EncodeBasicAuth("admin", "admin123")

	// Create book with auth
	bookData := map[string]interface{}{
		"id":        "auth-book-1",
		"title":     "Authenticated Book",
		"isbn":      "978-auth-001",
		"author_id": "author-1",
	}
	body, _ := json.Marshal(bookData)

	req, _ := http.NewRequest(http.MethodPost, ts.URL()+"/api/books", bytes.NewReader(body))
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Create: expected %d, got %d", http.StatusCreated, resp.StatusCode)
	}
	resp.Body.Close()

	// Get book with auth
	req, _ = http.NewRequest(http.MethodGet, ts.URL()+"/api/books/auth-book-1", nil)
	req.Header.Set("Authorization", authHeader)

	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Get: expected %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var book model.Book
	json.NewDecoder(resp.Body).Decode(&book)
	resp.Body.Close()

	if book.Title != "Authenticated Book" {
		t.Errorf("Title = %q, want %q", book.Title, "Authenticated Book")
	}

	// Delete with auth
	req, _ = http.NewRequest(http.MethodDelete, ts.URL()+"/api/books/auth-book-1", nil)
	req.Header.Set("Authorization", authHeader)

	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("Delete: expected %d, got %d", http.StatusNoContent, resp.StatusCode)
	}
	resp.Body.Close()
}

func TestE2E_Auth_DifferentUsers(t *testing.T) {
	ts := NewTestServerWithAuth()
	defer ts.Close()

	client := &http.Client{}

	users := []struct {
		username string
		password string
	}{
		{"admin", "admin123"},
		{"editor", "editor123"},
		{"viewer", "viewer123"},
	}

	for _, user := range users {
		t.Run(user.username, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, ts.URL()+"/api/books", nil)
			req.Header.Set("Authorization", middleware.EncodeBasicAuth(user.username, user.password))

			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("User %s: expected %d, got %d", user.username, http.StatusOK, resp.StatusCode)
			}
		})
	}
}

func TestE2E_Auth_MalformedAuthHeader(t *testing.T) {
	ts := NewTestServerWithAuth()
	defer ts.Close()

	client := &http.Client{}

	tests := []struct {
		name   string
		header string
	}{
		{"empty", ""},
		{"wrong scheme", "Bearer token123"},
		{"malformed basic", "Basic notbase64!!!"},
		{"basic no colon", "Basic " + "bm9jb2xvbg=="}, // "nocolon" base64
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, ts.URL()+"/api/books", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}

			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusUnauthorized {
				t.Errorf("Expected %d, got %d", http.StatusUnauthorized, resp.StatusCode)
			}
		})
	}
}
