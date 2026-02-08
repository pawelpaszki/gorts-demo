package validator

import (
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"
)

var (
	ErrRequired     = errors.New("field is required")
	ErrTooLong      = errors.New("field exceeds maximum length")
	ErrTooShort     = errors.New("field is below minimum length")
	ErrInvalidISBN  = errors.New("invalid ISBN format")
	ErrInvalidEmail = errors.New("invalid email format")
)

// ISBN patterns for ISBN-10 and ISBN-13
var (
	isbn10Pattern = regexp.MustCompile(`^(\d{9}[\dXx]|\d-\d{3}-\d{5}-[\dXx]|\d{1}-\d{5}-\d{3}-[\dXx])$`)
	isbn13Pattern = regexp.MustCompile(`^(97[89]\d{10}|97[89]-\d-\d{3}-\d{5}-\d|97[89]-\d{1}-\d{5}-\d{3}-\d)$`)
	emailPattern  = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)

// NotEmpty validates that a string is not empty after trimming whitespace.
func NotEmpty(value string) error {
	if strings.TrimSpace(value) == "" {
		return ErrRequired
	}
	return nil
}

// MaxLength validates that a string does not exceed the maximum length.
func MaxLength(value string, max int) error {
	if utf8.RuneCountInString(value) > max {
		return ErrTooLong
	}
	return nil
}

// MinLength validates that a string meets the minimum length.
func MinLength(value string, min int) error {
	if utf8.RuneCountInString(value) < min {
		return ErrTooShort
	}
	return nil
}

// LengthBetween validates that a string length is within the specified range.
func LengthBetween(value string, min, max int) error {
	length := utf8.RuneCountInString(value)
	if length < min {
		return ErrTooShort
	}
	if length > max {
		return ErrTooLong
	}
	return nil
}

// ISBN validates ISBN-10 or ISBN-13 format.
func ISBN(value string) error {
	// Remove hyphens for validation
	cleaned := strings.ReplaceAll(value, "-", "")

	if len(cleaned) == 10 {
		if !isValidISBN10(cleaned) {
			return ErrInvalidISBN
		}
		return nil
	}

	if len(cleaned) == 13 {
		if !isValidISBN13(cleaned) {
			return ErrInvalidISBN
		}
		return nil
	}

	return ErrInvalidISBN
}

// isValidISBN10 checks ISBN-10 checksum.
func isValidISBN10(isbn string) bool {
	if len(isbn) != 10 {
		return false
	}

	sum := 0
	for i := 0; i < 9; i++ {
		if isbn[i] < '0' || isbn[i] > '9' {
			return false
		}
		sum += int(isbn[i]-'0') * (10 - i)
	}

	// Last character can be 'X' representing 10
	last := isbn[9]
	if last == 'X' || last == 'x' {
		sum += 10
	} else if last >= '0' && last <= '9' {
		sum += int(last - '0')
	} else {
		return false
	}

	return sum%11 == 0
}

// isValidISBN13 checks ISBN-13 checksum.
func isValidISBN13(isbn string) bool {
	if len(isbn) != 13 {
		return false
	}

	sum := 0
	for i := 0; i < 13; i++ {
		if isbn[i] < '0' || isbn[i] > '9' {
			return false
		}
		digit := int(isbn[i] - '0')
		if i%2 == 0 {
			sum += digit
		} else {
			sum += digit * 3
		}
	}

	return sum%10 == 0
}

// Email validates email format.
func Email(value string) error {
	if !emailPattern.MatchString(value) {
		return ErrInvalidEmail
	}
	return nil
}

// StringField provides a fluent interface for string validation.
type StringField struct {
	value  string
	errors []error
}

// NewStringField creates a new string field validator.
func NewStringField(value string) *StringField {
	return &StringField{value: value}
}

// Required marks the field as required.
func (f *StringField) Required() *StringField {
	if err := NotEmpty(f.value); err != nil {
		f.errors = append(f.errors, err)
	}
	return f
}

// Max sets maximum length.
func (f *StringField) Max(max int) *StringField {
	if err := MaxLength(f.value, max); err != nil {
		f.errors = append(f.errors, err)
	}
	return f
}

// Min sets minimum length.
func (f *StringField) Min(min int) *StringField {
	if err := MinLength(f.value, min); err != nil {
		f.errors = append(f.errors, err)
	}
	return f
}

// IsISBN validates as ISBN.
func (f *StringField) IsISBN() *StringField {
	if f.value != "" {
		if err := ISBN(f.value); err != nil {
			f.errors = append(f.errors, err)
		}
	}
	return f
}

// IsEmail validates as email.
func (f *StringField) IsEmail() *StringField {
	if f.value != "" {
		if err := Email(f.value); err != nil {
			f.errors = append(f.errors, err)
		}
	}
	return f
}

// Error returns the first validation error, or nil if valid.
func (f *StringField) Error() error {
	if len(f.errors) > 0 {
		return f.errors[0]
	}
	return nil
}

// Errors returns all validation errors.
func (f *StringField) Errors() []error {
	return f.errors
}

// Valid returns true if there are no validation errors.
func (f *StringField) Valid() bool {
	return len(f.errors) == 0
}
