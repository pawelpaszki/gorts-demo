package service

import (
	"errors"
	"fmt"

	"github.com/pawelpaszki/gorts-demo/internal/model"
	"github.com/pawelpaszki/gorts-demo/internal/repository"
)

var (
	ErrInvalidReadingList  = errors.New("invalid reading list data")
	ErrReadingListNotFound = errors.New("reading list not found")
	ErrBookAlreadyInList   = errors.New("book already in reading list")
	ErrBookNotInList       = errors.New("book not in reading list")
)

// ReadingListService handles business logic for reading lists.
type ReadingListService struct {
	repo     *repository.ReadingListRepository
	bookRepo *repository.BookRepository
}

// NewReadingListService creates a new reading list service.
func NewReadingListService(repo *repository.ReadingListRepository, bookRepo *repository.BookRepository) *ReadingListService {
	return &ReadingListService{
		repo:     repo,
		bookRepo: bookRepo,
	}
}

// CreateReadingList validates and creates a new reading list.
func (s *ReadingListService) CreateReadingList(list *model.ReadingList) error {
	if err := list.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidReadingList, err)
	}

	if err := s.repo.Create(list); err != nil {
		return err
	}
	return nil
}

// GetReadingList retrieves a reading list by ID.
func (s *ReadingListService) GetReadingList(id string) (*model.ReadingList, error) {
	list, err := s.repo.Get(id)
	if err != nil {
		if errors.Is(err, repository.ErrReadingListNotFound) {
			return nil, ErrReadingListNotFound
		}
		return nil, err
	}
	return list, nil
}

// UpdateReadingList validates and updates an existing reading list.
func (s *ReadingListService) UpdateReadingList(list *model.ReadingList) error {
	if err := list.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidReadingList, err)
	}

	if err := s.repo.Update(list); err != nil {
		if errors.Is(err, repository.ErrReadingListNotFound) {
			return ErrReadingListNotFound
		}
		return err
	}
	return nil
}

// DeleteReadingList removes a reading list by ID.
func (s *ReadingListService) DeleteReadingList(id string) error {
	if err := s.repo.Delete(id); err != nil {
		if errors.Is(err, repository.ErrReadingListNotFound) {
			return ErrReadingListNotFound
		}
		return err
	}
	return nil
}

// ListReadingLists returns all reading lists.
func (s *ReadingListService) ListReadingLists() []*model.ReadingList {
	return s.repo.List()
}

// AddBookToList adds a book to a reading list.
func (s *ReadingListService) AddBookToList(listID, bookID string) error {
	// Verify book exists
	if _, err := s.bookRepo.Get(bookID); err != nil {
		if errors.Is(err, repository.ErrBookNotFound) {
			return ErrBookNotFound
		}
		return err
	}

	list, err := s.repo.Get(listID)
	if err != nil {
		if errors.Is(err, repository.ErrReadingListNotFound) {
			return ErrReadingListNotFound
		}
		return err
	}

	if !list.AddBook(bookID) {
		return ErrBookAlreadyInList
	}

	return s.repo.Update(list)
}

// RemoveBookFromList removes a book from a reading list.
func (s *ReadingListService) RemoveBookFromList(listID, bookID string) error {
	list, err := s.repo.Get(listID)
	if err != nil {
		if errors.Is(err, repository.ErrReadingListNotFound) {
			return ErrReadingListNotFound
		}
		return err
	}

	if !list.RemoveBook(bookID) {
		return ErrBookNotInList
	}

	return s.repo.Update(list)
}

// GetListsContainingBook returns all lists that contain a specific book.
func (s *ReadingListService) GetListsContainingBook(bookID string) []*model.ReadingList {
	return s.repo.FindByBook(bookID)
}

// GetReadingListCount returns the total number of reading lists.
func (s *ReadingListService) GetReadingListCount() int {
	return s.repo.Count()
}
