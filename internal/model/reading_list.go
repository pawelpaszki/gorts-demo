package model

import (
	"errors"
	"time"
)

// ReadingList represents a user's collection of books to read.
type ReadingList struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	BookIDs     []string  `json:"book_ids"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Validate checks if the reading list has valid data.
func (r *ReadingList) Validate() error {
	if r.Name == "" {
		return errors.New("name is required")
	}
	if len(r.Name) > 100 {
		return errors.New("name must be 100 characters or less")
	}
	if len(r.Description) > 500 {
		return errors.New("description must be 500 characters or less")
	}
	return nil
}

// AddBook adds a book ID to the reading list if not already present.
func (r *ReadingList) AddBook(bookID string) bool {
	for _, id := range r.BookIDs {
		if id == bookID {
			return false // Already exists
		}
	}
	r.BookIDs = append(r.BookIDs, bookID)
	return true
}

// RemoveBook removes a book ID from the reading list.
func (r *ReadingList) RemoveBook(bookID string) bool {
	for i, id := range r.BookIDs {
		if id == bookID {
			r.BookIDs = append(r.BookIDs[:i], r.BookIDs[i+1:]...)
			return true
		}
	}
	return false // Not found
}

// ContainsBook checks if a book is in the reading list.
func (r *ReadingList) ContainsBook(bookID string) bool {
	for _, id := range r.BookIDs {
		if id == bookID {
			return true
		}
	}
	return false
}
