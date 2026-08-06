package models

import (
	"time"
)

// ─── Investment Plans ──────────────────────────────────

// InvestmentPlan represents a CDT investment plan.
type InvestmentPlan struct {
	ID            string    `json:"id" db:"id"`
	Name          string    `json:"name" db:"name"`
	Description   string    `json:"description" db:"description"`
	MinimumAmount float64   `json:"minimum_amount" db:"minimum_amount"`
	MaximumAmount float64   `json:"maximum_amount" db:"maximum_amount"`
	Currency      string    `json:"currency" db:"currency"`
	ROIPercent    float64   `json:"roi_percent" db:"roi_percent"`
	Enabled       bool      `json:"enabled" db:"enabled"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// ─── Investments ───────────────────────────────────────

// InvestmentStatus represents the status of an investment.
type InvestmentStatus string

const (
	InvestmentStatusPending   InvestmentStatus = "pending"
	InvestmentStatusActive    InvestmentStatus = "active"
	InvestmentStatusCompleted InvestmentStatus = "completed"
	InvestmentStatusFailed    InvestmentStatus = "failed"
	InvestmentStatusCancelled InvestmentStatus = "cancelled"
)

// Investment represents a user's investment.
type Investment struct {
	ID               string           `json:"id" db:"id"`
	UserID           string           `json:"user_id" db:"user_id"`
	PlanID           string           `json:"plan_id" db:"plan_id"`
	PaymentProvider  string           `json:"payment_provider" db:"payment_provider"`
	PaymentReference string           `json:"payment_reference" db:"payment_reference"`
	PaymentStatus    string           `json:"payment_status" db:"payment_status"`
	AmountPaid       float64          `json:"amount_paid" db:"amount_paid"`
	Currency         string           `json:"currency" db:"currency"`
	ExchangeRate     float64          `json:"exchange_rate" db:"exchange_rate"`
	CDTPrice         float64          `json:"cdt_price" db:"cdt_price"`
	AllocatedCDT     float64          `json:"allocated_cdt" db:"allocated_cdt"`
	ROIPercent       float64          `json:"roi_percent" db:"roi_percent"`
	ROICDT           float64          `json:"roi_cdt" db:"roi_cdt"`
	LockPeriodDays   int              `json:"lock_period_days" db:"lock_period_days"`
	Status           InvestmentStatus `json:"status" db:"status"`
	StartedAt        *time.Time       `json:"started_at,omitempty" db:"started_at"`
	MaturesAt        *time.Time       `json:"matures_at,omitempty" db:"matures_at"`
	CompletedAt      *time.Time       `json:"completed_at,omitempty" db:"completed_at"`
	CreatedAt        time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at" db:"updated_at"`

	// Joined fields
	Plan *InvestmentPlan `json:"plan,omitempty"`
}

// ─── Payment Transactions ──────────────────────────────

// PaymentTransaction represents a payment transaction record.
type PaymentTransaction struct {
	ID        string     `json:"id" db:"id"`
	UserID    string     `json:"user_id" db:"user_id"`
	Provider  string     `json:"provider" db:"provider"`
	Reference string     `json:"reference" db:"reference"`
	Status    string     `json:"status" db:"status"`
	Amount    float64    `json:"amount" db:"amount"`
	Currency  string     `json:"currency" db:"currency"`
	Response  []byte     `json:"response,omitempty" db:"response"`
	PaidAt    *time.Time `json:"paid_at,omitempty" db:"paid_at"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

// ─── Wallets ───────────────────────────────────────────

// Wallet represents a user's wallet for a single currency.
// CoinDistro is multi-currency: one user owns many wallets, one per currency
// (NGN, USD, USDT, BTC, ETH, CDT, ...). Currency is always set and unique per user.
type Wallet struct {
	ID               string    `json:"id" db:"id"`
	UserID           string    `json:"user_id" db:"user_id"`
	Currency         string    `json:"currency" db:"currency"`
	AvailableBalance float64   `json:"available_balance" db:"available_balance"`
	LockedBalance    float64   `json:"locked_balance" db:"locked_balance"`
	StakingBalance   float64   `json:"staking_balance" db:"staking_balance"`
	TotalBalance     float64   `json:"total_balance" db:"total_balance"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

// ─── Wallet Transactions ───────────────────────────────

// WalletTransactionType represents the type of wallet transaction.
type WalletTransactionType string

const (
	WalletTxDeposit    WalletTransactionType = "deposit"
	WalletTxInvestment WalletTransactionType = "investment"
	WalletTxROI        WalletTransactionType = "roi"
	WalletTxUnlock     WalletTransactionType = "unlock"
	WalletTxWithdrawal WalletTransactionType = "withdrawal"
)

// WalletTransaction represents a wallet transaction record.
type WalletTransaction struct {
	ID            string                `json:"id" db:"id"`
	WalletID      string                `json:"wallet_id" db:"wallet_id"`
	Type          WalletTransactionType `json:"type" db:"type"`
	Amount        float64               `json:"amount" db:"amount"`
	BalanceBefore float64               `json:"balance_before" db:"balance_before"`
	BalanceAfter  float64               `json:"balance_after" db:"balance_after"`
	Reference     string                `json:"reference" db:"reference"`
	Description   string                `json:"description" db:"description"`
	CreatedAt     time.Time             `json:"created_at" db:"created_at"`
}

// ─── Pricing ───────────────────────────────────────────

// Pricing represents the current CDT price configuration.
type Pricing struct {
	ID        string    `json:"id" db:"id"`
	PriceNGN  float64   `json:"price_ngn" db:"price_ngn"`
	SetBy     string    `json:"set_by" db:"set_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// ─── Request DTOs ──────────────────────────────────────

type CreatePlanRequest struct {
	Name          string  `json:"name" binding:"required,min=2,max=100"`
	Description   string  `json:"description"`
	MinimumAmount float64 `json:"minimum_amount" binding:"required,gt=0"`
	MaximumAmount float64 `json:"maximum_amount" binding:"required,gt=0"`
	Currency      string  `json:"currency" binding:"required,len=3"`
	ROIPercent    float64 `json:"roi_percent" binding:"required,gt=0"`
	Enabled       bool    `json:"enabled"`
}

type UpdatePlanRequest struct {
	Name          *string  `json:"name"`
	Description   *string  `json:"description"`
	MinimumAmount *float64 `json:"minimum_amount"`
	MaximumAmount *float64 `json:"maximum_amount"`
	Currency      *string  `json:"currency"`
	ROIPercent    *float64 `json:"roi_percent"`
	Enabled       *bool    `json:"enabled"`
}

type InvestRequest struct {
	PlanID         string  `json:"plan_id" binding:"required,uuid"`
	Amount         float64 `json:"amount" binding:"required,gt=0"`
	LockPeriodDays int     `json:"lock_period_days" binding:"required,oneof=2 7 30 90 180 365"`
	PaymentGateway string  `json:"payment_gateway" binding:"required,oneof=paystack flutterwave"`
	Currency       string  `json:"currency" binding:"required,len=3"`
}

type InitPaymentRequest struct {
	PlanID         string  `json:"plan_id" binding:"required,uuid"`
	Amount         float64 `json:"amount" binding:"required,gt=0"`
	LockPeriodDays int     `json:"lock_period_days" binding:"required,oneof=2 7 30 90 180 365"`
	Currency       string  `json:"currency" binding:"required,len=3"`
}

// ─── Payment Init Params (stored on payment transaction before verification) ───

// InvestmentParams is serialized into the payment transaction's Response field
// at initialization time so that processSuccessfulPayment can create the
// investment AFTER Paystack/Flutterwave verification succeeds — never before.
type InvestmentParams struct {
	PlanID         string  `json:"plan_id"`
	PlanName       string  `json:"plan_name"`
	ROIPercent     float64 `json:"roi_percent"`
	LockPeriodDays int     `json:"lock_period_days"`
	AllocatedCDT   float64 `json:"allocated_cdt"`
	ROICDT         float64 `json:"roi_cdt"`
	CDTPrice       float64 `json:"cdt_price"`
}

type SetPricingRequest struct {
	PriceNGN float64 `json:"price_ngn" binding:"required,gt=0"`
}

// ─── Response DTOs ─────────────────────────────────────

type InitPaymentResponse struct {
	AuthorizationURL string `json:"authorization_url"`
	Reference        string `json:"reference"`
	AccessCode       string `json:"access_code,omitempty"`
}

type InvestmentDashboard struct {
	TotalInvested        float64              `json:"total_invested"`
	LockedCDT            float64              `json:"locked_cdt"`
	AvailableCDT         float64              `json:"available_cdt"`
	TotalROIEarned       float64              `json:"total_roi_earned"`
	ActiveInvestments    int                  `json:"active_investments"`
	CompletedInvestments int                  `json:"completed_investments"`
	UpcomingMaturity     *time.Time           `json:"upcoming_maturity,omitempty"`
	Investments          []*InvestmentSummary `json:"investments"`
}

type InvestmentSummary struct {
	ID             string           `json:"id"`
	PlanName       string           `json:"plan_name"`
	AmountPaid     float64          `json:"amount_paid"`
	AllocatedCDT   float64          `json:"allocated_cdt"`
	ROICDT         float64          `json:"roi_cdt"`
	ROIPercent     float64          `json:"roi_percent"`
	Status         InvestmentStatus `json:"status"`
	LockPeriodDays int              `json:"lock_period_days"`
	DaysRemaining  int              `json:"days_remaining"`
	ProgressPct    float64          `json:"progress_pct"`
	StartedAt      *time.Time       `json:"started_at,omitempty"`
	MaturesAt      *time.Time       `json:"matures_at,omitempty"`
	CreatedAt      time.Time        `json:"created_at"`
}

type AdminInvestmentStats struct {
	TotalInvested        float64 `json:"total_invested"`
	TotalLockedCDT       float64 `json:"total_locked_cdt"`
	TotalAvailableCDT    float64 `json:"total_available_cdt"`
	TotalROIPaid         float64 `json:"total_roi_paid"`
	ActiveInvestments    int     `json:"active_investments"`
	CompletedInvestments int     `json:"completed_investments"`
	TotalUsers           int     `json:"total_users"`
	PendingPayments      int     `json:"pending_payments"`
	FailedPayments       int     `json:"failed_payments"`
}

// Supported lock periods
var SupportedLockPeriods = []int{2, 7, 30, 90, 180, 365}
