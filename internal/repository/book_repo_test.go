package repository

import (
	"testing"
	"time"

	"github.com/pawelpaszki/gorts-demo/internal/model"
)

func TestBookRepository_Create(t *testing.T) {
	repo := NewBookRepository()

	book := &model.Book{
		ID:       "book-1",
		Title:    "The Go Programming Language",
		ISBN:     "978-0134190440",
		AuthorID: "author-1",
		Pages:    400,
	}

	err := repo.Create(book)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if repo.Count() != 1 {
		t.Errorf("Expected count 1, got %d", repo.Count())
	}

	// Verify timestamps were set
	if book.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
}

func TestBookRepository_Create_Duplicate(t *testing.T) {
	repo := NewBookRepository()

	book := &model.Book{
		ID:       "book-1",
		Title:    "Test Book",
		ISBN:     "123",
		AuthorID: "author-1",
	}

	_ = repo.Create(book)
	err := repo.Create(book)

	if err != ErrBookExists {
		t.Errorf("Expected ErrBookExists, got %v", err)
	}
}

func TestBookRepository_Get(t *testing.T) {
	repo := NewBookRepository()

	original := &model.Book{
		ID:       "book-1",
		Title:    "Test Book",
		ISBN:     "123",
		AuthorID: "author-1",
		Pages:    100,
	}
	_ = repo.Create(original)

	retrieved, err := repo.Get("book-1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.Title != original.Title {
		t.Errorf("Expected title %q, got %q", original.Title, retrieved.Title)
	}
}

func TestBookRepository_Get_NotFound(t *testing.T) {
	repo := NewBookRepository()

	_, err := repo.Get("nonexistent")
	if err != ErrBookNotFound {
		t.Errorf("Expected ErrBookNotFound, got %v", err)
	}
}

func TestBookRepository_Update(t *testing.T) {
	repo := NewBookRepository()

	book := &model.Book{
		ID:       "book-1",
		Title:    "Original Title",
		ISBN:     "123",
		AuthorID: "author-1",
	}
	_ = repo.Create(book)
	originalCreatedAt := book.CreatedAt

	time.Sleep(10 * time.Millisecond) // Ensure different timestamp

	updated := &model.Book{
		ID:       "book-1",
		Title:    "Updated Title",
		ISBN:     "123",
		AuthorID: "author-1",
	}
	err := repo.Update(updated)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	retrieved, _ := repo.Get("book-1")
	if retrieved.Title != "Updated Title" {
		t.Errorf("Expected updated title, got %q", retrieved.Title)
	}
	if retrieved.CreatedAt != originalCreatedAt {
		t.Error("CreatedAt should not change on update")
	}
	if !retrieved.UpdatedAt.After(originalCreatedAt) {
		t.Error("UpdatedAt should be after CreatedAt")
	}
}

func TestBookRepository_Update_NotFound(t *testing.T) {
	repo := NewBookRepository()

	book := &model.Book{ID: "nonexistent"}
	err := repo.Update(book)

	if err != ErrBookNotFound {
		t.Errorf("Expected ErrBookNotFound, got %v", err)
	}
}

func TestBookRepository_Delete(t *testing.T) {
	repo := NewBookRepository()

	book := &model.Book{
		ID:       "book-1",
		Title:    "Test",
		ISBN:     "123",
		AuthorID: "author-1",
	}
	_ = repo.Create(book)

	err := repo.Delete("book-1")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if repo.Count() != 0 {
		t.Error("Book should be deleted")
	}
}

func TestBookRepository_Delete_NotFound(t *testing.T) {
	repo := NewBookRepository()

	err := repo.Delete("nonexistent")
	if err != ErrBookNotFound {
		t.Errorf("Expected ErrBookNotFound, got %v", err)
	}
}

func TestBookRepository_List(t *testing.T) {
	repo := NewBookRepository()

	for i := 0; i < 3; i++ {
		book := &model.Book{
			ID:       string(rune('a' + i)),
			Title:    "Book",
			ISBN:     "123",
			AuthorID: "author-1",
		}
		_ = repo.Create(book)
	}

	books := repo.List()
	if len(books) != 3 {
		t.Errorf("Expected 3 books, got %d", len(books))
	}
}

func TestBookRepository_FindByAuthor(t *testing.T) {
	repo := NewBookRepository()

	// Create books by different authors
	_ = repo.Create(&model.Book{ID: "1", Title: "Book 1", ISBN: "1", AuthorID: "author-1"})
	_ = repo.Create(&model.Book{ID: "2", Title: "Book 2", ISBN: "2", AuthorID: "author-1"})
	_ = repo.Create(&model.Book{ID: "3", Title: "Book 3", ISBN: "3", AuthorID: "author-2"})

	books := repo.FindByAuthor("author-1")
	if len(books) != 2 {
		t.Errorf("Expected 2 books by author-1, got %d", len(books))
	}
}
