package service

import (
	"testing"

	"github.com/pawelpaszki/gorts-demo/internal/model"
	"github.com/pawelpaszki/gorts-demo/internal/repository"
)

func newTestBookService() *BookService {
	repo := repository.NewBookRepository()
	return NewBookService(repo)
}

func validBook(id string) *model.Book {
	return &model.Book{
		ID:       id,
		Title:    "Test Book",
		ISBN:     "978-" + id,
		AuthorID: "author-1",
		Pages:    200,
	}
}

func TestBookService_CreateBook(t *testing.T) {
	svc := newTestBookService()
	book := validBook("book-1")

	err := svc.CreateBook(book)
	if err != nil {
		t.Fatalf("CreateBook failed: %v", err)
	}

	if svc.GetBookCount() != 1 {
		t.Errorf("Expected 1 book, got %d", svc.GetBookCount())
	}
}

func TestBookService_CreateBook_InvalidData(t *testing.T) {
	svc := newTestBookService()
	book := &model.Book{
		ID: "book-1",
		// Missing required fields
	}

	err := svc.CreateBook(book)
	if err == nil {
		t.Error("Expected error for invalid book")
	}
}

func TestBookService_CreateBook_DuplicateISBN(t *testing.T) {
	svc := newTestBookService()

	book1 := validBook("book-1")
	book1.ISBN = "same-isbn"
	_ = svc.CreateBook(book1)

	book2 := validBook("book-2")
	book2.ISBN = "same-isbn"
	err := svc.CreateBook(book2)

	if err != ErrDuplicateISBN {
		t.Errorf("Expected ErrDuplicateISBN, got %v", err)
	}
}

func TestBookService_GetBook(t *testing.T) {
	svc := newTestBookService()
	original := validBook("book-1")
	_ = svc.CreateBook(original)

	retrieved, err := svc.GetBook("book-1")
	if err != nil {
		t.Fatalf("GetBook failed: %v", err)
	}

	if retrieved.Title != original.Title {
		t.Errorf("Expected title %q, got %q", original.Title, retrieved.Title)
	}
}

func TestBookService_GetBook_NotFound(t *testing.T) {
	svc := newTestBookService()

	_, err := svc.GetBook("nonexistent")
	if err != ErrBookNotFound {
		t.Errorf("Expected ErrBookNotFound, got %v", err)
	}
}

func TestBookService_UpdateBook(t *testing.T) {
	svc := newTestBookService()
	book := validBook("book-1")
	_ = svc.CreateBook(book)

	book.Title = "Updated Title"
	err := svc.UpdateBook(book)
	if err != nil {
		t.Fatalf("UpdateBook failed: %v", err)
	}

	retrieved, _ := svc.GetBook("book-1")
	if retrieved.Title != "Updated Title" {
		t.Errorf("Expected updated title, got %q", retrieved.Title)
	}
}

func TestBookService_UpdateBook_DuplicateISBN(t *testing.T) {
	svc := newTestBookService()

	book1 := validBook("book-1")
	book1.ISBN = "isbn-1"
	_ = svc.CreateBook(book1)

	book2 := validBook("book-2")
	book2.ISBN = "isbn-2"
	_ = svc.CreateBook(book2)

	// Try to update book2 with book1's ISBN
	book2.ISBN = "isbn-1"
	err := svc.UpdateBook(book2)

	if err != ErrDuplicateISBN {
		t.Errorf("Expected ErrDuplicateISBN, got %v", err)
	}
}

func TestBookService_DeleteBook(t *testing.T) {
	svc := newTestBookService()
	book := validBook("book-1")
	_ = svc.CreateBook(book)

	err := svc.DeleteBook("book-1")
	if err != nil {
		t.Fatalf("DeleteBook failed: %v", err)
	}

	if svc.GetBookCount() != 0 {
		t.Error("Book should be deleted")
	}
}

func TestBookService_DeleteBook_NotFound(t *testing.T) {
	svc := newTestBookService()

	err := svc.DeleteBook("nonexistent")
	if err != ErrBookNotFound {
		t.Errorf("Expected ErrBookNotFound, got %v", err)
	}
}

func TestBookService_ListBooks(t *testing.T) {
	svc := newTestBookService()

	for i := 0; i < 3; i++ {
		book := validBook(string(rune('a' + i)))
		_ = svc.CreateBook(book)
	}

	books := svc.ListBooks()
	if len(books) != 3 {
		t.Errorf("Expected 3 books, got %d", len(books))
	}
}

func TestBookService_GetBooksByAuthor(t *testing.T) {
	svc := newTestBookService()

	book1 := validBook("book-1")
	book1.AuthorID = "author-1"
	_ = svc.CreateBook(book1)

	book2 := validBook("book-2")
	book2.AuthorID = "author-1"
	_ = svc.CreateBook(book2)

	book3 := validBook("book-3")
	book3.AuthorID = "author-2"
	_ = svc.CreateBook(book3)

	books := svc.GetBooksByAuthor("author-1")
	if len(books) != 2 {
		t.Errorf("Expected 2 books, got %d", len(books))
	}
}
