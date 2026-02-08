// Package stringutil provides string manipulation utilities.
package stringutil

import (
	"strings"
	"unicode"
)

// Truncate shortens a string to the specified length, adding ellipsis if needed.
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// Slugify converts a string to a URL-friendly slug.
// It lowercases the input, replaces spaces and underscores with dashes,
// removes special characters, and collapses consecutive dashes.
func Slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.TrimSpace(s)

	var result strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			result.WriteRune(r)
		} else if r == ' ' || r == '-' || r == '_' {
			result.WriteRune('-')
		}
	}

	// Remove consecutive dashes
	slug := result.String()
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}

	return strings.Trim(slug, "-")
}

// Capitalize capitalizes the first letter of a string.
func Capitalize(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// ToLowerTrimmed returns a lowercase, trimmed version of the string.
func ToLowerTrimmed(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
