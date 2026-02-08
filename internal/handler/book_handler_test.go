package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pawelpaszki/gorts-demo/internal/model"
	"github.com/pawelpaszki/gorts-demo/internal/repository"
	"github.com/pawelpaszki/gorts-demo/internal/service"
)

func newTestHandler() (*BookHandler, *http.ServeMux) {
	repo := repository.NewBookRepository()
	svc := service.NewBookService(repo)
	handler := NewBookHandler(svc)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	return handler, mux
}

func TestBookHandler_CreateBook(t *testing.T) {
	_, mux := newTestHandler()

	book := map[string]interface{}{
		"id":        "book-1",
		"title":     "Test Book",
		"isbn":      "978-1234567890",
		"author_id": "author-1",
		"pages":     200,
	}
	body, _ := json.Marshal(book)

	req := httptest.NewRequest(http.MethodPost, "/api/books", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, rec.Code)
	}
}

func TestBookHandler_CreateBook_InvalidJSON(t *testing.T) {
	_, mux := newTestHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/books", bytes.NewReader([]byte("invalid")))
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestBookHandler_CreateBook_MissingFields(t *testing.T) {
	_, mux := newTestHandler()

	book := map[string]interface{}{
		"id": "book-1",
		// Missing required fields
	}
	body, _ := json.Marshal(book)

	req := httptest.NewRequest(http.MethodPost, "/api/books", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestBookHandler_GetBook(t *testing.T) {
	_, mux := newTestHandler()

	// Create a book first
	book := map[string]interface{}{
		"id":        "book-1",
		"title":     "Test Book",
		"isbn":      "978-1234567890",
		"author_id": "author-1",
	}
	body, _ := json.Marshal(book)
	req := httptest.NewRequest(http.MethodPost, "/api/books", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	// Get the book
	req = httptest.NewRequest(http.MethodGet, "/api/books/book-1", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var retrieved model.Book
	json.NewDecoder(rec.Body).Decode(&retrieved)
	if retrieved.Title != "Test Book" {
		t.Errorf("Expected title 'Test Book', got %q", retrieved.Title)
	}
}

func TestBookHandler_GetBook_NotFound(t *testing.T) {
	_, mux := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/books/nonexistent", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestBookHandler_UpdateBook(t *testing.T) {
	_, mux := newTestHandler()

	// Create a book first
	book := map[string]interface{}{
		"id":        "book-1",
		"title":     "Original Title",
		"isbn":      "978-1234567890",
		"author_id": "author-1",
	}
	body, _ := json.Marshal(book)
	req := httptest.NewRequest(http.MethodPost, "/api/books", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	// Update the book
	book["title"] = "Updated Title"
	body, _ = json.Marshal(book)
	req = httptest.NewRequest(http.MethodPut, "/api/books/book-1", bytes.NewReader(body))
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestBookHandler_DeleteBook(t *testing.T) {
	_, mux := newTestHandler()

	// Create a book first
	book := map[string]interface{}{
		"id":        "book-1",
		"title":     "Test Book",
		"isbn":      "978-1234567890",
		"author_id": "author-1",
	}
	body, _ := json.Marshal(book)
	req := httptest.NewRequest(http.MethodPost, "/api/books", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	// Delete the book
	req = httptest.NewRequest(http.MethodDelete, "/api/books/book-1", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, rec.Code)
	}

	// Verify it's gone
	req = httptest.NewRequest(http.MethodGet, "/api/books/book-1", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status %d after delete, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestBookHandler_ListBooks(t *testing.T) {
	_, mux := newTestHandler()

	// Create two books
	for i, isbn := range []string{"isbn-1", "isbn-2"} {
		book := map[string]interface{}{
			"id":        string(rune('a' + i)),
			"title":     "Book",
			"isbn":      isbn,
			"author_id": "author-1",
		}
		body, _ := json.Marshal(book)
		req := httptest.NewRequest(http.MethodPost, "/api/books", bytes.NewReader(body))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
	}

	// List books
	req := httptest.NewRequest(http.MethodGet, "/api/books", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var books []model.Book
	json.NewDecoder(rec.Body).Decode(&books)
	if len(books) != 2 {
		t.Errorf("Expected 2 books, got %d", len(books))
	}
}

func TestBookHandler_MethodNotAllowed(t *testing.T) {
	_, mux := newTestHandler()

	req := httptest.NewRequest(http.MethodPatch, "/api/books", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
	}
}
