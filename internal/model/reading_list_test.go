package model

import (
	"strings"
	"testing"
)

func TestReadingList_Validate(t *testing.T) {
	tests := []struct {
		name    string
		list    ReadingList
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid reading list",
			list: ReadingList{
				ID:          "list-1",
				Name:        "My Favorites",
				Description: "Books I love",
				BookIDs:     []string{"book-1", "book-2"},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			list: ReadingList{
				ID:          "list-1",
				Description: "Some books",
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "empty name",
			list: ReadingList{
				ID:   "list-1",
				Name: "",
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "name too long",
			list: ReadingList{
				ID:   "list-1",
				Name: strings.Repeat("a", 101),
			},
			wantErr: true,
			errMsg:  "name must be 100 characters or less",
		},
		{
			name: "name at max length",
			list: ReadingList{
				ID:   "list-1",
				Name: strings.Repeat("a", 100),
			},
			wantErr: false,
		},
		{
			name: "description too long",
			list: ReadingList{
				ID:          "list-1",
				Name:        "My List",
				Description: strings.Repeat("a", 501),
			},
			wantErr: true,
			errMsg:  "description must be 500 characters or less",
		},
		{
			name: "description at max length",
			list: ReadingList{
				ID:          "list-1",
				Name:        "My List",
				Description: strings.Repeat("a", 500),
			},
			wantErr: false,
		},
		{
			name: "empty book list allowed",
			list: ReadingList{
				ID:      "list-1",
				Name:    "Empty List",
				BookIDs: []string{},
			},
			wantErr: false,
		},
		{
			name: "nil book list allowed",
			list: ReadingList{
				ID:      "list-1",
				Name:    "Nil List",
				BookIDs: nil,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.list.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadingList.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil && tt.errMsg != "" {
				if err.Error() != tt.errMsg {
					t.Errorf("ReadingList.Validate() error = %q, want %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestReadingList_AddBook(t *testing.T) {
	list := &ReadingList{
		ID:      "list-1",
		Name:    "Test List",
		BookIDs: []string{},
	}

	// Add first book
	added := list.AddBook("book-1")
	if !added {
		t.Error("AddBook should return true for new book")
	}
	if len(list.BookIDs) != 1 {
		t.Errorf("Expected 1 book, got %d", len(list.BookIDs))
	}

	// Add second book
	added = list.AddBook("book-2")
	if !added {
		t.Error("AddBook should return true for new book")
	}
	if len(list.BookIDs) != 2 {
		t.Errorf("Expected 2 books, got %d", len(list.BookIDs))
	}

	// Try to add duplicate
	added = list.AddBook("book-1")
	if added {
		t.Error("AddBook should return false for duplicate book")
	}
	if len(list.BookIDs) != 2 {
		t.Errorf("Expected 2 books after duplicate add, got %d", len(list.BookIDs))
	}
}

func TestReadingList_RemoveBook(t *testing.T) {
	list := &ReadingList{
		ID:      "list-1",
		Name:    "Test List",
		BookIDs: []string{"book-1", "book-2", "book-3"},
	}

	// Remove middle book
	removed := list.RemoveBook("book-2")
	if !removed {
		t.Error("RemoveBook should return true for existing book")
	}
	if len(list.BookIDs) != 2 {
		t.Errorf("Expected 2 books, got %d", len(list.BookIDs))
	}
	if list.ContainsBook("book-2") {
		t.Error("book-2 should be removed")
	}

	// Remove first book
	removed = list.RemoveBook("book-1")
	if !removed {
		t.Error("RemoveBook should return true for existing book")
	}
	if len(list.BookIDs) != 1 {
		t.Errorf("Expected 1 book, got %d", len(list.BookIDs))
	}

	// Try to remove non-existent book
	removed = list.RemoveBook("book-999")
	if removed {
		t.Error("RemoveBook should return false for non-existent book")
	}

	// Remove last book
	removed = list.RemoveBook("book-3")
	if !removed {
		t.Error("RemoveBook should return true for existing book")
	}
	if len(list.BookIDs) != 0 {
		t.Errorf("Expected 0 books, got %d", len(list.BookIDs))
	}
}

func TestReadingList_ContainsBook(t *testing.T) {
	list := &ReadingList{
		ID:      "list-1",
		Name:    "Test List",
		BookIDs: []string{"book-1", "book-2"},
	}

	if !list.ContainsBook("book-1") {
		t.Error("ContainsBook should return true for book-1")
	}
	if !list.ContainsBook("book-2") {
		t.Error("ContainsBook should return true for book-2")
	}
	if list.ContainsBook("book-3") {
		t.Error("ContainsBook should return false for book-3")
	}
	if list.ContainsBook("") {
		t.Error("ContainsBook should return false for empty string")
	}
}

func TestReadingList_EmptyList(t *testing.T) {
	list := &ReadingList{
		ID:      "list-1",
		Name:    "Empty List",
		BookIDs: []string{},
	}

	if list.ContainsBook("book-1") {
		t.Error("Empty list should not contain any book")
	}

	removed := list.RemoveBook("book-1")
	if removed {
		t.Error("RemoveBook on empty list should return false")
	}

	added := list.AddBook("book-1")
	if !added {
		t.Error("AddBook on empty list should return true")
	}
}
