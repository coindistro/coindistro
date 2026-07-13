package errors

import (
	"net/http"

	apperrors "github.com/coindistro/backend/internal/errors"
)

// Identity-specific errors
var (
	// Registration errors
	ErrReferralRequired     = apperrors.New("REFERRAL_REQUIRED", "You need a valid invitation code to join Coindistro", http.StatusBadRequest)
	ErrInvalidReferralCode  = apperrors.New("INVALID_REFERRAL_CODE", "The referral code is invalid or has expired", http.StatusBadRequest)
	ErrReferralSelfReferral = apperrors.New("SELF_REFERRAL", "You cannot refer yourself", http.StatusBadRequest)
	ErrReferralAlreadyUsed  = apperrors.New("REFERRAL_ALREADY_USED", "This referral code has already been used", http.StatusConflict)
	ErrEmailAlreadyExists   = apperrors.New("EMAIL_ALREADY_EXISTS", "An account with this email already exists", http.StatusConflict)
	ErrUsernameTaken        = apperrors.New("USERNAME_TAKEN", "This username is already taken", http.StatusConflict)
	ErrRegistrationDisabled = apperrors.New("REGISTRATION_DISABLED", "Registration is currently disabled", http.StatusServiceUnavailable)
	ErrInviteOnly           = apperrors.New("INVITE_ONLY", "Coindistro is currently invite-only. You need an invitation code to register.", http.StatusForbidden)
	ErrNoInvitationCredits  = apperrors.New("NO_INVITATION_CREDITS", "You have no invitation credits remaining", http.StatusForbidden)

	// Authentication errors
	ErrInvalidCredentials  = apperrors.New("INVALID_CREDENTIALS", "Invalid email or password", http.StatusUnauthorized)
	ErrAccountLocked       = apperrors.New("ACCOUNT_LOCKED", "Account is temporarily locked due to too many failed attempts", http.StatusUnauthorized)
	ErrAccountNotVerified  = apperrors.New("ACCOUNT_NOT_VERIFIED", "Please verify your email before logging in", http.StatusForbidden)
	ErrAccountSuspended    = apperrors.New("ACCOUNT_SUSPENDED", "Your account has been suspended", http.StatusForbidden)
	ErrAccountBanned       = apperrors.New("ACCOUNT_BANNED", "Your account has been banned", http.StatusForbidden)
	ErrSessionExpired      = apperrors.New("SESSION_EXPIRED", "Your session has expired. Please log in again.", http.StatusUnauthorized)
	ErrSessionRevoked      = apperrors.New("SESSION_REVOKED", "Your session has been revoked", http.StatusUnauthorized)
	ErrInvalidRefreshToken = apperrors.New("INVALID_REFRESH_TOKEN", "Invalid or expired refresh token", http.StatusUnauthorized)

	// Verification errors
	ErrInvalidVerificationToken = apperrors.New("INVALID_VERIFICATION_TOKEN", "Invalid or expired verification token", http.StatusBadRequest)
	ErrAlreadyVerified          = apperrors.New("ALREADY_VERIFIED", "Email is already verified", http.StatusConflict)
	ErrVerificationTooSoon      = apperrors.New("VERIFICATION_TOO_SOON", "Please wait before requesting another verification email", http.StatusTooManyRequests)

	// Password errors
	ErrInvalidResetToken = apperrors.New("INVALID_RESET_TOKEN", "Invalid or expired password reset token", http.StatusBadRequest)
	ErrPasswordTooWeak   = apperrors.New("PASSWORD_TOO_WEAK", "Password must be at least 8 characters with uppercase, lowercase, number, and special character", http.StatusBadRequest)
	ErrPasswordSameAsOld = apperrors.New("PASSWORD_SAME_AS_OLD", "New password must be different from current password", http.StatusBadRequest)
	ErrResetTooSoon      = apperrors.New("RESET_TOO_SOON", "Please wait before requesting another password reset", http.StatusTooManyRequests)

	// Session errors
	ErrSessionNotFound        = apperrors.New("SESSION_NOT_FOUND", "Session not found", http.StatusNotFound)
	ErrCannotTerminateCurrent = apperrors.New("CANNOT_TERMINATE_CURRENT", "Cannot terminate current session from this endpoint", http.StatusBadRequest)

	// Invitation errors
	ErrInvitationExpired     = apperrors.New("INVITATION_EXPIRED", "This invitation has expired", http.StatusGone)
	ErrInvitationAlreadyUsed = apperrors.New("INVITATION_ALREADY_USED", "This invitation has already been used", http.StatusConflict)
	ErrInvitationNotFound    = apperrors.New("INVITATION_NOT_FOUND", "Invitation not found", http.StatusNotFound)

	// Genesis errors
	ErrGenesisFull          = apperrors.New("GENESIS_FULL", "The genesis program has reached its maximum capacity", http.StatusConflict)
	ErrGenesisAlreadyMember = apperrors.New("GENESIS_ALREADY_MEMBER", "User is already a genesis member", http.StatusConflict)
)
