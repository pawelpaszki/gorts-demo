package service

import (
	"errors"
	"fmt"

	"github.com/pawelpaszki/gorts-demo/internal/model"
	"github.com/pawelpaszki/gorts-demo/internal/repository"
)

var (
	ErrInvalidAuthor  = errors.New("invalid author data")
	ErrAuthorNotFound = errors.New("author not found")
)

// AuthorService handles business logic for authors.
type AuthorService struct {
	repo *repository.AuthorRepository
}

// NewAuthorService creates a new author service.
func NewAuthorService(repo *repository.AuthorRepository) *AuthorService {
	return &AuthorService{repo: repo}
}

// CreateAuthor validates and creates a new author.
// Returns ErrInvalidAuthor if validation fails.
func (s *AuthorService) CreateAuthor(author *model.Author) error {
	if err := author.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidAuthor, err)
	}

	return s.repo.Create(author)
}

// GetAuthor retrieves an author by ID.
func (s *AuthorService) GetAuthor(id string) (*model.Author, error) {
	author, err := s.repo.Get(id)
	if err != nil {
		if errors.Is(err, repository.ErrAuthorNotFound) {
			return nil, ErrAuthorNotFound
		}
		return nil, err
	}
	return author, nil
}

// UpdateAuthor validates and updates an existing author.
func (s *AuthorService) UpdateAuthor(author *model.Author) error {
	if err := author.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidAuthor, err)
	}

	if err := s.repo.Update(author); err != nil {
		if errors.Is(err, repository.ErrAuthorNotFound) {
			return ErrAuthorNotFound
		}
		return err
	}
	return nil
}

// DeleteAuthor removes an author by ID.
func (s *AuthorService) DeleteAuthor(id string) error {
	if err := s.repo.Delete(id); err != nil {
		if errors.Is(err, repository.ErrAuthorNotFound) {
			return ErrAuthorNotFound
		}
		return err
	}
	return nil
}

// ListAuthors returns all authors.
func (s *AuthorService) ListAuthors() []*model.Author {
	return s.repo.List()
}

// GetAuthorsByCountry returns all authors from a specific country.
func (s *AuthorService) GetAuthorsByCountry(country string) []*model.Author {
	return s.repo.FindByCountry(country)
}

// GetAuthorCount returns the total number of authors.
func (s *AuthorService) GetAuthorCount() int {
	return s.repo.Count()
}
