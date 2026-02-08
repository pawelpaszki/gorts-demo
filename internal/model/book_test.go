package model

import (
	"strings"
	"testing"
	"time"
)

func TestBook_Validate(t *testing.T) {
	tests := []struct {
		name    string
		book    Book
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid book",
			book: Book{
				ID:       "book-1",
				Title:    "The Go Programming Language",
				ISBN:     "978-0134190440",
				AuthorID: "author-1",
				Pages:    400,
			},
			wantErr: false,
		},
		{
			name: "missing title",
			book: Book{
				ID:       "book-1",
				ISBN:     "978-0134190440",
				AuthorID: "author-1",
			},
			wantErr: true,
			errMsg:  "title is required",
		},
		{
			name: "title too long",
			book: Book{
				ID:       "book-1",
				Title:    strings.Repeat("a", 201),
				ISBN:     "978-0134190440",
				AuthorID: "author-1",
			},
			wantErr: true,
			errMsg:  "title must be 200 characters or less",
		},
		{
			name: "title at max length",
			book: Book{
				ID:       "book-1",
				Title:    strings.Repeat("a", 200),
				ISBN:     "978-0134190440",
				AuthorID: "author-1",
			},
			wantErr: false,
		},
		{
			name: "missing ISBN",
			book: Book{
				ID:       "book-1",
				Title:    "Test Book",
				AuthorID: "author-1",
			},
			wantErr: true,
			errMsg:  "isbn is required",
		},
		{
			name: "missing author_id",
			book: Book{
				ID:    "book-1",
				Title: "Test Book",
				ISBN:  "978-0134190440",
			},
			wantErr: true,
			errMsg:  "author_id is required",
		},
		{
			name: "negative pages",
			book: Book{
				ID:       "book-1",
				Title:    "Test Book",
				ISBN:     "978-0134190440",
				AuthorID: "author-1",
				Pages:    -1,
			},
			wantErr: true,
			errMsg:  "pages cannot be negative",
		},
		{
			name: "zero pages allowed",
			book: Book{
				ID:       "book-1",
				Title:    "Test Book",
				ISBN:     "978-0134190440",
				AuthorID: "author-1",
				Pages:    0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.book.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Book.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil && tt.errMsg != "" {
				if err.Error() != tt.errMsg {
					t.Errorf("Book.Validate() error = %q, want %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestBook_Fields(t *testing.T) {
	now := time.Now()
	book := Book{
		ID:          "book-123",
		Title:       "Test Book",
		ISBN:        "978-0134190440",
		AuthorID:    "author-456",
		PublishedAt: now,
		Pages:       300,
		Genre:       "Programming",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if book.ID != "book-123" {
		t.Errorf("ID = %s, want book-123", book.ID)
	}
	if book.Title != "Test Book" {
		t.Errorf("Title = %s, want Test Book", book.Title)
	}
	if book.Genre != "Programming" {
		t.Errorf("Genre = %s, want Programming", book.Genre)
	}
	if book.Pages != 300 {
		t.Errorf("Pages = %d, want 300", book.Pages)
	}
}
