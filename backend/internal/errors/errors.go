package errors

import (
	"errors"
	"net/http"
)

// AppError represents a standardized application error.
type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
	Err        error  `json:"-"`
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// Unwrap returns the wrapped error.
func (e *AppError) Unwrap() error {
	return e.Err
}

// New creates a new AppError.
func New(code string, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// Wrap wraps an existing error with an AppError.
func Wrap(err error, code string, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Err:        err,
	}
}

// Predefined application errors
var (
	// General errors
	ErrInternalServer = New("INTERNAL_SERVER_ERROR", "An unexpected error occurred", http.StatusInternalServerError)
	ErrNotFound       = New("NOT_FOUND", "Resource not found", http.StatusNotFound)
	ErrBadRequest     = New("BAD_REQUEST", "Invalid request", http.StatusBadRequest)
	ErrConflict       = New("CONFLICT", "Resource already exists", http.StatusConflict)

	// Authentication errors
	ErrUnauthorized = New("UNAUTHORIZED", "Authentication required", http.StatusUnauthorized)
	ErrInvalidToken = New("INVALID_TOKEN", "Invalid or expired token", http.StatusUnauthorized)
	ErrForbidden    = New("FORBIDDEN", "Insufficient permissions", http.StatusForbidden)
	ErrTokenExpired = New("TOKEN_EXPIRED", "Token has expired", http.StatusUnauthorized)

	// Validation errors
	ErrValidation   = New("VALIDATION_ERROR", "Validation failed", http.StatusUnprocessableEntity)
	ErrInvalidInput = New("INVALID_INPUT", "Invalid input provided", http.StatusBadRequest)

	// Rate limiting
	ErrRateLimited = New("RATE_LIMITED", "Too many requests, please try again later", http.StatusTooManyRequests)

	// Database errors
	ErrDatabase       = New("DATABASE_ERROR", "A database error occurred", http.StatusInternalServerError)
	ErrDuplicateEntry = New("DUPLICATE_ENTRY", "Duplicate entry", http.StatusConflict)
)

// IsAppError checks if an error is an AppError.
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// GetAppError extracts an AppError from an error chain.
func GetAppError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return nil
}

// GetStatusCode returns the HTTP status code from an error.
// Returns 500 if the error is not an AppError.
func GetStatusCode(err error) int {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.StatusCode
	}
	return http.StatusInternalServerError
}

// GetCode returns the error code from an error.
// Returns "UNKNOWN_ERROR" if the error is not an AppError.
func GetCode(err error) string {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return "UNKNOWN_ERROR"
}
