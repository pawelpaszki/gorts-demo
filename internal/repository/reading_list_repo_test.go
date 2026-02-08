package repository

import (
	"testing"

	"github.com/pawelpaszki/gorts-demo/internal/model"
)

func TestReadingListRepository_Create(t *testing.T) {
	repo := NewReadingListRepository()

	list := &model.ReadingList{
		ID:   "list-1",
		Name: "My Reading List",
	}

	err := repo.Create(list)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if repo.Count() != 1 {
		t.Errorf("Expected count 1, got %d", repo.Count())
	}
}

func TestReadingListRepository_Create_Duplicate(t *testing.T) {
	repo := NewReadingListRepository()

	list := &model.ReadingList{
		ID:   "list-1",
		Name: "My List",
	}

	_ = repo.Create(list)
	err := repo.Create(list)

	if err != ErrReadingListExists {
		t.Errorf("Expected ErrReadingListExists, got %v", err)
	}
}

func TestReadingListRepository_Get(t *testing.T) {
	repo := NewReadingListRepository()

	original := &model.ReadingList{
		ID:      "list-1",
		Name:    "My List",
		BookIDs: []string{"book-1", "book-2"},
	}
	_ = repo.Create(original)

	retrieved, err := repo.Get("list-1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.Name != original.Name {
		t.Errorf("Expected name %q, got %q", original.Name, retrieved.Name)
	}
	if len(retrieved.BookIDs) != 2 {
		t.Errorf("Expected 2 books, got %d", len(retrieved.BookIDs))
	}
}

func TestReadingListRepository_Get_NotFound(t *testing.T) {
	repo := NewReadingListRepository()

	_, err := repo.Get("nonexistent")
	if err != ErrReadingListNotFound {
		t.Errorf("Expected ErrReadingListNotFound, got %v", err)
	}
}

func TestReadingListRepository_Update(t *testing.T) {
	repo := NewReadingListRepository()

	list := &model.ReadingList{
		ID:   "list-1",
		Name: "Original Name",
	}
	_ = repo.Create(list)

	updated := &model.ReadingList{
		ID:      "list-1",
		Name:    "Updated Name",
		BookIDs: []string{"book-1"},
	}
	err := repo.Update(updated)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	retrieved, _ := repo.Get("list-1")
	if retrieved.Name != "Updated Name" {
		t.Errorf("Expected updated name, got %q", retrieved.Name)
	}
	if len(retrieved.BookIDs) != 1 {
		t.Errorf("Expected 1 book, got %d", len(retrieved.BookIDs))
	}
}

func TestReadingListRepository_Delete(t *testing.T) {
	repo := NewReadingListRepository()

	list := &model.ReadingList{
		ID:   "list-1",
		Name: "My List",
	}
	_ = repo.Create(list)

	err := repo.Delete("list-1")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if repo.Count() != 0 {
		t.Error("List should be deleted")
	}
}

func TestReadingListRepository_List(t *testing.T) {
	repo := NewReadingListRepository()

	for i := 0; i < 3; i++ {
		list := &model.ReadingList{
			ID:   string(rune('a' + i)),
			Name: "List",
		}
		_ = repo.Create(list)
	}

	lists := repo.List()
	if len(lists) != 3 {
		t.Errorf("Expected 3 lists, got %d", len(lists))
	}
}

func TestReadingListRepository_FindByBook(t *testing.T) {
	repo := NewReadingListRepository()

	_ = repo.Create(&model.ReadingList{ID: "1", Name: "List 1", BookIDs: []string{"book-1", "book-2"}})
	_ = repo.Create(&model.ReadingList{ID: "2", Name: "List 2", BookIDs: []string{"book-1"}})
	_ = repo.Create(&model.ReadingList{ID: "3", Name: "List 3", BookIDs: []string{"book-3"}})

	lists := repo.FindByBook("book-1")
	if len(lists) != 2 {
		t.Errorf("Expected 2 lists containing book-1, got %d", len(lists))
	}
}
