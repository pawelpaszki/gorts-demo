package model

import (
	"errors"
	"time"
)

// Author represents a book author.
type Author struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Bio       string    `json:"bio"`
	BirthDate time.Time `json:"birth_date,omitempty"`
	Country   string    `json:"country"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Validate checks if the author has valid data.
func (a *Author) Validate() error {
	if a.Name == "" {
		return errors.New("name is required")
	}
	if len(a.Name) > 100 {
		return errors.New("name must be 100 characters or less")
	}
	if len(a.Bio) > 2000 {
		return errors.New("bio must be 2000 characters or less")
	}
	return nil
}

// HasBio returns true if the author has a biography.
func (a *Author) HasBio() bool {
	return a.Bio != ""
}
