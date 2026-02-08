package model

import (
	"strings"
	"testing"
	"time"
)

func TestAuthor_Validate(t *testing.T) {
	tests := []struct {
		name    string
		author  Author
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid author",
			author: Author{
				ID:      "author-1",
				Name:    "Jane Doe",
				Bio:     "A prolific writer",
				Country: "USA",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			author: Author{
				ID:      "author-1",
				Bio:     "A writer",
				Country: "USA",
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "empty name",
			author: Author{
				ID:   "author-1",
				Name: "",
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "name too long",
			author: Author{
				ID:   "author-1",
				Name: strings.Repeat("a", 101),
			},
			wantErr: true,
			errMsg:  "name must be 100 characters or less",
		},
		{
			name: "name at max length",
			author: Author{
				ID:   "author-1",
				Name: strings.Repeat("a", 100),
			},
			wantErr: false,
		},
		{
			name: "bio too long",
			author: Author{
				ID:   "author-1",
				Name: "Jane Doe",
				Bio:  strings.Repeat("a", 2001),
			},
			wantErr: true,
			errMsg:  "bio must be 2000 characters or less",
		},
		{
			name: "bio at max length",
			author: Author{
				ID:   "author-1",
				Name: "Jane Doe",
				Bio:  strings.Repeat("a", 2000),
			},
			wantErr: false,
		},
		{
			name: "empty bio allowed",
			author: Author{
				ID:   "author-1",
				Name: "Jane Doe",
				Bio:  "",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.author.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Author.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil && tt.errMsg != "" {
				if err.Error() != tt.errMsg {
					t.Errorf("Author.Validate() error = %q, want %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestAuthor_Fields(t *testing.T) {
	now := time.Now()
	birthDate := time.Date(1980, 1, 15, 0, 0, 0, 0, time.UTC)

	author := Author{
		ID:        "author-123",
		Name:      "Jane Doe",
		Bio:       "Award-winning author",
		BirthDate: birthDate,
		Country:   "Canada",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if author.ID != "author-123" {
		t.Errorf("ID = %s, want author-123", author.ID)
	}
	if author.Name != "Jane Doe" {
		t.Errorf("Name = %s, want Jane Doe", author.Name)
	}
	if author.Country != "Canada" {
		t.Errorf("Country = %s, want Canada", author.Country)
	}
	if author.BirthDate != birthDate {
		t.Errorf("BirthDate = %v, want %v", author.BirthDate, birthDate)
	}
}
