package errors_test

import (
	"net/http"
	"testing"

	apperrors "github.com/coindistro/backend/internal/errors"
)

func TestNew(t *testing.T) {
	err := apperrors.New("TEST_ERROR", "test message", http.StatusBadRequest)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Code != "TEST_ERROR" {
		t.Errorf("expected code TEST_ERROR, got %s", err.Code)
	}
	if err.Message != "test message" {
		t.Errorf("expected message 'test message', got %s", err.Message)
	}
	if err.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status code %d, got %d", http.StatusBadRequest, err.StatusCode)
	}
}

func TestWrap(t *testing.T) {
	original := apperrors.New("ORIGINAL", "original error", http.StatusInternalServerError)
	wrapped := apperrors.Wrap(original, "WRAPPED", "wrapped error", http.StatusBadGateway)

	if wrapped.Err != original {
		t.Error("expected wrapped error to contain original error")
	}
	if wrapped.Code != "WRAPPED" {
		t.Errorf("expected code WRAPPED, got %s", wrapped.Code)
	}
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        *apperrors.AppError
		statusCode int
		code       string
	}{
		{"ErrInternalServer", apperrors.ErrInternalServer, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR"},
		{"ErrNotFound", apperrors.ErrNotFound, http.StatusNotFound, "NOT_FOUND"},
		{"ErrBadRequest", apperrors.ErrBadRequest, http.StatusBadRequest, "BAD_REQUEST"},
		{"ErrUnauthorized", apperrors.ErrUnauthorized, http.StatusUnauthorized, "UNAUTHORIZED"},
		{"ErrForbidden", apperrors.ErrForbidden, http.StatusForbidden, "FORBIDDEN"},
		{"ErrValidation", apperrors.ErrValidation, http.StatusUnprocessableEntity, "VALIDATION_ERROR"},
		{"ErrRateLimited", apperrors.ErrRateLimited, http.StatusTooManyRequests, "RATE_LIMITED"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.StatusCode != tt.statusCode {
				t.Errorf("expected status code %d, got %d", tt.statusCode, tt.err.StatusCode)
			}
			if tt.err.Code != tt.code {
				t.Errorf("expected code %s, got %s", tt.code, tt.err.Code)
			}
		})
	}
}

func TestIsAppError(t *testing.T) {
	err := apperrors.ErrNotFound
	if !apperrors.IsAppError(err) {
		t.Error("expected IsAppError to return true for AppError")
	}

	if apperrors.IsAppError(nil) {
		t.Error("expected IsAppError to return false for nil")
	}
}

func TestGetStatusCode(t *testing.T) {
	statusCode := apperrors.GetStatusCode(apperrors.ErrNotFound)
	if statusCode != http.StatusNotFound {
		t.Errorf("expected status code %d, got %d", http.StatusNotFound, statusCode)
	}

	statusCode = apperrors.GetStatusCode(nil)
	if statusCode != http.StatusInternalServerError {
		t.Errorf("expected status code %d for nil error, got %d", http.StatusInternalServerError, statusCode)
	}
}

func TestGetCode(t *testing.T) {
	code := apperrors.GetCode(apperrors.ErrUnauthorized)
	if code != "UNAUTHORIZED" {
		t.Errorf("expected code UNAUTHORIZED, got %s", code)
	}

	code = apperrors.GetCode(nil)
	if code != "UNKNOWN_ERROR" {
		t.Errorf("expected code UNKNOWN_ERROR for nil, got %s", code)
	}
}

func TestErrorString(t *testing.T) {
	err := apperrors.New("TEST", "test message", http.StatusBadRequest)
	if err.Error() != "test message" {
		t.Errorf("expected 'test message', got '%s'", err.Error())
	}

	wrapped := apperrors.Wrap(err, "WRAPPED", "wrapped message", http.StatusBadGateway)
	if wrapped.Error() != "wrapped message: test message" {
		t.Errorf("expected 'wrapped message: test message', got '%s'", wrapped.Error())
	}
}
