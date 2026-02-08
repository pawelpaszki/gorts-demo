package e2e

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/pawelpaszki/gorts-demo/internal/model"
)

func TestE2E_CreateAndGetBook(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()

	// Create a book
	bookData := map[string]interface{}{
		"id":        "e2e-book-1",
		"title":     "E2E Test Book",
		"isbn":      "978-1234567890",
		"author_id": "author-1",
		"pages":     250,
		"genre":     "Testing",
	}
	body, _ := json.Marshal(bookData)

	resp, err := http.Post(ts.URL()+"/api/books", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to create book: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status %d, got %d: %s", http.StatusCreated, resp.StatusCode, string(bodyBytes))
	}

	// Verify response has request ID header
	if resp.Header.Get("X-Request-ID") == "" {
		t.Error("Expected X-Request-ID header")
	}

	// Get the book
	resp, err = http.Get(ts.URL() + "/api/books/e2e-book-1")
	if err != nil {
		t.Fatalf("Failed to get book: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var retrieved model.Book
	if err := json.NewDecoder(resp.Body).Decode(&retrieved); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if retrieved.Title != "E2E Test Book" {
		t.Errorf("Title = %q, want %q", retrieved.Title, "E2E Test Book")
	}
	if retrieved.ISBN != "978-1234567890" {
		t.Errorf("ISBN = %q, want %q", retrieved.ISBN, "978-1234567890")
	}
	if retrieved.Pages != 250 {
		t.Errorf("Pages = %d, want %d", retrieved.Pages, 250)
	}
	if retrieved.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
}

func TestE2E_GetBook_NotFound(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()

	resp, err := http.Get(ts.URL() + "/api/books/non-existent-book")
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
	}

	var errResp map[string]string
	json.NewDecoder(resp.Body).Decode(&errResp)
	if errResp["error"] != "Book not found" {
		t.Errorf("Error message = %q, want %q", errResp["error"], "Book not found")
	}
}

func TestE2E_CreateBook_InvalidData(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()

	tests := []struct {
		name       string
		bookData   map[string]interface{}
		wantStatus int
	}{
		{
			name: "missing title",
			bookData: map[string]interface{}{
				"id":        "book-1",
				"isbn":      "123",
				"author_id": "author-1",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "missing ISBN",
			bookData: map[string]interface{}{
				"id":        "book-2",
				"title":     "Test",
				"author_id": "author-1",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "missing author_id",
			bookData: map[string]interface{}{
				"id":    "book-3",
				"title": "Test",
				"isbn":  "123",
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.bookData)
			resp, err := http.Post(ts.URL()+"/api/books", "application/json", bytes.NewReader(body))
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, resp.StatusCode)
			}
		})
	}
}

func TestE2E_CreateBook_InvalidJSON(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()

	resp, err := http.Post(ts.URL()+"/api/books", "application/json", bytes.NewReader([]byte("invalid json")))
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestE2E_CreateBook_DuplicateISBN(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()

	// Create first book
	book1 := map[string]interface{}{
		"id":        "book-1",
		"title":     "First Book",
		"isbn":      "duplicate-isbn",
		"author_id": "author-1",
	}
	body, _ := json.Marshal(book1)
	resp, _ := http.Post(ts.URL()+"/api/books", "application/json", bytes.NewReader(body))
	resp.Body.Close()

	// Try to create second book with same ISBN
	book2 := map[string]interface{}{
		"id":        "book-2",
		"title":     "Second Book",
		"isbn":      "duplicate-isbn",
		"author_id": "author-2",
	}
	body, _ = json.Marshal(book2)
	resp, err := http.Post(ts.URL()+"/api/books", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusConflict {
		t.Errorf("Expected status %d, got %d", http.StatusConflict, resp.StatusCode)
	}
}

func TestE2E_RootEndpoint(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()

	resp, err := http.Get(ts.URL() + "/")
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "Bookshelf API v1.0.0-test" {
		t.Errorf("Body = %q, want %q", string(body), "Bookshelf API v1.0.0-test")
	}
}

func TestE2E_HealthEndpoint(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()

	resp, err := http.Get(ts.URL() + "/health")
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var health map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&health)

	if health["status"] != "healthy" {
		t.Errorf("Health status = %v, want %q", health["status"], "healthy")
	}
	if health["version"] != "1.0.0-test" {
		t.Errorf("Health version = %v, want %q", health["version"], "1.0.0-test")
	}
}
