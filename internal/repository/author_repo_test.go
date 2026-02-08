package repository

import (
	"testing"

	"github.com/pawelpaszki/gorts-demo/internal/model"
)

func TestAuthorRepository_Create(t *testing.T) {
	repo := NewAuthorRepository()

	author := &model.Author{
		ID:      "author-1",
		Name:    "Jane Doe",
		Country: "USA",
	}

	err := repo.Create(author)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if repo.Count() != 1 {
		t.Errorf("Expected count 1, got %d", repo.Count())
	}
}

func TestAuthorRepository_Create_Duplicate(t *testing.T) {
	repo := NewAuthorRepository()

	author := &model.Author{
		ID:   "author-1",
		Name: "Jane Doe",
	}

	_ = repo.Create(author)
	err := repo.Create(author)

	if err != ErrAuthorExists {
		t.Errorf("Expected ErrAuthorExists, got %v", err)
	}
}

func TestAuthorRepository_Get(t *testing.T) {
	repo := NewAuthorRepository()

	original := &model.Author{
		ID:      "author-1",
		Name:    "Jane Doe",
		Country: "USA",
	}
	_ = repo.Create(original)

	retrieved, err := repo.Get("author-1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.Name != original.Name {
		t.Errorf("Expected name %q, got %q", original.Name, retrieved.Name)
	}
}

func TestAuthorRepository_Get_NotFound(t *testing.T) {
	repo := NewAuthorRepository()

	_, err := repo.Get("nonexistent")
	if err != ErrAuthorNotFound {
		t.Errorf("Expected ErrAuthorNotFound, got %v", err)
	}
}

func TestAuthorRepository_Update(t *testing.T) {
	repo := NewAuthorRepository()

	author := &model.Author{
		ID:   "author-1",
		Name: "Jane Doe",
	}
	_ = repo.Create(author)

	updated := &model.Author{
		ID:   "author-1",
		Name: "Jane Smith",
	}
	err := repo.Update(updated)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	retrieved, _ := repo.Get("author-1")
	if retrieved.Name != "Jane Smith" {
		t.Errorf("Expected updated name, got %q", retrieved.Name)
	}
}

func TestAuthorRepository_Delete(t *testing.T) {
	repo := NewAuthorRepository()

	author := &model.Author{
		ID:   "author-1",
		Name: "Jane Doe",
	}
	_ = repo.Create(author)

	err := repo.Delete("author-1")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if repo.Count() != 0 {
		t.Error("Author should be deleted")
	}
}

func TestAuthorRepository_List(t *testing.T) {
	repo := NewAuthorRepository()

	for i := 0; i < 3; i++ {
		author := &model.Author{
			ID:   string(rune('a' + i)),
			Name: "Author",
		}
		_ = repo.Create(author)
	}

	authors := repo.List()
	if len(authors) != 3 {
		t.Errorf("Expected 3 authors, got %d", len(authors))
	}
}

func TestAuthorRepository_FindByCountry(t *testing.T) {
	repo := NewAuthorRepository()

	_ = repo.Create(&model.Author{ID: "1", Name: "Author 1", Country: "USA"})
	_ = repo.Create(&model.Author{ID: "2", Name: "Author 2", Country: "USA"})
	_ = repo.Create(&model.Author{ID: "3", Name: "Author 3", Country: "UK"})

	authors := repo.FindByCountry("USA")
	if len(authors) != 2 {
		t.Errorf("Expected 2 authors from USA, got %d", len(authors))
	}
}
