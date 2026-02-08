package repository

import (
	"errors"
	"sync"
	"time"

	"github.com/pawelpaszki/gorts-demo/internal/model"
)

var (
	ErrBookNotFound = errors.New("book not found")
	ErrBookExists   = errors.New("book already exists")
)

// BookRepository provides CRUD operations for books.
type BookRepository struct {
	mu    sync.RWMutex
	books map[string]*model.Book
}

// NewBookRepository creates a new in-memory book repository.
func NewBookRepository() *BookRepository {
	return &BookRepository{
		books: make(map[string]*model.Book),
	}
}

// Create adds a new book to the repository.
func (r *BookRepository) Create(book *model.Book) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.books[book.ID]; exists {
		return ErrBookExists
	}

	now := time.Now()
	book.CreatedAt = now
	book.UpdatedAt = now

	// Store a copy to prevent external mutations
	stored := *book
	r.books[book.ID] = &stored
	return nil
}

// Get retrieves a book by ID.
func (r *BookRepository) Get(id string) (*model.Book, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	book, exists := r.books[id]
	if !exists {
		return nil, ErrBookNotFound
	}

	// Return a copy to prevent external mutations
	result := *book
	return &result, nil
}

// Update modifies an existing book.
func (r *BookRepository) Update(book *model.Book) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.books[book.ID]
	if !exists {
		return ErrBookNotFound
	}

	book.CreatedAt = existing.CreatedAt
	book.UpdatedAt = time.Now()

	stored := *book
	r.books[book.ID] = &stored
	return nil
}

// Delete removes a book by ID.
func (r *BookRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.books[id]; !exists {
		return ErrBookNotFound
	}

	delete(r.books, id)
	return nil
}

// List returns all books.
func (r *BookRepository) List() []*model.Book {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*model.Book, 0, len(r.books))
	for _, book := range r.books {
		copy := *book
		result = append(result, &copy)
	}
	return result
}

// FindByAuthor returns all books by a specific author.
func (r *BookRepository) FindByAuthor(authorID string) []*model.Book {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*model.Book
	for _, book := range r.books {
		if book.AuthorID == authorID {
			copy := *book
			result = append(result, &copy)
		}
	}
	return result
}

// Count returns the total number of books.
func (r *BookRepository) Count() int {
	r.mu.RLock()
	count := len(r.books)
	r.mu.RUnlock()
	return count
}
