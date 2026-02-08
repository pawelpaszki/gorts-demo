package repository

import (
	"errors"
	"sync"
	"time"

	"github.com/pawelpaszki/gorts-demo/internal/model"
)

var (
	ErrReadingListNotFound = errors.New("reading list not found")
	ErrReadingListExists   = errors.New("reading list already exists")
)

// ReadingListRepository provides CRUD operations for reading lists.
type ReadingListRepository struct {
	mu    sync.RWMutex
	lists map[string]*model.ReadingList
}

// NewReadingListRepository creates a new in-memory reading list repository.
func NewReadingListRepository() *ReadingListRepository {
	return &ReadingListRepository{
		lists: make(map[string]*model.ReadingList),
	}
}

// Create adds a new reading list to the repository.
func (r *ReadingListRepository) Create(list *model.ReadingList) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.lists[list.ID]; exists {
		return ErrReadingListExists
	}

	now := time.Now()
	list.CreatedAt = now
	list.UpdatedAt = now

	if list.BookIDs == nil {
		list.BookIDs = []string{}
	}

	stored := *list
	stored.BookIDs = make([]string, len(list.BookIDs))
	copy(stored.BookIDs, list.BookIDs)
	r.lists[list.ID] = &stored
	return nil
}

// Get retrieves a reading list by ID.
func (r *ReadingListRepository) Get(id string) (*model.ReadingList, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list, exists := r.lists[id]
	if !exists {
		return nil, ErrReadingListNotFound
	}

	result := *list
	result.BookIDs = make([]string, len(list.BookIDs))
	copy(result.BookIDs, list.BookIDs)
	return &result, nil
}

// Update modifies an existing reading list.
func (r *ReadingListRepository) Update(list *model.ReadingList) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.lists[list.ID]
	if !exists {
		return ErrReadingListNotFound
	}

	list.CreatedAt = existing.CreatedAt
	list.UpdatedAt = time.Now()

	stored := *list
	stored.BookIDs = make([]string, len(list.BookIDs))
	copy(stored.BookIDs, list.BookIDs)
	r.lists[list.ID] = &stored
	return nil
}

// Delete removes a reading list by ID.
func (r *ReadingListRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.lists[id]; !exists {
		return ErrReadingListNotFound
	}

	delete(r.lists, id)
	return nil
}

// List returns all reading lists.
func (r *ReadingListRepository) List() []*model.ReadingList {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*model.ReadingList, 0, len(r.lists))
	for _, list := range r.lists {
		copy := *list
		copy.BookIDs = make([]string, len(list.BookIDs))
		for i, id := range list.BookIDs {
			copy.BookIDs[i] = id
		}
		result = append(result, &copy)
	}
	return result
}

// FindByBook returns all reading lists containing a specific book.
func (r *ReadingListRepository) FindByBook(bookID string) []*model.ReadingList {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*model.ReadingList
	for _, list := range r.lists {
		if list.ContainsBook(bookID) {
			copy := *list
			copy.BookIDs = make([]string, len(list.BookIDs))
			for i, id := range list.BookIDs {
				copy.BookIDs[i] = id
			}
			result = append(result, &copy)
		}
	}
	return result
}

// Count returns the total number of reading lists.
func (r *ReadingListRepository) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.lists)
}
