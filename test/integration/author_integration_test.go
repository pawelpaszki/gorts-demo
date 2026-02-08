package integration

import (
	"testing"
	"time"

	"github.com/pawelpaszki/gorts-demo/internal/model"
	"github.com/pawelpaszki/gorts-demo/internal/repository"
	"github.com/pawelpaszki/gorts-demo/internal/service"
)

// TestAuthorServiceIntegration tests the author service with a real repository.
func TestAuthorServiceIntegration(t *testing.T) {
	repo := repository.NewAuthorRepository()
	svc := service.NewAuthorService(repo)

	t.Run("full CRUD lifecycle", func(t *testing.T) {
		// Create
		author := &model.Author{
			ID:        "integration-author-1",
			Name:      "Jane Doe",
			Bio:       "Award-winning author of multiple bestsellers",
			Country:   "USA",
			BirthDate: time.Date(1975, 6, 15, 0, 0, 0, 0, time.UTC),
		}

		err := svc.CreateAuthor(author)
		if err != nil {
			t.Fatalf("CreateAuthor failed: %v", err)
		}

		// Read
		retrieved, err := svc.GetAuthor("integration-author-1")
		if err != nil {
			t.Fatalf("GetAuthor failed: %v", err)
		}
		if retrieved.Name != author.Name {
			t.Errorf("Name mismatch: got %q, want %q", retrieved.Name, author.Name)
		}
		if retrieved.Country != "USA" {
			t.Errorf("Country mismatch: got %q, want %q", retrieved.Country, "USA")
		}
		if retrieved.CreatedAt.IsZero() {
			t.Error("CreatedAt should be set")
		}

		// Update
		retrieved.Name = "Jane Smith"
		retrieved.Bio = "Updated bio information"
		retrieved.Country = "Canada"
		err = svc.UpdateAuthor(retrieved)
		if err != nil {
			t.Fatalf("UpdateAuthor failed: %v", err)
		}

		updated, _ := svc.GetAuthor("integration-author-1")
		if updated.Name != "Jane Smith" {
			t.Errorf("Name not updated: got %q", updated.Name)
		}
		if updated.Country != "Canada" {
			t.Errorf("Country not updated: got %q", updated.Country)
		}
		if updated.CreatedAt != retrieved.CreatedAt {
			t.Error("CreatedAt should not change on update")
		}

		// Delete
		err = svc.DeleteAuthor("integration-author-1")
		if err != nil {
			t.Fatalf("DeleteAuthor failed: %v", err)
		}

		_, err = svc.GetAuthor("integration-author-1")
		if err != service.ErrAuthorNotFound {
			t.Errorf("Expected ErrAuthorNotFound after delete, got %v", err)
		}
	})
}

func TestAuthorServiceIntegration_MultipleAuthors(t *testing.T) {
	repo := repository.NewAuthorRepository()
	svc := service.NewAuthorService(repo)

	// Create multiple authors from different countries
	authors := []*model.Author{
		{ID: "author-1", Name: "Author One", Country: "USA", Bio: "American writer"},
		{ID: "author-2", Name: "Author Two", Country: "USA", Bio: "Another American"},
		{ID: "author-3", Name: "Author Three", Country: "UK", Bio: "British author"},
		{ID: "author-4", Name: "Author Four", Country: "UK", Bio: "Another Brit"},
		{ID: "author-5", Name: "Author Five", Country: "Canada", Bio: "Canadian author"},
	}

	for _, author := range authors {
		if err := svc.CreateAuthor(author); err != nil {
			t.Fatalf("Failed to create %s: %v", author.ID, err)
		}
	}

	// Verify count
	if count := svc.GetAuthorCount(); count != 5 {
		t.Errorf("Expected 5 authors, got %d", count)
	}

	// List all
	allAuthors := svc.ListAuthors()
	if len(allAuthors) != 5 {
		t.Errorf("Expected 5 authors in list, got %d", len(allAuthors))
	}

	// Filter by country
	usaAuthors := svc.GetAuthorsByCountry("USA")
	if len(usaAuthors) != 2 {
		t.Errorf("Expected 2 authors from USA, got %d", len(usaAuthors))
	}

	ukAuthors := svc.GetAuthorsByCountry("UK")
	if len(ukAuthors) != 2 {
		t.Errorf("Expected 2 authors from UK, got %d", len(ukAuthors))
	}

	canadaAuthors := svc.GetAuthorsByCountry("Canada")
	if len(canadaAuthors) != 1 {
		t.Errorf("Expected 1 author from Canada, got %d", len(canadaAuthors))
	}

	// Non-existent country
	germanyAuthors := svc.GetAuthorsByCountry("Germany")
	if len(germanyAuthors) != 0 {
		t.Errorf("Expected 0 authors from Germany, got %d", len(germanyAuthors))
	}

	// Delete some authors
	_ = svc.DeleteAuthor("author-2")
	_ = svc.DeleteAuthor("author-4")

	if count := svc.GetAuthorCount(); count != 3 {
		t.Errorf("Expected 3 authors after delete, got %d", count)
	}

	// Verify country counts after deletion
	usaAuthors = svc.GetAuthorsByCountry("USA")
	if len(usaAuthors) != 1 {
		t.Errorf("Expected 1 author from USA after delete, got %d", len(usaAuthors))
	}
}

func TestAuthorServiceIntegration_ValidationErrors(t *testing.T) {
	repo := repository.NewAuthorRepository()
	svc := service.NewAuthorService(repo)

	tests := []struct {
		name   string
		author *model.Author
	}{
		{
			name:   "missing name",
			author: &model.Author{ID: "a1", Bio: "Some bio", Country: "USA"},
		},
		{
			name:   "empty name",
			author: &model.Author{ID: "a2", Name: "", Bio: "Some bio"},
		},
		{
			name:   "name too long",
			author: &model.Author{ID: "a3", Name: string(make([]byte, 101))},
		},
		{
			name:   "bio too long",
			author: &model.Author{ID: "a4", Name: "Valid Name", Bio: string(make([]byte, 2001))},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.CreateAuthor(tt.author)
			if err == nil {
				t.Error("Expected validation error")
			}
		})
	}
}

func TestAuthorServiceIntegration_UpdateNonExistent(t *testing.T) {
	repo := repository.NewAuthorRepository()
	svc := service.NewAuthorService(repo)

	author := &model.Author{
		ID:   "non-existent",
		Name: "Ghost Author",
	}

	err := svc.UpdateAuthor(author)
	if err != service.ErrAuthorNotFound {
		t.Errorf("Expected ErrAuthorNotFound, got %v", err)
	}
}

func TestAuthorServiceIntegration_DeleteNonExistent(t *testing.T) {
	repo := repository.NewAuthorRepository()
	svc := service.NewAuthorService(repo)

	err := svc.DeleteAuthor("non-existent")
	if err != service.ErrAuthorNotFound {
		t.Errorf("Expected ErrAuthorNotFound, got %v", err)
	}
}

func TestAuthorServiceIntegration_ConcurrentAccess(t *testing.T) {
	repo := repository.NewAuthorRepository()
	svc := service.NewAuthorService(repo)

	// Create initial author
	author := &model.Author{
		ID:      "concurrent-author",
		Name:    "Concurrent Access Test",
		Country: "Test Country",
	}
	_ = svc.CreateAuthor(author)

	// Concurrent reads
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_, _ = svc.GetAuthor("concurrent-author")
				_ = svc.ListAuthors()
				_ = svc.GetAuthorsByCountry("Test Country")
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify data integrity
	retrieved, err := svc.GetAuthor("concurrent-author")
	if err != nil {
		t.Fatalf("Author should still exist: %v", err)
	}
	if retrieved.Name != "Concurrent Access Test" {
		t.Error("Author data corrupted")
	}
}

func TestAuthorServiceIntegration_TimestampBehavior(t *testing.T) {
	repo := repository.NewAuthorRepository()
	svc := service.NewAuthorService(repo)

	// Create author
	author := &model.Author{
		ID:   "timestamp-test",
		Name: "Timestamp Test",
	}
	_ = svc.CreateAuthor(author)

	// Get and check timestamps
	created, _ := svc.GetAuthor("timestamp-test")
	if created.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set on create")
	}
	if created.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set on create")
	}
	if created.CreatedAt != created.UpdatedAt {
		t.Error("CreatedAt and UpdatedAt should be equal on create")
	}

	// Wait a bit and update
	time.Sleep(10 * time.Millisecond)
	created.Name = "Updated Name"
	_ = svc.UpdateAuthor(created)

	// Check timestamps after update
	updated, _ := svc.GetAuthor("timestamp-test")
	if updated.CreatedAt != created.CreatedAt {
		t.Error("CreatedAt should not change on update")
	}
	if !updated.UpdatedAt.After(created.CreatedAt) {
		t.Error("UpdatedAt should be after CreatedAt after update")
	}
}
