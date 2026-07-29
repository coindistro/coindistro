package service

import (
	"errors"
	"testing"

	apperrors "github.com/coindistro/backend/internal/errors"
	ide "github.com/coindistro/backend/internal/identity/errors"
	"github.com/coindistro/backend/internal/identity/models"
	"github.com/coindistro/backend/internal/identity/store"
)

// mapLoginLookup mirrors Login() user-lookup error handling for unit testing.
func mapLoginLookup(user *models.User, err error) error {
	if err != nil {
		if errors.Is(err, store.ErrUserNotFound) {
			return ide.ErrInvalidCredentials
		}
		return apperrors.ErrInternalServer
	}
	if user == nil {
		return ide.ErrInvalidCredentials
	}
	return nil
}

func TestLoginLookup_MissingUserIsUnauthorized(t *testing.T) {
	err := mapLoginLookup(nil, store.ErrUserNotFound)
	appErr := apperrors.GetAppError(err)
	if appErr == nil {
		t.Fatal("expected AppError")
	}
	if appErr.Code != "INVALID_CREDENTIALS" {
		t.Fatalf("code = %s", appErr.Code)
	}
	if appErr.StatusCode != 401 {
		t.Fatalf("status = %d, want 401", appErr.StatusCode)
	}
}

func TestLoginLookup_NilUserIsUnauthorized(t *testing.T) {
	err := mapLoginLookup(nil, nil)
	appErr := apperrors.GetAppError(err)
	if appErr == nil || appErr.StatusCode != 401 {
		t.Fatalf("expected 401, got %v", err)
	}
}

func TestLoginLookup_DatabaseFailureIsInternal(t *testing.T) {
	err := mapLoginLookup(nil, errors.New("connection refused"))
	appErr := apperrors.GetAppError(err)
	if appErr == nil || appErr.StatusCode != 500 {
		t.Fatalf("expected 500, got %v", err)
	}
}

func TestLoginLookup_ExistingUserOK(t *testing.T) {
	err := mapLoginLookup(&models.User{ID: "u1", Email: "a@b.com"}, nil)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestAccountStatusMapping(t *testing.T) {
	cases := []struct {
		status string
		code   string
		http   int
	}{
		{"suspended", "ACCOUNT_SUSPENDED", 403},
		{"banned", "ACCOUNT_BANNED", 403},
		{"inactive", "ACCOUNT_INACTIVE", 403},
		{"disabled", "ACCOUNT_INACTIVE", 403},
		{"pending", "ACCOUNT_NOT_VERIFIED", 403},
	}
	for _, tc := range cases {
		t.Run(tc.status, func(t *testing.T) {
			err := mapAccountStatus(tc.status)
			appErr := apperrors.GetAppError(err)
			if appErr == nil {
				t.Fatal("expected error")
			}
			if appErr.Code != tc.code || appErr.StatusCode != tc.http {
				t.Fatalf("got %s/%d want %s/%d", appErr.Code, appErr.StatusCode, tc.code, tc.http)
			}
		})
	}
}

func TestAccountLockedStatus(t *testing.T) {
	appErr := apperrors.GetAppError(ide.ErrAccountLocked)
	if appErr.StatusCode != 423 {
		t.Fatalf("locked status = %d, want 423", appErr.StatusCode)
	}
}

func TestInvalidPasswordMapsToUnauthorized(t *testing.T) {
	appErr := apperrors.GetAppError(ide.ErrInvalidCredentials)
	if appErr.StatusCode != 401 || appErr.Code != "INVALID_CREDENTIALS" {
		t.Fatalf("got %+v", appErr)
	}
}

// mapAccountStatus mirrors Login status switch for unit tests.
func mapAccountStatus(status string) error {
	switch status {
	case "suspended":
		return ide.ErrAccountSuspended
	case "banned":
		return ide.ErrAccountBanned
	case "inactive", "disabled":
		return ide.ErrAccountInactive
	case "pending":
		return ide.ErrAccountNotVerified
	default:
		return nil
	}
}
