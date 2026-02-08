package integration

import (
	"testing"
	"time"

	"github.com/pawelpaszki/gorts-demo/internal/model"
	"github.com/pawelpaszki/gorts-demo/internal/repository"
	"github.com/pawelpaszki/gorts-demo/internal/service"
)

// TestBookServiceIntegration tests the book service with a real repository.
func TestBookServiceIntegration(t *testing.T) {
	repo := repository.NewBookRepository()
	svc := service.NewBookService(repo)

	t.Run("full CRUD lifecycle", func(t *testing.T) {
		// Create
		book := &model.Book{
			ID:          "integration-book-1",
			Title:       "Integration Testing in Go",
			ISBN:        "978-1234567890",
			AuthorID:    "author-1",
			Pages:       350,
			Genre:       "Technology",
			PublishedAt: time.Now(),
		}

		err := svc.CreateBook(book)
		if err != nil {
			t.Fatalf("CreateBook failed: %v", err)
		}

		// Read
		retrieved, err := svc.GetBook("integration-book-1")
		if err != nil {
			t.Fatalf("GetBook failed: %v", err)
		}
		if retrieved.Title != book.Title {
			t.Errorf("Title mismatch: got %q, want %q", retrieved.Title, book.Title)
		}
		if retrieved.CreatedAt.IsZero() {
			t.Error("CreatedAt should be set")
		}

		// Update
		retrieved.Title = "Updated Integration Testing"
		retrieved.Pages = 400
		err = svc.UpdateBook(retrieved)
		if err != nil {
			t.Fatalf("UpdateBook failed: %v", err)
		}

		updated, _ := svc.GetBook("integration-book-1")
		if updated.Title != "Updated Integration Testing" {
			t.Errorf("Title not updated: got %q", updated.Title)
		}
		if updated.Pages != 400 {
			t.Errorf("Pages not updated: got %d", updated.Pages)
		}

		// Delete
		err = svc.DeleteBook("integration-book-1")
		if err != nil {
			t.Fatalf("DeleteBook failed: %v", err)
		}

		_, err = svc.GetBook("integration-book-1")
		if err != service.ErrBookNotFound {
			t.Errorf("Expected ErrBookNotFound after delete, got %v", err)
		}
	})
}

func TestBookServiceIntegration_MultipleBooks(t *testing.T) {
	repo := repository.NewBookRepository()
	svc := service.NewBookService(repo)

	// Create multiple books
	books := []*model.Book{
		{ID: "book-1", Title: "Book One", ISBN: "isbn-1", AuthorID: "author-1", Pages: 100},
		{ID: "book-2", Title: "Book Two", ISBN: "isbn-2", AuthorID: "author-1", Pages: 200},
		{ID: "book-3", Title: "Book Three", ISBN: "isbn-3", AuthorID: "author-2", Pages: 300},
		{ID: "book-4", Title: "Book Four", ISBN: "isbn-4", AuthorID: "author-2", Pages: 400},
		{ID: "book-5", Title: "Book Five", ISBN: "isbn-5", AuthorID: "author-3", Pages: 500},
	}

	for _, book := range books {
		if err := svc.CreateBook(book); err != nil {
			t.Fatalf("Failed to create %s: %v", book.ID, err)
		}
	}

	// Verify count
	if count := svc.GetBookCount(); count != 5 {
		t.Errorf("Expected 5 books, got %d", count)
	}

	// List all
	allBooks := svc.ListBooks()
	if len(allBooks) != 5 {
		t.Errorf("Expected 5 books in list, got %d", len(allBooks))
	}

	// Filter by author
	author1Books := svc.GetBooksByAuthor("author-1")
	if len(author1Books) != 2 {
		t.Errorf("Expected 2 books by author-1, got %d", len(author1Books))
	}

	author2Books := svc.GetBooksByAuthor("author-2")
	if len(author2Books) != 2 {
		t.Errorf("Expected 2 books by author-2, got %d", len(author2Books))
	}

	author3Books := svc.GetBooksByAuthor("author-3")
	if len(author3Books) != 1 {
		t.Errorf("Expected 1 book by author-3, got %d", len(author3Books))
	}

	// Delete some books
	_ = svc.DeleteBook("book-2")
	_ = svc.DeleteBook("book-4")

	if count := svc.GetBookCount(); count != 3 {
		t.Errorf("Expected 3 books after delete, got %d", count)
	}
}

func TestBookServiceIntegration_ISBNUniqueness(t *testing.T) {
	repo := repository.NewBookRepository()
	svc := service.NewBookService(repo)

	// Create first book
	book1 := &model.Book{
		ID:       "book-1",
		Title:    "First Book",
		ISBN:     "unique-isbn-123",
		AuthorID: "author-1",
	}
	if err := svc.CreateBook(book1); err != nil {
		t.Fatalf("Failed to create first book: %v", err)
	}

	// Try to create second book with same ISBN
	book2 := &model.Book{
		ID:       "book-2",
		Title:    "Second Book",
		ISBN:     "unique-isbn-123", // Same ISBN
		AuthorID: "author-2",
	}
	err := svc.CreateBook(book2)
	if err != service.ErrDuplicateISBN {
		t.Errorf("Expected ErrDuplicateISBN, got %v", err)
	}

	// Create with different ISBN should work
	book2.ISBN = "unique-isbn-456"
	if err := svc.CreateBook(book2); err != nil {
		t.Fatalf("Failed to create book with unique ISBN: %v", err)
	}

	// Update book2 to use book1's ISBN should fail
	book2.ISBN = "unique-isbn-123"
	err = svc.UpdateBook(book2)
	if err != service.ErrDuplicateISBN {
		t.Errorf("Expected ErrDuplicateISBN on update, got %v", err)
	}
}

func TestBookServiceIntegration_ValidationErrors(t *testing.T) {
	repo := repository.NewBookRepository()
	svc := service.NewBookService(repo)

	tests := []struct {
		name string
		book *model.Book
	}{
		{
			name: "missing title",
			book: &model.Book{ID: "b1", ISBN: "123", AuthorID: "a1"},
		},
		{
			name: "missing ISBN",
			book: &model.Book{ID: "b2", Title: "Test", AuthorID: "a1"},
		},
		{
			name: "missing author",
			book: &model.Book{ID: "b3", Title: "Test", ISBN: "123"},
		},
		{
			name: "negative pages",
			book: &model.Book{ID: "b4", Title: "Test", ISBN: "123", AuthorID: "a1", Pages: -10},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.CreateBook(tt.book)
			if err == nil {
				t.Error("Expected validation error")
			}
		})
	}
}

func TestBookServiceIntegration_ConcurrentAccess(t *testing.T) {
	repo := repository.NewBookRepository()
	svc := service.NewBookService(repo)

	// Create initial book
	book := &model.Book{
		ID:       "concurrent-book",
		Title:    "Concurrent Access Test",
		ISBN:     "concurrent-isbn",
		AuthorID: "author-1",
	}
	_ = svc.CreateBook(book)

	// Concurrent reads
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_, _ = svc.GetBook("concurrent-book")
				_ = svc.ListBooks()
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify data integrity
	retrieved, err := svc.GetBook("concurrent-book")
	if err != nil {
		t.Fatalf("Book should still exist: %v", err)
	}
	if retrieved.Title != "Concurrent Access Test" {
		t.Error("Book data corrupted")
	}
}
