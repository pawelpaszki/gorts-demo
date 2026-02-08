package e2e

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/pawelpaszki/gorts-demo/internal/model"
)

func TestE2E_BookCRUD_FullLifecycle(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()

	client := &http.Client{}

	// ==================== CREATE ====================
	t.Log("Step 1: Create a book")
	createData := map[string]interface{}{
		"id":        "crud-book-1",
		"title":     "Original Title",
		"isbn":      "978-0000000001",
		"author_id": "author-1",
		"pages":     100,
		"genre":     "Fiction",
	}
	body, _ := json.Marshal(createData)

	resp, err := client.Post(ts.URL()+"/api/books", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("Create: expected %d, got %d: %s", http.StatusCreated, resp.StatusCode, string(bodyBytes))
	}

	var createdBook model.Book
	json.NewDecoder(resp.Body).Decode(&createdBook)
	resp.Body.Close()

	if createdBook.Title != "Original Title" {
		t.Errorf("Created book title = %q, want %q", createdBook.Title, "Original Title")
	}
	if createdBook.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set after create")
	}

	// ==================== READ ====================
	t.Log("Step 2: Read the book")
	resp, err = client.Get(ts.URL() + "/api/books/crud-book-1")
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Read: expected %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var readBook model.Book
	json.NewDecoder(resp.Body).Decode(&readBook)
	resp.Body.Close()

	if readBook.ID != "crud-book-1" {
		t.Errorf("Read book ID = %q, want %q", readBook.ID, "crud-book-1")
	}
	if readBook.Pages != 100 {
		t.Errorf("Read book pages = %d, want %d", readBook.Pages, 100)
	}

	// ==================== UPDATE ====================
	t.Log("Step 3: Update the book")
	updateData := map[string]interface{}{
		"id":        "crud-book-1",
		"title":     "Updated Title",
		"isbn":      "978-0000000001",
		"author_id": "author-1",
		"pages":     200,
		"genre":     "Non-Fiction",
	}
	body, _ = json.Marshal(updateData)

	req, _ := http.NewRequest(http.MethodPut, ts.URL()+"/api/books/crud-book-1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("Update: expected %d, got %d: %s", http.StatusOK, resp.StatusCode, string(bodyBytes))
	}

	var updatedBook model.Book
	json.NewDecoder(resp.Body).Decode(&updatedBook)
	resp.Body.Close()

	if updatedBook.Title != "Updated Title" {
		t.Errorf("Updated book title = %q, want %q", updatedBook.Title, "Updated Title")
	}
	if updatedBook.Pages != 200 {
		t.Errorf("Updated book pages = %d, want %d", updatedBook.Pages, 200)
	}

	// ==================== VERIFY UPDATE ====================
	t.Log("Step 4: Verify update persisted")
	resp, _ = client.Get(ts.URL() + "/api/books/crud-book-1")
	var verifyBook model.Book
	json.NewDecoder(resp.Body).Decode(&verifyBook)
	resp.Body.Close()

	if verifyBook.Title != "Updated Title" {
		t.Errorf("Verified book title = %q, want %q", verifyBook.Title, "Updated Title")
	}
	if verifyBook.Genre != "Non-Fiction" {
		t.Errorf("Verified book genre = %q, want %q", verifyBook.Genre, "Non-Fiction")
	}

	// ==================== LIST ====================
	t.Log("Step 5: List all books")
	resp, _ = client.Get(ts.URL() + "/api/books")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("List: expected %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var books []model.Book
	json.NewDecoder(resp.Body).Decode(&books)
	resp.Body.Close()

	if len(books) != 1 {
		t.Errorf("List should have 1 book, got %d", len(books))
	}

	// ==================== DELETE ====================
	t.Log("Step 6: Delete the book")
	req, _ = http.NewRequest(http.MethodDelete, ts.URL()+"/api/books/crud-book-1", nil)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("Delete: expected %d, got %d", http.StatusNoContent, resp.StatusCode)
	}
	resp.Body.Close()

	// ==================== VERIFY DELETE ====================
	t.Log("Step 7: Verify deletion")
	resp, _ = client.Get(ts.URL() + "/api/books/crud-book-1")
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("After delete: expected %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
	resp.Body.Close()

	// Verify list is empty
	resp, _ = client.Get(ts.URL() + "/api/books")
	json.NewDecoder(resp.Body).Decode(&books)
	resp.Body.Close()

	if len(books) != 0 {
		t.Errorf("List should be empty after delete, got %d", len(books))
	}
}

func TestE2E_BookCRUD_MultipleBooks(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()

	client := &http.Client{}

	// Create multiple books
	booksToCreate := []map[string]interface{}{
		{"id": "multi-1", "title": "Book One", "isbn": "isbn-1", "author_id": "author-1", "pages": 100},
		{"id": "multi-2", "title": "Book Two", "isbn": "isbn-2", "author_id": "author-1", "pages": 200},
		{"id": "multi-3", "title": "Book Three", "isbn": "isbn-3", "author_id": "author-2", "pages": 300},
		{"id": "multi-4", "title": "Book Four", "isbn": "isbn-4", "author_id": "author-2", "pages": 400},
		{"id": "multi-5", "title": "Book Five", "isbn": "isbn-5", "author_id": "author-3", "pages": 500},
	}

	for _, bookData := range booksToCreate {
		body, _ := json.Marshal(bookData)
		resp, err := client.Post(ts.URL()+"/api/books", "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatalf("Failed to create book: %v", err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("Expected %d, got %d", http.StatusCreated, resp.StatusCode)
		}
		resp.Body.Close()
	}

	// List and verify count
	resp, _ := client.Get(ts.URL() + "/api/books")
	var books []model.Book
	json.NewDecoder(resp.Body).Decode(&books)
	resp.Body.Close()

	if len(books) != 5 {
		t.Errorf("Expected 5 books, got %d", len(books))
	}

	// Update one book
	updateData := map[string]interface{}{
		"id":        "multi-3",
		"title":     "Book Three Updated",
		"isbn":      "isbn-3",
		"author_id": "author-2",
		"pages":     350,
	}
	body, _ := json.Marshal(updateData)
	req, _ := http.NewRequest(http.MethodPut, ts.URL()+"/api/books/multi-3", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ = client.Do(req)
	resp.Body.Close()

	// Delete two books
	req, _ = http.NewRequest(http.MethodDelete, ts.URL()+"/api/books/multi-2", nil)
	resp, _ = client.Do(req)
	resp.Body.Close()

	req, _ = http.NewRequest(http.MethodDelete, ts.URL()+"/api/books/multi-4", nil)
	resp, _ = client.Do(req)
	resp.Body.Close()

	// Verify final count
	resp, _ = client.Get(ts.URL() + "/api/books")
	json.NewDecoder(resp.Body).Decode(&books)
	resp.Body.Close()

	if len(books) != 3 {
		t.Errorf("Expected 3 books after deletes, got %d", len(books))
	}

	// Verify updated book
	resp, _ = client.Get(ts.URL() + "/api/books/multi-3")
	var updatedBook model.Book
	json.NewDecoder(resp.Body).Decode(&updatedBook)
	resp.Body.Close()

	if updatedBook.Title != "Book Three Updated" {
		t.Errorf("Expected updated title, got %q", updatedBook.Title)
	}
	if updatedBook.Pages != 350 {
		t.Errorf("Expected 350 pages, got %d", updatedBook.Pages)
	}
}

func TestE2E_BookCRUD_UpdateNonExistent(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()

	client := &http.Client{}

	updateData := map[string]interface{}{
		"id":        "non-existent",
		"title":     "Ghost Book",
		"isbn":      "ghost-isbn",
		"author_id": "author-1",
	}
	body, _ := json.Marshal(updateData)

	req, _ := http.NewRequest(http.MethodPut, ts.URL()+"/api/books/non-existent", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
}

func TestE2E_BookCRUD_DeleteNonExistent(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()

	client := &http.Client{}

	req, _ := http.NewRequest(http.MethodDelete, ts.URL()+"/api/books/non-existent", nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
}

func TestE2E_BookCRUD_UpdateWithInvalidData(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()

	client := &http.Client{}

	// First create a valid book
	createData := map[string]interface{}{
		"id":        "update-test",
		"title":     "Valid Book",
		"isbn":      "valid-isbn",
		"author_id": "author-1",
	}
	body, _ := json.Marshal(createData)
	resp, _ := client.Post(ts.URL()+"/api/books", "application/json", bytes.NewReader(body))
	resp.Body.Close()

	// Try to update with invalid data (missing title)
	updateData := map[string]interface{}{
		"id":        "update-test",
		"title":     "", // Invalid: empty title
		"isbn":      "valid-isbn",
		"author_id": "author-1",
	}
	body, _ = json.Marshal(updateData)

	req, _ := http.NewRequest(http.MethodPut, ts.URL()+"/api/books/update-test", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}
