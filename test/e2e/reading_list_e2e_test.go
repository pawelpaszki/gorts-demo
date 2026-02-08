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

// TestServerWithReadingLists extends TestServer with reading list support.
type TestServerWithReadingLists struct {
	*TestServer
	ReadingListRepo    *repository.ReadingListRepository
	ReadingListService *service.ReadingListService
}

// NewTestServerWithReadingLists creates a test server with reading list support.
func NewTestServerWithReadingLists() *TestServerWithReadingLists {
	// Create repositories
	bookRepo := repository.NewBookRepository()
	authorRepo := repository.NewAuthorRepository()
	readingListRepo := repository.NewReadingListRepository()

	// Create services
	bookService := service.NewBookService(bookRepo)
	readingListService := service.NewReadingListService(readingListRepo, bookRepo)

	// Create handlers
	bookHandler := handler.NewBookHandler(bookService)
	readingListHandler := handler.NewReadingListHandler(readingListService)
	healthHandler := handler.NewHealthHandler("1.0.0-test")

	// Setup routes
	mux := http.NewServeMux()
	bookHandler.RegisterRoutes(mux)
	readingListHandler.RegisterRoutes(mux)
	healthHandler.RegisterRoutes(mux)

	// Wrap with middleware
	var h http.Handler = mux
	h = middleware.Logging(h)
	h = middleware.RequestID(h)

	// Create test server
	server := httptest.NewServer(h)

	return &TestServerWithReadingLists{
		TestServer: &TestServer{
			Server:      server,
			Mux:         mux,
			BookRepo:    bookRepo,
			AuthorRepo:  authorRepo,
			BookService: bookService,
		},
		ReadingListRepo:    readingListRepo,
		ReadingListService: readingListService,
	}
}

func TestE2E_ReadingList_CRUD(t *testing.T) {
	ts := NewTestServerWithReadingLists()
	defer ts.Close()

	client := ts.Server.Client()
	baseURL := ts.URL()

	// Create a reading list
	listData := map[string]interface{}{
		"id":          "my-list-1",
		"name":        "My Favorite Books",
		"description": "A collection of books I love",
	}
	body, _ := json.Marshal(listData)

	resp, err := client.Post(baseURL+"/api/lists", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Create list failed: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Create: expected %d, got %d", http.StatusCreated, resp.StatusCode)
	}
	resp.Body.Close()

	// Get the list
	resp, _ = client.Get(baseURL + "/api/lists/my-list-1")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Get: expected %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var list model.ReadingList
	json.NewDecoder(resp.Body).Decode(&list)
	resp.Body.Close()

	if list.Name != "My Favorite Books" {
		t.Errorf("Name = %q, want %q", list.Name, "My Favorite Books")
	}

	// Update the list
	updateData := map[string]interface{}{
		"id":          "my-list-1",
		"name":        "Updated List Name",
		"description": "Updated description",
	}
	body, _ = json.Marshal(updateData)
	req, _ := http.NewRequest(http.MethodPut, baseURL+"/api/lists/my-list-1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ = client.Do(req)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Update: expected %d, got %d", http.StatusOK, resp.StatusCode)
	}
	resp.Body.Close()

	// Delete the list
	req, _ = http.NewRequest(http.MethodDelete, baseURL+"/api/lists/my-list-1", nil)
	resp, _ = client.Do(req)

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("Delete: expected %d, got %d", http.StatusNoContent, resp.StatusCode)
	}
	resp.Body.Close()

	// Verify deleted
	resp, _ = client.Get(baseURL + "/api/lists/my-list-1")
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("After delete: expected %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
	resp.Body.Close()
}

func TestE2E_ReadingList_AddRemoveBooks(t *testing.T) {
	ts := NewTestServerWithReadingLists()
	defer ts.Close()

	client := ts.Server.Client()
	baseURL := ts.URL()

	// Create books first
	books := []map[string]interface{}{
		{"id": "book-1", "title": "Book One", "isbn": "isbn-1", "author_id": "author-1"},
		{"id": "book-2", "title": "Book Two", "isbn": "isbn-2", "author_id": "author-1"},
		{"id": "book-3", "title": "Book Three", "isbn": "isbn-3", "author_id": "author-2"},
	}

	for _, book := range books {
		body, _ := json.Marshal(book)
		resp, _ := client.Post(baseURL+"/api/books", "application/json", bytes.NewReader(body))
		resp.Body.Close()
	}

	// Create a reading list
	listData := map[string]interface{}{
		"id":   "reading-list-1",
		"name": "My Reading List",
	}
	body, _ := json.Marshal(listData)
	resp, _ := client.Post(baseURL+"/api/lists", "application/json", bytes.NewReader(body))
	resp.Body.Close()

	// Add books to the list
	req, _ := http.NewRequest(http.MethodPost, baseURL+"/api/lists/reading-list-1/books/book-1", nil)
	resp, _ = client.Do(req)
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("Add book-1: expected %d, got %d", http.StatusNoContent, resp.StatusCode)
	}
	resp.Body.Close()

	req, _ = http.NewRequest(http.MethodPost, baseURL+"/api/lists/reading-list-1/books/book-2", nil)
	resp, _ = client.Do(req)
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("Add book-2: expected %d, got %d", http.StatusNoContent, resp.StatusCode)
	}
	resp.Body.Close()

	// Verify books in list
	resp, _ = client.Get(baseURL + "/api/lists/reading-list-1")
	var list model.ReadingList
	json.NewDecoder(resp.Body).Decode(&list)
	resp.Body.Close()

	if len(list.BookIDs) != 2 {
		t.Errorf("Expected 2 books in list, got %d", len(list.BookIDs))
	}

	// Try to add duplicate book
	req, _ = http.NewRequest(http.MethodPost, baseURL+"/api/lists/reading-list-1/books/book-1", nil)
	resp, _ = client.Do(req)
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("Add duplicate: expected %d, got %d", http.StatusConflict, resp.StatusCode)
	}
	resp.Body.Close()

	// Remove a book
	req, _ = http.NewRequest(http.MethodDelete, baseURL+"/api/lists/reading-list-1/books/book-1", nil)
	resp, _ = client.Do(req)
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("Remove book-1: expected %d, got %d", http.StatusNoContent, resp.StatusCode)
	}
	resp.Body.Close()

	// Verify book removed
	resp, _ = client.Get(baseURL + "/api/lists/reading-list-1")
	json.NewDecoder(resp.Body).Decode(&list)
	resp.Body.Close()

	if len(list.BookIDs) != 1 {
		t.Errorf("Expected 1 book after removal, got %d", len(list.BookIDs))
	}
	if list.BookIDs[0] != "book-2" {
		t.Errorf("Expected book-2 to remain, got %s", list.BookIDs[0])
	}
}

func TestE2E_ReadingList_AddNonExistentBook(t *testing.T) {
	ts := NewTestServerWithReadingLists()
	defer ts.Close()

	client := ts.Server.Client()
	baseURL := ts.URL()

	// Create a reading list
	listData := map[string]interface{}{
		"id":   "list-1",
		"name": "Test List",
	}
	body, _ := json.Marshal(listData)
	resp, _ := client.Post(baseURL+"/api/lists", "application/json", bytes.NewReader(body))
	resp.Body.Close()

	// Try to add non-existent book
	req, _ := http.NewRequest(http.MethodPost, baseURL+"/api/lists/list-1/books/non-existent-book", nil)
	resp, _ = client.Do(req)

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
	resp.Body.Close()
}

func TestE2E_ReadingList_OperationsOnNonExistentList(t *testing.T) {
	ts := NewTestServerWithReadingLists()
	defer ts.Close()

	client := ts.Server.Client()
	baseURL := ts.URL()

	// Get non-existent list
	resp, _ := client.Get(baseURL + "/api/lists/non-existent")
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Get: expected %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
	resp.Body.Close()

	// Delete non-existent list
	req, _ := http.NewRequest(http.MethodDelete, baseURL+"/api/lists/non-existent", nil)
	resp, _ = client.Do(req)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Delete: expected %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
	resp.Body.Close()

	// Add book to non-existent list
	req, _ = http.NewRequest(http.MethodPost, baseURL+"/api/lists/non-existent/books/book-1", nil)
	resp, _ = client.Do(req)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Add book: expected %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
	resp.Body.Close()
}

func TestE2E_ReadingList_ListAll(t *testing.T) {
	ts := NewTestServerWithReadingLists()
	defer ts.Close()

	client := ts.Server.Client()
	baseURL := ts.URL()

	// Create multiple lists
	for i := 1; i <= 3; i++ {
		listData := map[string]interface{}{
			"id":   "list-" + string(rune('0'+i)),
			"name": "List " + string(rune('0'+i)),
		}
		body, _ := json.Marshal(listData)
		resp, _ := client.Post(baseURL+"/api/lists", "application/json", bytes.NewReader(body))
		resp.Body.Close()
	}

	// List all
	resp, _ := client.Get(baseURL + "/api/lists")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("List: expected %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var lists []model.ReadingList
	json.NewDecoder(resp.Body).Decode(&lists)
	resp.Body.Close()

	if len(lists) != 3 {
		t.Errorf("Expected 3 lists, got %d", len(lists))
	}
}
