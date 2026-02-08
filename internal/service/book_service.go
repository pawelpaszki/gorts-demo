package service

import (
	"errors"
	"fmt"

	"github.com/pawelpaszki/gorts-demo/internal/model"
	"github.com/pawelpaszki/gorts-demo/internal/repository"
)

var (
	ErrInvalidBook   = errors.New("invalid book data")
	ErrBookNotFound  = errors.New("book not found")
	ErrDuplicateISBN = errors.New("book with this ISBN already exists")
)

// BookService handles business logic for books.
type BookService struct {
	repo *repository.BookRepository
}

// NewBookService creates a new book service.
func NewBookService(repo *repository.BookRepository) *BookService {
	return &BookService{repo: repo}
}

// CreateBook validates and creates a new book.
func (s *BookService) CreateBook(book *model.Book) error {
	if err := book.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidBook, err)
	}

	// Check for duplicate ISBN
	existingBooks := s.repo.List()
	for _, existing := range existingBooks {
		if existing.ISBN == book.ISBN {
			return ErrDuplicateISBN
		}
	}

	if err := s.repo.Create(book); err != nil {
		return err
	}
	return nil
}

// GetBook retrieves a book by ID.
func (s *BookService) GetBook(id string) (*model.Book, error) {
	book, err := s.repo.Get(id)
	if err != nil {
		if errors.Is(err, repository.ErrBookNotFound) {
			return nil, ErrBookNotFound
		}
		return nil, err
	}
	return book, nil
}

// UpdateBook validates and updates an existing book.
func (s *BookService) UpdateBook(book *model.Book) error {
	if err := book.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidBook, err)
	}

	// Check ISBN uniqueness (excluding current book)
	existingBooks := s.repo.List()
	for _, existing := range existingBooks {
		if existing.ISBN == book.ISBN && existing.ID != book.ID {
			return ErrDuplicateISBN
		}
	}

	if err := s.repo.Update(book); err != nil {
		if errors.Is(err, repository.ErrBookNotFound) {
			return ErrBookNotFound
		}
		return err
	}
	return nil
}

// DeleteBook removes a book by ID.
func (s *BookService) DeleteBook(id string) error {
	if err := s.repo.Delete(id); err != nil {
		if errors.Is(err, repository.ErrBookNotFound) {
			return ErrBookNotFound
		}
		return err
	}
	return nil
}

// ListBooks returns all books.
func (s *BookService) ListBooks() []*model.Book {
	return s.repo.List()
}

// GetBooksByAuthor returns all books by a specific author.
func (s *BookService) GetBooksByAuthor(authorID string) []*model.Book {
	return s.repo.FindByAuthor(authorID)
}

// GetBookCount returns the total number of books.
func (s *BookService) GetBookCount() int {
	return s.repo.Count()
}
