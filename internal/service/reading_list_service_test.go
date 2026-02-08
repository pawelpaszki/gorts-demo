package service

import (
	"testing"

	"github.com/pawelpaszki/gorts-demo/internal/model"
	"github.com/pawelpaszki/gorts-demo/internal/repository"
)

func newTestReadingListService() (*ReadingListService, *repository.BookRepository) {
	listRepo := repository.NewReadingListRepository()
	bookRepo := repository.NewBookRepository()
	return NewReadingListService(listRepo, bookRepo), bookRepo
}

func validReadingList(id string) *model.ReadingList {
	return &model.ReadingList{
		ID:          id,
		Name:        "Test Reading List",
		Description: "A test list",
	}
}

func TestReadingListService_CreateReadingList(t *testing.T) {
	svc, _ := newTestReadingListService()
	list := validReadingList("list-1")

	err := svc.CreateReadingList(list)
	if err != nil {
		t.Fatalf("CreateReadingList failed: %v", err)
	}

	if svc.GetReadingListCount() != 1 {
		t.Errorf("Expected 1 list, got %d", svc.GetReadingListCount())
	}
}

func TestReadingListService_CreateReadingList_InvalidData(t *testing.T) {
	svc, _ := newTestReadingListService()
	list := &model.ReadingList{
		ID: "list-1",
		// Missing required Name
	}

	err := svc.CreateReadingList(list)
	if err == nil {
		t.Error("Expected error for invalid list")
	}
}

func TestReadingListService_GetReadingList(t *testing.T) {
	svc, _ := newTestReadingListService()
	original := validReadingList("list-1")
	_ = svc.CreateReadingList(original)

	retrieved, err := svc.GetReadingList("list-1")
	if err != nil {
		t.Fatalf("GetReadingList failed: %v", err)
	}

	if retrieved.Name != original.Name {
		t.Errorf("Expected name %q, got %q", original.Name, retrieved.Name)
	}
}

func TestReadingListService_GetReadingList_NotFound(t *testing.T) {
	svc, _ := newTestReadingListService()

	_, err := svc.GetReadingList("nonexistent")
	if err != ErrReadingListNotFound {
		t.Errorf("Expected ErrReadingListNotFound, got %v", err)
	}
}

func TestReadingListService_AddBookToList(t *testing.T) {
	svc, bookRepo := newTestReadingListService()

	// Create a book first
	book := &model.Book{
		ID:       "book-1",
		Title:    "Test Book",
		ISBN:     "123",
		AuthorID: "author-1",
	}
	_ = bookRepo.Create(book)

	// Create a reading list
	list := validReadingList("list-1")
	_ = svc.CreateReadingList(list)

	// Add book to list
	err := svc.AddBookToList("list-1", "book-1")
	if err != nil {
		t.Fatalf("AddBookToList failed: %v", err)
	}

	// Verify book is in list
	retrieved, _ := svc.GetReadingList("list-1")
	if !retrieved.ContainsBook("book-1") {
		t.Error("Book should be in list")
	}
}

func TestReadingListService_AddBookToList_BookNotFound(t *testing.T) {
	svc, _ := newTestReadingListService()

	list := validReadingList("list-1")
	_ = svc.CreateReadingList(list)

	err := svc.AddBookToList("list-1", "nonexistent-book")
	if err != ErrBookNotFound {
		t.Errorf("Expected ErrBookNotFound, got %v", err)
	}
}

func TestReadingListService_AddBookToList_AlreadyInList(t *testing.T) {
	svc, bookRepo := newTestReadingListService()

	book := &model.Book{ID: "book-1", Title: "Test", ISBN: "123", AuthorID: "a"}
	_ = bookRepo.Create(book)

	list := validReadingList("list-1")
	_ = svc.CreateReadingList(list)
	_ = svc.AddBookToList("list-1", "book-1")

	err := svc.AddBookToList("list-1", "book-1")
	if err != ErrBookAlreadyInList {
		t.Errorf("Expected ErrBookAlreadyInList, got %v", err)
	}
}

func TestReadingListService_RemoveBookFromList(t *testing.T) {
	svc, bookRepo := newTestReadingListService()

	book := &model.Book{ID: "book-1", Title: "Test", ISBN: "123", AuthorID: "a"}
	_ = bookRepo.Create(book)

	list := validReadingList("list-1")
	_ = svc.CreateReadingList(list)
	_ = svc.AddBookToList("list-1", "book-1")

	err := svc.RemoveBookFromList("list-1", "book-1")
	if err != nil {
		t.Fatalf("RemoveBookFromList failed: %v", err)
	}

	retrieved, _ := svc.GetReadingList("list-1")
	if retrieved.ContainsBook("book-1") {
		t.Error("Book should be removed from list")
	}
}

func TestReadingListService_RemoveBookFromList_NotInList(t *testing.T) {
	svc, _ := newTestReadingListService()

	list := validReadingList("list-1")
	_ = svc.CreateReadingList(list)

	err := svc.RemoveBookFromList("list-1", "book-1")
	if err != ErrBookNotInList {
		t.Errorf("Expected ErrBookNotInList, got %v", err)
	}
}

func TestReadingListService_DeleteReadingList(t *testing.T) {
	svc, _ := newTestReadingListService()
	list := validReadingList("list-1")
	_ = svc.CreateReadingList(list)

	err := svc.DeleteReadingList("list-1")
	if err != nil {
		t.Fatalf("DeleteReadingList failed: %v", err)
	}

	if svc.GetReadingListCount() != 0 {
		t.Error("List should be deleted")
	}
}

func TestReadingListService_GetListsContainingBook(t *testing.T) {
	svc, bookRepo := newTestReadingListService()

	// Create books
	_ = bookRepo.Create(&model.Book{ID: "book-1", Title: "Book 1", ISBN: "1", AuthorID: "a"})
	_ = bookRepo.Create(&model.Book{ID: "book-2", Title: "Book 2", ISBN: "2", AuthorID: "a"})

	// Create lists and add books
	list1 := validReadingList("list-1")
	list2 := validReadingList("list-2")
	list3 := validReadingList("list-3")
	_ = svc.CreateReadingList(list1)
	_ = svc.CreateReadingList(list2)
	_ = svc.CreateReadingList(list3)

	_ = svc.AddBookToList("list-1", "book-1")
	_ = svc.AddBookToList("list-2", "book-1")
	_ = svc.AddBookToList("list-3", "book-2")

	lists := svc.GetListsContainingBook("book-1")
	if len(lists) != 2 {
		t.Errorf("Expected 2 lists containing book-1, got %d", len(lists))
	}
}
