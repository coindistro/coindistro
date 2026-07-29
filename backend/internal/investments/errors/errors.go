package errors

import (
	"net/http"

	apperrors "github.com/coindistro/backend/internal/errors"
)

var (
	ErrPlanNotFound              = apperrors.New("PLAN_NOT_FOUND", "Investment plan not found", http.StatusNotFound)
	ErrPlanDisabled              = apperrors.New("PLAN_DISABLED", "Investment plan is disabled", http.StatusBadRequest)
	ErrPlanNameExists            = apperrors.New("PLAN_NAME_EXISTS", "Investment plan name already exists", http.StatusConflict)
	ErrInvestmentNotFound        = apperrors.New("INVESTMENT_NOT_FOUND", "Investment not found", http.StatusNotFound)
	ErrInvalidAmount             = apperrors.New("INVALID_AMOUNT", "Amount is outside plan limits", http.StatusBadRequest)
	ErrInvalidLockPeriod         = apperrors.New("INVALID_LOCK_PERIOD", "Lock period is not supported", http.StatusBadRequest)
	ErrPaymentNotFound           = apperrors.New("PAYMENT_NOT_FOUND", "Payment transaction not found", http.StatusNotFound)
	ErrPaymentAlreadyProcessed   = apperrors.New("PAYMENT_ALREADY_PROCESSED", "Payment has already been processed", http.StatusConflict)
	ErrPaymentVerificationFailed = apperrors.New("PAYMENT_VERIFICATION_FAILED", "Payment verification failed", http.StatusBadRequest)
	ErrInvalidSignature          = apperrors.New("INVALID_SIGNATURE", "Invalid webhook signature", http.StatusUnauthorized)
	ErrWalletNotFound            = apperrors.New("WALLET_NOT_FOUND", "Wallet not found", http.StatusNotFound)
	ErrInsufficientBalance       = apperrors.New("INSUFFICIENT_BALANCE", "Insufficient wallet balance", http.StatusBadRequest)
	ErrPricingNotFound           = apperrors.New("PRICING_NOT_FOUND", "CDT pricing not configured", http.StatusNotFound)
	ErrGatewayNotConfigured      = apperrors.New("GATEWAY_NOT_CONFIGURED", "Payment gateway is not configured", http.StatusServiceUnavailable)
	ErrDuplicateWebhook          = apperrors.New("DUPLICATE_WEBHOOK", "Duplicate webhook event", http.StatusConflict)
	ErrInvestmentNotMature       = apperrors.New("INVESTMENT_NOT_MATURE", "Investment has not reached maturity", http.StatusBadRequest)
)
