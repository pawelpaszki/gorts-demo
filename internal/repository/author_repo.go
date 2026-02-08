package repository

import (
	"errors"
	"sync"
	"time"

	"github.com/pawelpaszki/gorts-demo/internal/model"
)

var (
	ErrAuthorNotFound = errors.New("author not found")
	ErrAuthorExists   = errors.New("author already exists")
)

// AuthorRepository provides CRUD operations for authors.
type AuthorRepository struct {
	mu      sync.RWMutex
	authors map[string]*model.Author
}

// NewAuthorRepository creates a new in-memory author repository.
func NewAuthorRepository() *AuthorRepository {
	return &AuthorRepository{
		authors: make(map[string]*model.Author),
	}
}

// Create adds a new author to the repository.
func (r *AuthorRepository) Create(author *model.Author) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.authors[author.ID]; exists {
		return ErrAuthorExists
	}

	now := time.Now()
	author.CreatedAt = now
	author.UpdatedAt = now

	stored := *author
	r.authors[author.ID] = &stored
	return nil
}

// Get retrieves an author by ID.
func (r *AuthorRepository) Get(id string) (*model.Author, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	author, exists := r.authors[id]
	if !exists {
		return nil, ErrAuthorNotFound
	}

	result := *author
	return &result, nil
}

// Update modifies an existing author.
func (r *AuthorRepository) Update(author *model.Author) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.authors[author.ID]
	if !exists {
		return ErrAuthorNotFound
	}

	author.CreatedAt = existing.CreatedAt
	author.UpdatedAt = time.Now()

	stored := *author
	r.authors[author.ID] = &stored
	return nil
}

// Delete removes an author by ID.
func (r *AuthorRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.authors[id]; !exists {
		return ErrAuthorNotFound
	}

	delete(r.authors, id)
	return nil
}

// List returns all authors.
func (r *AuthorRepository) List() []*model.Author {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*model.Author, 0, len(r.authors))
	for _, author := range r.authors {
		copy := *author
		result = append(result, &copy)
	}
	return result
}

// FindByCountry returns all authors from a specific country.
func (r *AuthorRepository) FindByCountry(country string) []*model.Author {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*model.Author
	for _, author := range r.authors {
		if author.Country == country {
			copy := *author
			result = append(result, &copy)
		}
	}
	return result
}

// Count returns the total number of authors.
func (r *AuthorRepository) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.authors)
}
