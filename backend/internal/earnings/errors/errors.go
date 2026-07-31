package errors

import (
	"net/http"

	apperrors "github.com/coindistro/backend/internal/errors"
)

var (
	ErrInvestmentNotFound          = apperrors.New("EARNINGS_INVESTMENT_NOT_FOUND", "Investment not found", http.StatusNotFound)
	ErrInvestmentNotActive         = apperrors.New("EARNINGS_INVESTMENT_NOT_ACTIVE", "Investment is not active", http.StatusBadRequest)
	ErrInvestmentAlreadyCompleted  = apperrors.New("EARNINGS_INVESTMENT_ALREADY_COMPLETED", "Investment already completed", http.StatusBadRequest)
	ErrInvestmentNotMature         = apperrors.New("EARNINGS_INVESTMENT_NOT_MATURE", "Investment has not reached maturity", http.StatusBadRequest)
	ErrInvalidAmount               = apperrors.New("EARNINGS_INVALID_AMOUNT", "Invalid investment amount", http.StatusBadRequest)
	ErrPaymentVerificationFailed   = apperrors.New("EARNINGS_PAYMENT_VERIFICATION_FAILED", "Payment verification failed", http.StatusBadRequest)
	ErrPaymentAlreadyProcessed     = apperrors.New("EARNINGS_PAYMENT_ALREADY_PROCESSED", "Payment already processed", http.StatusConflict)
	ErrInvalidSignature            = apperrors.New("EARNINGS_INVALID_SIGNATURE", "Invalid webhook signature", http.StatusUnauthorized)
	ErrGatewayNotConfigured        = apperrors.New("EARNINGS_GATEWAY_NOT_CONFIGURED", "Payment gateway not configured", http.StatusServiceUnavailable)
	ErrDuplicateWebhook            = apperrors.New("EARNINGS_DUPLICATE_WEBHOOK", "Duplicate webhook event", http.StatusConflict)
	ErrSettingsNotFound            = apperrors.New("EARNINGS_SETTINGS_NOT_FOUND", "Investment settings not configured", http.StatusNotFound)
	ErrExchangeRateNotFound        = apperrors.New("EARNINGS_EXCHANGE_RATE_NOT_FOUND", "Exchange rate not configured", http.StatusNotFound)
	ErrMinimumInvestmentNotMet     = apperrors.New("EARNINGS_MINIMUM_INVESTMENT_NOT_MET", "Minimum investment is $30", http.StatusBadRequest)
	ErrWithdrawalNotFound          = apperrors.New("EARNINGS_WITHDRAWAL_NOT_FOUND", "Withdrawal not found", http.StatusNotFound)
	ErrWithdrawalAlreadyProcessed  = apperrors.New("EARNINGS_WITHDRAWAL_ALREADY_PROCESSED", "Withdrawal already processed", http.StatusConflict)
	ErrInsufficientBalance         = apperrors.New("EARNINGS_INSUFFICIENT_BALANCE", "Insufficient balance", http.StatusBadRequest)
	ErrNotBusinessDay              = apperrors.New("EARNINGS_NOT_BUSINESS_DAY", "Today is not a business day", http.StatusBadRequest)
	ErrReferralNotFound            = apperrors.New("EARNINGS_REFERRAL_NOT_FOUND", "Referral not found", http.StatusNotFound)
	ErrDuplicateReferralCommission = apperrors.New("EARNINGS_DUPLICATE_REFERRAL_COMMISSION", "Duplicate referral commission", http.StatusConflict)
	ErrSystemDisabled              = apperrors.New("EARNINGS_SYSTEM_DISABLED", "Investment system is currently disabled", http.StatusServiceUnavailable)
)
