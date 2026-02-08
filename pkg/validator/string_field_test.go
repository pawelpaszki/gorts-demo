package validator

import (
	"testing"
)

func TestStringField_Required(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid", "hello", false},
		{"empty", "", true},
		{"whitespace", "   ", true},
		{"with spaces", " hello ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewStringField(tt.value).Required()
			if f.Valid() == tt.wantErr {
				t.Errorf("StringField(%q).Required().Valid() = %v, want %v", tt.value, f.Valid(), !tt.wantErr)
			}
		})
	}
}

func TestStringField_Max(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		max     int
		wantErr bool
	}{
		{"under max", "hello", 10, false},
		{"at max", "hello", 5, false},
		{"over max", "hello world", 5, true},
		{"empty", "", 5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewStringField(tt.value).Max(tt.max)
			if f.Valid() == tt.wantErr {
				t.Errorf("StringField(%q).Max(%d).Valid() = %v, want %v", tt.value, tt.max, f.Valid(), !tt.wantErr)
			}
		})
	}
}

func TestStringField_Min(t *testing.T) {
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
			f := NewStringField(tt.value).Min(tt.min)
			if f.Valid() == tt.wantErr {
				t.Errorf("StringField(%q).Min(%d).Valid() = %v, want %v", tt.value, tt.min, f.Valid(), !tt.wantErr)
			}
		})
	}
}

func TestStringField_Chaining(t *testing.T) {
	// Valid: all conditions met
	f := NewStringField("test@example.com").Required().Min(5).Max(50).IsEmail()
	if !f.Valid() {
		t.Errorf("Expected valid, got errors: %v", f.Errors())
	}

	// Invalid: fails required
	f = NewStringField("").Required().Min(5).Max(50)
	if f.Valid() {
		t.Error("Expected invalid for empty required field")
	}
	if len(f.Errors()) != 2 { // required + min
		t.Errorf("Expected 2 errors, got %d", len(f.Errors()))
	}

	// Invalid: too short
	f = NewStringField("ab").Required().Min(5)
	if f.Valid() {
		t.Error("Expected invalid for too short field")
	}

	// Invalid: too long
	f = NewStringField("this is a very long string that exceeds the max").Max(20)
	if f.Valid() {
		t.Error("Expected invalid for too long field")
	}
}

func TestStringField_IsEmail(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid email", "user@example.com", false},
		{"valid with subdomain", "user@mail.example.com", false},
		{"invalid no @", "userexample.com", true},
		{"invalid no domain", "user@", true},
		{"empty skipped", "", false}, // Empty is not validated for email
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewStringField(tt.value).IsEmail()
			if f.Valid() == tt.wantErr {
				t.Errorf("StringField(%q).IsEmail().Valid() = %v, want %v", tt.value, f.Valid(), !tt.wantErr)
			}
		})
	}
}

func TestStringField_IsISBN(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid ISBN-10", "0306406152", false},
		{"valid ISBN-13", "9780306406157", false},
		{"invalid ISBN", "1234567890", true},
		{"empty skipped", "", false}, // Empty is not validated for ISBN
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewStringField(tt.value).IsISBN()
			if f.Valid() == tt.wantErr {
				t.Errorf("StringField(%q).IsISBN().Valid() = %v, want %v", tt.value, f.Valid(), !tt.wantErr)
			}
		})
	}
}

func TestStringField_Error(t *testing.T) {
	// Single error
	f := NewStringField("").Required()
	err := f.Error()
	if err == nil {
		t.Error("Expected error for empty required field")
	}
	if err != ErrRequired {
		t.Errorf("Expected ErrRequired, got %v", err)
	}

	// No error
	f = NewStringField("hello").Required()
	err = f.Error()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestStringField_Errors(t *testing.T) {
	// Multiple errors
	f := NewStringField("").Required().Min(5)
	errs := f.Errors()
	if len(errs) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(errs))
	}

	// No errors
	f = NewStringField("hello world").Required().Min(5)
	errs = f.Errors()
	if len(errs) != 0 {
		t.Errorf("Expected 0 errors, got %d", len(errs))
	}
}
