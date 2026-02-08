package model

import (
	"errors"
	"time"
)

// Book represents a book in the bookshelf.
type Book struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	ISBN        string    `json:"isbn"`
	AuthorID    string    `json:"author_id"`
	PublishedAt time.Time `json:"published_at"`
	Pages       int       `json:"pages"`
	Genre       string    `json:"genre"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Validate checks if the book has valid data.
func (b *Book) Validate() error {
	if b.Title == "" {
		return errors.New("title is required")
	}
	if len(b.Title) > 200 {
		return errors.New("title must be 200 characters or less")
	}
	if b.ISBN == "" {
		return errors.New("isbn is required")
	}
	if b.AuthorID == "" {
		return errors.New("author_id is required")
	}
	if b.Pages < 0 {
		return errors.New("pages cannot be negative")
	}
	return nil
}

// IsPublished returns true if the book has a publication date in the past.
func (b *Book) IsPublished() bool {
	return !b.PublishedAt.IsZero() && b.PublishedAt.Before(time.Now())
}
