package service

import (
	"testing"

	"github.com/pawelpaszki/gorts-demo/internal/model"
	"github.com/pawelpaszki/gorts-demo/internal/repository"
)

func newTestAuthorService() *AuthorService {
	repo := repository.NewAuthorRepository()
	return NewAuthorService(repo)
}

func validAuthor(id string) *model.Author {
	return &model.Author{
		ID:      id,
		Name:    "Test Author",
		Country: "USA",
		Bio:     "A prolific writer.",
	}
}

func TestAuthorService_CreateAuthor(t *testing.T) {
	svc := newTestAuthorService()
	author := validAuthor("author-1")

	err := svc.CreateAuthor(author)
	if err != nil {
		t.Fatalf("CreateAuthor failed: %v", err)
	}

	if svc.GetAuthorCount() != 1 {
		t.Errorf("Expected 1 author, got %d", svc.GetAuthorCount())
	}
}

func TestAuthorService_CreateAuthor_InvalidData(t *testing.T) {
	svc := newTestAuthorService()
	author := &model.Author{
		ID: "author-1",
		// Missing required Name
	}

	err := svc.CreateAuthor(author)
	if err == nil {
		t.Error("Expected error for invalid author")
	}
}

func TestAuthorService_GetAuthor(t *testing.T) {
	svc := newTestAuthorService()
	original := validAuthor("author-1")
	_ = svc.CreateAuthor(original)

	retrieved, err := svc.GetAuthor("author-1")
	if err != nil {
		t.Fatalf("GetAuthor failed: %v", err)
	}

	if retrieved.Name != original.Name {
		t.Errorf("Expected name %q, got %q", original.Name, retrieved.Name)
	}
}

func TestAuthorService_GetAuthor_NotFound(t *testing.T) {
	svc := newTestAuthorService()

	_, err := svc.GetAuthor("nonexistent")
	if err != ErrAuthorNotFound {
		t.Errorf("Expected ErrAuthorNotFound, got %v", err)
	}
}

func TestAuthorService_UpdateAuthor(t *testing.T) {
	svc := newTestAuthorService()
	author := validAuthor("author-1")
	_ = svc.CreateAuthor(author)

	author.Name = "Updated Name"
	err := svc.UpdateAuthor(author)
	if err != nil {
		t.Fatalf("UpdateAuthor failed: %v", err)
	}

	retrieved, _ := svc.GetAuthor("author-1")
	if retrieved.Name != "Updated Name" {
		t.Errorf("Expected updated name, got %q", retrieved.Name)
	}
}

func TestAuthorService_DeleteAuthor(t *testing.T) {
	svc := newTestAuthorService()
	author := validAuthor("author-1")
	_ = svc.CreateAuthor(author)

	err := svc.DeleteAuthor("author-1")
	if err != nil {
		t.Fatalf("DeleteAuthor failed: %v", err)
	}

	if svc.GetAuthorCount() != 0 {
		t.Error("Author should be deleted")
	}
}

func TestAuthorService_ListAuthors(t *testing.T) {
	svc := newTestAuthorService()

	for i := 0; i < 3; i++ {
		author := validAuthor(string(rune('a' + i)))
		_ = svc.CreateAuthor(author)
	}

	authors := svc.ListAuthors()
	if len(authors) != 3 {
		t.Errorf("Expected 3 authors, got %d", len(authors))
	}
}

func TestAuthorService_GetAuthorsByCountry(t *testing.T) {
	svc := newTestAuthorService()

	author1 := validAuthor("author-1")
	author1.Country = "USA"
	_ = svc.CreateAuthor(author1)

	author2 := validAuthor("author-2")
	author2.Country = "USA"
	_ = svc.CreateAuthor(author2)

	author3 := validAuthor("author-3")
	author3.Country = "UK"
	_ = svc.CreateAuthor(author3)

	authors := svc.GetAuthorsByCountry("USA")
	if len(authors) != 2 {
		t.Errorf("Expected 2 authors, got %d", len(authors))
	}
}
