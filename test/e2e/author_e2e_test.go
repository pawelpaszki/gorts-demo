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

// TestServerWithAuthors creates a test server with author endpoints.
type TestServerWithAuthors struct {
	Server        *httptest.Server
	AuthorRepo    *repository.AuthorRepository
	AuthorService *service.AuthorService
}

// NewTestServerWithAuthors creates a test server with author support.
func NewTestServerWithAuthors() *TestServerWithAuthors {
	// Create repositories
	authorRepo := repository.NewAuthorRepository()

	// Create services
	authorService := service.NewAuthorService(authorRepo)

	// Create handlers
	authorHandler := handler.NewAuthorHandler(authorService)
	healthHandler := handler.NewHealthHandler("1.0.0-test")

	// Setup routes
	mux := http.NewServeMux()
	authorHandler.RegisterRoutes(mux)
	healthHandler.RegisterRoutes(mux)

	// Wrap with middleware
	var h http.Handler = mux
	h = middleware.Logging(h)
	h = middleware.RequestID(h)

	server := httptest.NewServer(h)

	return &TestServerWithAuthors{
		Server:        server,
		AuthorRepo:    authorRepo,
		AuthorService: authorService,
	}
}

func (ts *TestServerWithAuthors) Close() {
	ts.Server.Close()
}

func (ts *TestServerWithAuthors) URL() string {
	return ts.Server.URL
}

func TestE2E_Author_CreateAndGet(t *testing.T) {
	ts := NewTestServerWithAuthors()
	defer ts.Close()

	client := &http.Client{}

	// Create an author
	authorData := map[string]interface{}{
		"id":      "author-1",
		"name":    "Jane Doe",
		"bio":     "Award-winning author",
		"country": "USA",
	}
	body, _ := json.Marshal(authorData)

	resp, err := client.Post(ts.URL()+"/api/authors", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Create: expected %d, got %d", http.StatusCreated, resp.StatusCode)
	}
	resp.Body.Close()

	// Get the author
	resp, err = client.Get(ts.URL() + "/api/authors/author-1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Get: expected %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var author model.Author
	json.NewDecoder(resp.Body).Decode(&author)
	resp.Body.Close()

	if author.Name != "Jane Doe" {
		t.Errorf("Name = %q, want %q", author.Name, "Jane Doe")
	}
	if author.Country != "USA" {
		t.Errorf("Country = %q, want %q", author.Country, "USA")
	}
}

func TestE2E_Author_CRUD_FullLifecycle(t *testing.T) {
	ts := NewTestServerWithAuthors()
	defer ts.Close()

	client := &http.Client{}

	// CREATE
	authorData := map[string]interface{}{
		"id":      "lifecycle-author",
		"name":    "Original Name",
		"bio":     "Original bio",
		"country": "Canada",
	}
	body, _ := json.Marshal(authorData)

	resp, _ := client.Post(ts.URL()+"/api/authors", "application/json", bytes.NewReader(body))
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Create: expected %d, got %d", http.StatusCreated, resp.StatusCode)
	}
	resp.Body.Close()

	// READ
	resp, _ = client.Get(ts.URL() + "/api/authors/lifecycle-author")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Read: expected %d, got %d", http.StatusOK, resp.StatusCode)
	}
	resp.Body.Close()

	// UPDATE
	updateData := map[string]interface{}{
		"id":      "lifecycle-author",
		"name":    "Updated Name",
		"bio":     "Updated bio",
		"country": "UK",
	}
	body, _ = json.Marshal(updateData)
	req, _ := http.NewRequest(http.MethodPut, ts.URL()+"/api/authors/lifecycle-author", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ = client.Do(req)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Update: expected %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var updated model.Author
	json.NewDecoder(resp.Body).Decode(&updated)
	resp.Body.Close()

	if updated.Name != "Updated Name" {
		t.Errorf("Name after update = %q, want %q", updated.Name, "Updated Name")
	}

	// DELETE
	req, _ = http.NewRequest(http.MethodDelete, ts.URL()+"/api/authors/lifecycle-author", nil)
	resp, _ = client.Do(req)

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("Delete: expected %d, got %d", http.StatusNoContent, resp.StatusCode)
	}
	resp.Body.Close()

	// VERIFY DELETED
	resp, _ = client.Get(ts.URL() + "/api/authors/lifecycle-author")
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("After delete: expected %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
	resp.Body.Close()
}

func TestE2E_Author_ListAll(t *testing.T) {
	ts := NewTestServerWithAuthors()
	defer ts.Close()

	client := &http.Client{}

	// Create multiple authors
	authors := []map[string]interface{}{
		{"id": "a1", "name": "Author One", "country": "USA"},
		{"id": "a2", "name": "Author Two", "country": "USA"},
		{"id": "a3", "name": "Author Three", "country": "UK"},
	}

	for _, a := range authors {
		body, _ := json.Marshal(a)
		resp, _ := client.Post(ts.URL()+"/api/authors", "application/json", bytes.NewReader(body))
		resp.Body.Close()
	}

	// List all
	resp, _ := client.Get(ts.URL() + "/api/authors")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("List: expected %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var list []model.Author
	json.NewDecoder(resp.Body).Decode(&list)
	resp.Body.Close()

	if len(list) != 3 {
		t.Errorf("Expected 3 authors, got %d", len(list))
	}
}

func TestE2E_Author_FilterByCountry(t *testing.T) {
	ts := NewTestServerWithAuthors()
	defer ts.Close()

	client := &http.Client{}

	// Create authors from different countries
	authors := []map[string]interface{}{
		{"id": "a1", "name": "Author One", "country": "USA"},
		{"id": "a2", "name": "Author Two", "country": "USA"},
		{"id": "a3", "name": "Author Three", "country": "UK"},
		{"id": "a4", "name": "Author Four", "country": "Canada"},
	}

	for _, a := range authors {
		body, _ := json.Marshal(a)
		resp, _ := client.Post(ts.URL()+"/api/authors", "application/json", bytes.NewReader(body))
		resp.Body.Close()
	}

	// Filter by USA
	resp, _ := client.Get(ts.URL() + "/api/authors?country=USA")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Filter: expected %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var usaAuthors []model.Author
	json.NewDecoder(resp.Body).Decode(&usaAuthors)
	resp.Body.Close()

	if len(usaAuthors) != 2 {
		t.Errorf("Expected 2 USA authors, got %d", len(usaAuthors))
	}

	// Filter by UK
	resp, _ = client.Get(ts.URL() + "/api/authors?country=UK")
	var ukAuthors []model.Author
	json.NewDecoder(resp.Body).Decode(&ukAuthors)
	resp.Body.Close()

	if len(ukAuthors) != 1 {
		t.Errorf("Expected 1 UK author, got %d", len(ukAuthors))
	}
}

func TestE2E_Author_NotFound(t *testing.T) {
	ts := NewTestServerWithAuthors()
	defer ts.Close()

	client := &http.Client{}

	resp, _ := client.Get(ts.URL() + "/api/authors/non-existent")
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
	resp.Body.Close()
}

func TestE2E_Author_InvalidData(t *testing.T) {
	ts := NewTestServerWithAuthors()
	defer ts.Close()

	client := &http.Client{}

	// Missing name
	authorData := map[string]interface{}{
		"id":      "invalid-author",
		"country": "USA",
	}
	body, _ := json.Marshal(authorData)

	resp, _ := client.Post(ts.URL()+"/api/authors", "application/json", bytes.NewReader(body))
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
	resp.Body.Close()
}
