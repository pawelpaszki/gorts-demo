package validator

import (
	"testing"
)

func TestISBN10_Checksum(t *testing.T) {
	tests := []struct {
		name    string
		isbn    string
		wantErr bool
	}{
		// Valid ISBN-10s with correct checksums
		{"valid 0306406152", "0306406152", false},
		{"valid 0470059028", "0470059028", false},
		{"valid with X", "080442957X", false},
		{"valid lowercase x", "080442957x", false},

		// Invalid ISBN-10s with wrong checksums
		{"wrong checksum 0306406151", "0306406151", true},
		{"wrong checksum 0306406153", "0306406153", true},
		{"wrong checksum ending", "0470059027", true},

		// Invalid format
		{"too short", "030640615", true},
		{"too long", "03064061522", true},
		{"letters in middle", "030a406152", true},
		{"all letters", "abcdefghij", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ISBN(tt.isbn)
			if (err != nil) != tt.wantErr {
				t.Errorf("ISBN(%q) error = %v, wantErr %v", tt.isbn, err, tt.wantErr)
			}
		})
	}
}

func TestISBN13_Checksum(t *testing.T) {
	tests := []struct {
		name    string
		isbn    string
		wantErr bool
	}{
		// Valid ISBN-13s with correct checksums
		{"valid 9780306406157", "9780306406157", false},
		{"valid 9781234567897", "9781234567897", false},
		{"valid 9780470059029", "9780470059029", false},

		// Invalid ISBN-13s with wrong checksums
		{"wrong checksum 9780306406158", "9780306406158", true},
		{"wrong checksum 9780306406156", "9780306406156", true},
		{"wrong checksum 9781234567890", "9781234567890", true},

		// Invalid format
		{"too short", "978030640615", true},
		{"too long", "97803064061577", true},
		{"doesn't start with 978/979", "1234567890123", true},
		{"letters", "978abc0406157", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ISBN(tt.isbn)
			if (err != nil) != tt.wantErr {
				t.Errorf("ISBN(%q) error = %v, wantErr %v", tt.isbn, err, tt.wantErr)
			}
		})
	}
}

func TestISBN_WithHyphens(t *testing.T) {
	tests := []struct {
		name    string
		isbn    string
		wantErr bool
	}{
		// Valid ISBN-10 with hyphens
		{"ISBN-10 with hyphens", "0-306-40615-2", false},
		{"ISBN-10 different format", "0-470-05902-8", false},

		// Valid ISBN-13 with hyphens
		{"ISBN-13 with hyphens", "978-0-306-40615-7", false},
		{"ISBN-13 different format", "978-1-234-56789-7", false},

		// Invalid with hyphens
		{"wrong checksum with hyphens", "0-306-40615-1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ISBN(tt.isbn)
			if (err != nil) != tt.wantErr {
				t.Errorf("ISBN(%q) error = %v, wantErr %v", tt.isbn, err, tt.wantErr)
			}
		})
	}
}

func TestIsValidISBN10(t *testing.T) {
	tests := []struct {
		isbn string
		want bool
	}{
		{"0306406152", true},
		{"080442957X", true},
		{"0306406151", false},
		{"123456789", false},  // too short
		{"12345678901", false}, // too long
	}

	for _, tt := range tests {
		t.Run(tt.isbn, func(t *testing.T) {
			got := isValidISBN10(tt.isbn)
			if got != tt.want {
				t.Errorf("isValidISBN10(%q) = %v, want %v", tt.isbn, got, tt.want)
			}
		})
	}
}

func TestIsValidISBN13(t *testing.T) {
	tests := []struct {
		isbn string
		want bool
	}{
		{"9780306406157", true},
		{"9781234567897", true},
		{"9780306406158", false},
		{"978030640615", false},   // too short
		{"97803064061577", false}, // too long
	}

	for _, tt := range tests {
		t.Run(tt.isbn, func(t *testing.T) {
			got := isValidISBN13(tt.isbn)
			if got != tt.want {
				t.Errorf("isValidISBN13(%q) = %v, want %v", tt.isbn, got, tt.want)
			}
		})
	}
}
