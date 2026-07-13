package common

import (
	"math/rand"
	"strings"
	"time"
)

// Constants for common use
const (
	// DefaultPageSize is the default number of items per page
	DefaultPageSize = 20
	// MaxPageSize is the maximum number of items per page
	MaxPageSize = 100
	// MinPageSize is the minimum number of items per page
	MinPageSize = 1
)

// PaginationParams holds pagination request parameters.
type PaginationParams struct {
	Page    int `form:"page" json:"page" binding:"omitempty,min=1"`
	PerPage int `form:"per_page" json:"per_page" binding:"omitempty,min=1,max=100"`
}

// Normalize normalizes pagination parameters with defaults.
func (p *PaginationParams) Normalize() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PerPage < MinPageSize || p.PerPage > MaxPageSize {
		p.PerPage = DefaultPageSize
	}
}

// Offset returns the offset for SQL queries.
func (p *PaginationParams) Offset() int {
	return (p.Page - 1) * p.PerPage
}

// GenerateReference generates a unique reference string.
func GenerateReference(prefix string) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, 16)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return prefix + "_" + string(b)
}

// SanitizeEmail normalizes an email address to lowercase.
func SanitizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// TruncateString truncates a string to the given length.
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// Contains checks if a string slice contains a value.
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// RemoveDuplicates removes duplicate strings from a slice.
func RemoveDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0)
	for _, s := range slice {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}
