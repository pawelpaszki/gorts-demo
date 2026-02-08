package validator

import (
	"testing"
)

func TestNotEmpty(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid string", "hello", false},
		{"empty string", "", true},
		{"whitespace only", "   ", true},
		{"string with spaces", "  hello  ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NotEmpty(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("NotEmpty(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestMaxLength(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		max     int
		wantErr bool
	}{
		{"under max", "hello", 10, false},
		{"at max", "hello", 5, false},
		{"over max", "hello world", 5, true},
		{"unicode", "héllo", 5, false},
		{"unicode over", "héllo wörld", 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := MaxLength(tt.value, tt.max)
			if (err != nil) != tt.wantErr {
				t.Errorf("MaxLength(%q, %d) error = %v, wantErr %v", tt.value, tt.max, err, tt.wantErr)
			}
		})
	}
}

func TestMinLength(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		min     int
		wantErr bool
	}{
		{"over min", "hello", 3, false},
		{"at min", "hello", 5, false},
		{"under min", "hi", 5, true},
		{"empty", "", 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := MinLength(tt.value, tt.min)
			if (err != nil) != tt.wantErr {
				t.Errorf("MinLength(%q, %d) error = %v, wantErr %v", tt.value, tt.min, err, tt.wantErr)
			}
		})
	}
}

func TestLengthBetween(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		min     int
		max     int
		wantErr bool
	}{
		{"in range", "hello", 3, 10, false},
		{"at min", "hello", 5, 10, false},
		{"at max", "hello", 1, 5, false},
		{"under min", "hi", 5, 10, true},
		{"over max", "hello world", 1, 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := LengthBetween(tt.value, tt.min, tt.max)
			if (err != nil) != tt.wantErr {
				t.Errorf("LengthBetween(%q, %d, %d) error = %v, wantErr %v", tt.value, tt.min, tt.max, err, tt.wantErr)
			}
		})
	}
}

func TestISBN(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		// Valid ISBN-10
		{"valid ISBN-10", "0306406152", false},
		{"valid ISBN-10 with X", "080442957X", false},
		// Valid ISBN-13
		{"valid ISBN-13", "9780306406157", false},
		{"valid ISBN-13 978", "9781234567897", false},
		// Invalid
		{"too short", "12345", true},
		{"too long", "12345678901234", true},
		{"invalid checksum ISBN-10", "0306406151", true},
		{"invalid chars", "abcdefghij", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ISBN(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ISBN(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestEmail(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid email", "test@example.com", false},
		{"valid with subdomain", "user@mail.example.com", false},
		{"valid with plus", "user+tag@example.com", false},
		{"missing @", "testexample.com", true},
		{"missing domain", "test@", true},
		{"missing local", "@example.com", true},
		{"missing tld", "test@example", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Email(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Email(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestStringField_Fluent(t *testing.T) {
	// Valid case
	f := NewStringField("hello@example.com").Required().Max(50).IsEmail()
	if !f.Valid() {
		t.Errorf("Expected valid, got errors: %v", f.Errors())
	}

	// Invalid - required
	f = NewStringField("").Required()
	if f.Valid() {
		t.Error("Expected invalid for empty required field")
	}

	// Invalid - too long
	f = NewStringField("this is a very long string").Max(10)
	if f.Valid() {
		t.Error("Expected invalid for too long string")
	}

	// Multiple errors
	f = NewStringField("").Required().Min(5)
	if len(f.Errors()) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(f.Errors()))
	}
}

func TestStringField_ISBN(t *testing.T) {
	f := NewStringField("9780306406157").Required().IsISBN()
	if !f.Valid() {
		t.Errorf("Expected valid ISBN, got errors: %v", f.Errors())
	}

	f = NewStringField("invalid-isbn").IsISBN()
	if f.Valid() {
		t.Error("Expected invalid for bad ISBN")
	}
}
