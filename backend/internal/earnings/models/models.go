package models

import (
	"time"
)

// ─── Investment Settings ──────────────────────────────────

// InvestmentSettings holds admin-configurable investment parameters.
type InvestmentSettings struct {
	ID                            string    `json:"id" db:"id"`
	MinimumInvestmentUSD          float64   `json:"minimum_investment_usd" db:"minimum_investment_usd"`
	DailyRewardNGN                float64   `json:"daily_reward_ngn" db:"daily_reward_ngn"`
	MaxBusinessDays               int       `json:"max_business_days" db:"max_business_days"`
	ROIPercent                    float64   `json:"roi_percent" db:"roi_percent"`
	ReferralPercent               float64   `json:"referral_percent" db:"referral_percent"`
	MinReferralsForPayout         int       `json:"min_referrals_for_payout" db:"min_referrals_for_payout"`
	EarlyWithdrawalPenaltyPercent float64   `json:"early_withdrawal_penalty_percent" db:"early_withdrawal_penalty_percent"`
	EarlyWithdrawalFeePercent     float64   `json:"early_withdrawal_fee_percent" db:"early_withdrawal_fee_percent"`
	WithdrawalProcessingHours     int       `json:"withdrawal_processing_hours" db:"withdrawal_processing_hours"`
	WithdrawalIntervalDays        int       `json:"withdrawal_interval_days" db:"withdrawal_interval_days"`
	Enabled                       bool      `json:"enabled" db:"enabled"`
	UpdatedBy                     *string   `json:"updated_by,omitempty" db:"updated_by"`
	CreatedAt                     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt                     time.Time `json:"updated_at" db:"updated_at"`
}

// ─── Exchange Rate ────────────────────────────────────────

// ExchangeRate represents the current USD to NGN exchange rate.
type ExchangeRate struct {
	ID        string    `json:"id" db:"id"`
	USDTNGN   float64   `json:"usd_to_ngn" db:"usd_to_ngn"`
	SetBy     *string   `json:"set_by,omitempty" db:"set_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// ─── Withdrawal Fee Tier ──────────────────────────────────

// WithdrawalFeeTier represents a fee tier for withdrawals.
type WithdrawalFeeTier struct {
	ID         string    `json:"id" db:"id"`
	MinAmount  float64   `json:"min_amount" db:"min_amount"`
	MaxAmount  float64   `json:"max_amount" db:"max_amount"`
	FeePercent float64   `json:"fee_percent" db:"fee_percent"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// ─── Earnings Investment ──────────────────────────────────

// InvestmentStatus represents the status of an earnings investment.
type InvestmentStatus string

const (
	InvestmentStatusPendingPayment  InvestmentStatus = "pending_payment"
	InvestmentStatusActive          InvestmentStatus = "active"
	InvestmentStatusCompleted       InvestmentStatus = "completed"
	InvestmentStatusPaused          InvestmentStatus = "paused"
	InvestmentStatusCancelled       InvestmentStatus = "cancelled"
	InvestmentStatusEarlyWithdrawal InvestmentStatus = "early_withdrawal"
)

// EarningsInvestment represents a user's Naira-based earnings investment.
type EarningsInvestment struct {
	ID                string           `json:"id" db:"id"`
	UserID            string           `json:"user_id" db:"user_id"`
	AmountUSD         float64          `json:"amount_usd" db:"amount_usd"`
	AmountNGN         float64          `json:"amount_ngn" db:"amount_ngn"`
	ExchangeRate      float64          `json:"exchange_rate" db:"exchange_rate"`
	PaymentProvider   string           `json:"payment_provider" db:"payment_provider"`
	PaymentReference  string           `json:"payment_reference" db:"payment_reference"`
	PaymentStatus     string           `json:"payment_status" db:"payment_status"`
	DailyRewardNGN    float64          `json:"daily_reward_ngn" db:"daily_reward_ngn"`
	MaxBusinessDays   int              `json:"max_business_days" db:"max_business_days"`
	PaidBusinessDays  int              `json:"paid_business_days" db:"paid_business_days"`
	TotalEarnedNGN    float64          `json:"total_earned_ngn" db:"total_earned_ngn"`
	TotalPendingNGN   float64          `json:"total_pending_ngn" db:"total_pending_ngn"`
	Status            InvestmentStatus `json:"status" db:"status"`
	// IsDemo marks temporary seed/demo investments (safe to purge).
	IsDemo            bool             `json:"is_demo" db:"is_demo"`
	PlanName          string           `json:"plan_name,omitempty" db:"-"`
	MaturityDate      *time.Time       `json:"maturity_date,omitempty" db:"maturity_date"`
	StartedAt         *time.Time       `json:"started_at,omitempty" db:"started_at"`
	CompletedAt       *time.Time       `json:"completed_at,omitempty" db:"completed_at"`
	PausedAt          *time.Time       `json:"paused_at,omitempty" db:"paused_at"`
	CancelledAt       *time.Time       `json:"cancelled_at,omitempty" db:"cancelled_at"`
	EarlyWithdrawalAt *time.Time       `json:"early_withdrawal_at,omitempty" db:"early_withdrawal_at"`
	CreatedAt         time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at" db:"updated_at"`
}

// ─── Investment Reward ────────────────────────────────────

// InvestmentReward represents a single daily reward transaction.
type InvestmentReward struct {
	ID                string    `json:"id" db:"id"`
	InvestmentID      string    `json:"investment_id" db:"investment_id"`
	UserID            string    `json:"user_id" db:"user_id"`
	AmountNGN         float64   `json:"amount_ngn" db:"amount_ngn"`
	RewardDate        time.Time `json:"reward_date" db:"reward_date"`
	BusinessDayNumber int       `json:"business_day_number" db:"business_day_number"`
	Status            string    `json:"status" db:"status"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
}

// ─── Withdrawal ───────────────────────────────────────────

// WithdrawalType represents the type of withdrawal.
type WithdrawalType string

const (
	WithdrawalTypeEarnings WithdrawalType = "earnings"
	WithdrawalTypeEarly    WithdrawalType = "early"
	WithdrawalTypeNormal   WithdrawalType = "normal"
)

// WithdrawalStatus represents the status of a withdrawal.
type WithdrawalStatus string

const (
	WithdrawalStatusPendingReview WithdrawalStatus = "pending_review"
	WithdrawalStatusApproved      WithdrawalStatus = "approved"
	WithdrawalStatusProcessing    WithdrawalStatus = "processing"
	WithdrawalStatusCompleted     WithdrawalStatus = "completed"
	WithdrawalStatusRejected      WithdrawalStatus = "rejected"
)

// Withdrawal represents a withdrawal request.
type Withdrawal struct {
	ID              string           `json:"id" db:"id"`
	UserID          string           `json:"user_id" db:"user_id"`
	InvestmentID    *string          `json:"investment_id,omitempty" db:"investment_id"`
	AmountNGN       float64          `json:"amount_ngn" db:"amount_ngn"`
	FeeNGN          float64          `json:"fee_ngn" db:"fee_ngn"`
	PenaltyNGN      float64          `json:"penalty_ngn" db:"penalty_ngn"`
	NetAmountNGN    float64          `json:"net_amount_ngn" db:"net_amount_ngn"`
	WithdrawalType  WithdrawalType   `json:"withdrawal_type" db:"withdrawal_type"`
	Status          WithdrawalStatus `json:"status" db:"status"`
	ReviewedBy      *string          `json:"reviewed_by,omitempty" db:"reviewed_by"`
	ReviewedAt      *time.Time       `json:"reviewed_at,omitempty" db:"reviewed_at"`
	ProcessedAt     *time.Time       `json:"processed_at,omitempty" db:"processed_at"`
	CompletedAt     *time.Time       `json:"completed_at,omitempty" db:"completed_at"`
	RejectionReason *string          `json:"rejection_reason,omitempty" db:"rejection_reason"`
	CreatedAt       time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at" db:"updated_at"`
}

// ─── Payment Transaction ──────────────────────────────────

// EarningsPaymentTransaction represents a payment transaction record.
type EarningsPaymentTransaction struct {
	ID           string     `json:"id" db:"id"`
	UserID       string     `json:"user_id" db:"user_id"`
	InvestmentID *string    `json:"investment_id,omitempty" db:"investment_id"`
	Provider     string     `json:"provider" db:"provider"`
	Reference    string     `json:"reference" db:"reference"`
	Type         string     `json:"type" db:"type"`
	Status       string     `json:"status" db:"status"`
	AmountNGN    float64    `json:"amount_ngn" db:"amount_ngn"`
	AmountUSD    float64    `json:"amount_usd" db:"amount_usd"`
	ExchangeRate float64    `json:"exchange_rate" db:"exchange_rate"`
	Response     []byte     `json:"response,omitempty" db:"response"`
	PaidAt       *time.Time `json:"paid_at,omitempty" db:"paid_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
}

// ─── Referral Commission ──────────────────────────────────

// ReferralCommission represents a referral reward.
type ReferralCommission struct {
	ID           string     `json:"id" db:"id"`
	ReferrerID   string     `json:"referrer_id" db:"referrer_id"`
	ReferredID   string     `json:"referred_id" db:"referred_id"`
	InvestmentID string     `json:"investment_id" db:"investment_id"`
	AmountUSD    float64    `json:"amount_usd" db:"amount_usd"`
	AmountNGN    float64    `json:"amount_ngn" db:"amount_ngn"`
	Percent      float64    `json:"percent" db:"percent"`
	Status       string     `json:"status" db:"status"`
	PaidAt       *time.Time `json:"paid_at,omitempty" db:"paid_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
}

// ─── Investment Notification ──────────────────────────────

// InvestmentNotification represents a notification about an investment event.
type InvestmentNotification struct {
	ID        string                 `json:"id" db:"id"`
	UserID    string                 `json:"user_id" db:"user_id"`
	Type      string                 `json:"type" db:"type"`
	Title     string                 `json:"title" db:"title"`
	Message   string                 `json:"message" db:"message"`
	Data      map[string]interface{} `json:"data,omitempty" db:"data"`
	IsRead    bool                   `json:"is_read" db:"is_read"`
	ReadAt    *time.Time             `json:"read_at,omitempty" db:"read_at"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
}

// ─── Request DTOs ─────────────────────────────────────────

type InitEarningsPaymentRequest struct {
	AmountUSD float64 `json:"amount_usd"`
	Amount    float64 `json:"amount"` // alias accepted by clients
	Currency  string  `json:"currency"`
}

// Normalize fills AmountUSD/Currency from aliases/defaults.
func (r *InitEarningsPaymentRequest) Normalize() {
	if r.AmountUSD <= 0 && r.Amount > 0 {
		r.AmountUSD = r.Amount
	}
	if r.Currency == "" {
		r.Currency = "NGN"
	}
}

type InitEarningsPaymentResponse struct {
	AuthorizationURL string `json:"authorization_url"`
	Reference        string `json:"reference"`
	AccessCode       string `json:"access_code,omitempty"`
}

type WithdrawalRequest struct {
	InvestmentID *string `json:"investment_id,omitempty"`
	AmountNGN    float64 `json:"amount_ngn"`
	Amount       float64 `json:"amount"` // alias accepted by clients
}

// Normalize fills AmountNGN from the amount alias when needed.
func (r *WithdrawalRequest) Normalize() {
	if r.AmountNGN <= 0 && r.Amount > 0 {
		r.AmountNGN = r.Amount
	}
}

type AdminUpdateSettingsRequest struct {
	MinimumInvestmentUSD          *float64 `json:"minimum_investment_usd"`
	DailyRewardNGN                *float64 `json:"daily_reward_ngn"`
	MaxBusinessDays               *int     `json:"max_business_days"`
	ROIPercent                    *float64 `json:"roi_percent"`
	ReferralPercent               *float64 `json:"referral_percent"`
	MinReferralsForPayout         *int     `json:"min_referrals_for_payout"`
	EarlyWithdrawalPenaltyPercent *float64 `json:"early_withdrawal_penalty_percent"`
	EarlyWithdrawalFeePercent     *float64 `json:"early_withdrawal_fee_percent"`
	WithdrawalProcessingHours     *int     `json:"withdrawal_processing_hours"`
	WithdrawalIntervalDays        *int     `json:"withdrawal_interval_days"`
	Enabled                       *bool    `json:"enabled"`
}

type AdminUpdateExchangeRateRequest struct {
	USDTNGN float64 `json:"usd_to_ngn" binding:"required,gt=0"`
}

type AdminUpdateFeeTierRequest struct {
	MinAmount  float64 `json:"min_amount" binding:"required,gte=0"`
	MaxAmount  float64 `json:"max_amount" binding:"required,gt=0"`
	FeePercent float64 `json:"fee_percent" binding:"required,gt=0"`
}

type AdminWithdrawalActionRequest struct {
	Action          string  `json:"action" binding:"required,oneof=approve reject"`
	RejectionReason *string `json:"rejection_reason,omitempty"`
}

// ─── Response DTOs ────────────────────────────────────────

type EarningsDashboard struct {
	TotalInvestedUSD     float64 `json:"total_invested_usd"`
	TotalInvestedNGN     float64 `json:"total_invested_ngn"`
	// PortfolioValue = capital invested + total profit (earnings). Display only — not free cash.
	PortfolioValueUSD float64 `json:"portfolio_value_usd"`
	PortfolioValueNGN float64 `json:"portfolio_value_ngn"`
	// TotalProfit / ProfitEarned is lifetime investment rewards credited to the investor.
	TotalProfitUSD  float64 `json:"total_profit_usd"`
	TotalProfitNGN  float64 `json:"total_profit_ngn"`
	ProfitEarnedUSD float64 `json:"profit_earned_usd"`
	ProfitEarnedNGN float64 `json:"profit_earned_ngn"`
	// CapitalInvested aliases TotalInvested for portfolio clients.
	CapitalInvestedUSD float64 `json:"capital_invested_usd"`
	CapitalInvestedNGN float64 `json:"capital_invested_ngn"`
	// LockedBalance is capital + still-locked profit/referrals (wallet.locked).
	LockedBalanceUSD float64 `json:"locked_balance_usd"`
	LockedBalanceNGN float64 `json:"locked_balance_ngn"`
	// AvailableBalanceUSD is free cash in the investor wallet (wallet.available).
	AvailableBalanceUSD float64 `json:"available_balance_usd"`
	// WithdrawableBalance is earnings that can be withdrawn when referrals unlock.
	// Zero while WithdrawalsUnlocked is false (capital stays locked separately).
	WithdrawableBalanceNGN float64 `json:"withdrawable_balance_ngn"`
	WithdrawableBalanceUSD float64 `json:"withdrawable_balance_usd"`
	// ROIPercentage = (profit / capital) * 100 across the portfolio.
	ROIPercentage        float64 `json:"roi_percentage"`
	TodayEarningsNGN     float64 `json:"today_earnings_ngn"`
	MonthlyEarningsNGN   float64 `json:"monthly_earnings_ngn"`
	AvailableBalanceNGN  float64 `json:"available_balance_ngn"`
	PendingWithdrawalNGN float64 `json:"pending_withdrawal_ngn"`
	ReferralEarningsNGN  float64 `json:"referral_earnings_ngn"`
	ReferralEarningsUSD  float64 `json:"referral_earnings_usd"`
	ActiveInvestments    int     `json:"active_investments"`
	CompletedInvestments int     `json:"completed_investments"`
	ExchangeRate         float64 `json:"exchange_rate"`
	// Withdrawal gate: unlock when successful referrals >= MinReferralsRequired.
	WithdrawalsUnlocked   bool   `json:"withdrawals_unlocked"`
	WithdrawalLockMessage string `json:"withdrawal_lock_message,omitempty"`
	ActiveReferrals       int    `json:"active_referrals"`
	MinReferralsRequired  int    `json:"min_referrals_required"`
	RemainingReferrals    int    `json:"remaining_referrals"`
	// LastWithdrawalAt is the timestamp of the user's most recent withdrawal request,
	// used to enforce the one-withdrawal-every-7-days rule.
	LastWithdrawalAt *time.Time         `json:"last_withdrawal_at,omitempty"`
	Investments      []*EarningsSummary `json:"investments"`
	ReferralInfo     *ReferralInfo      `json:"referral_info,omitempty"`
}

type EarningsSummary struct {
	ID               string  `json:"id"`
	AmountUSD        float64 `json:"amount_usd"`
	AmountNGN        float64 `json:"amount_ngn"`
	ExchangeRate     float64 `json:"exchange_rate"`
	DailyRewardNGN   float64 `json:"daily_reward_ngn"`
	PaidBusinessDays int     `json:"paid_business_days"`
	MaxBusinessDays  int     `json:"max_business_days"`
	RemainingDays    int     `json:"remaining_days"`
	TotalEarnedNGN   float64 `json:"total_earned_ngn"`
	// TotalEarnedUSD is total_earned_ngn / exchange_rate for display.
	TotalEarnedUSD  float64 `json:"total_earned_usd"`
	TotalPendingNGN float64 `json:"total_pending_ngn"`
	// PortfolioValue = capital + earned profit (position value, not free cash).
	PortfolioValueUSD float64 `json:"portfolio_value_usd"`
	PortfolioValueNGN float64 `json:"portfolio_value_ngn"`
	// ROIPercentage = (earned / capital) * 100 for this position.
	ROIPercentage float64          `json:"roi_percentage"`
	Status        InvestmentStatus `json:"status"`
	ProgressPct   float64          `json:"progress_pct"`
	MaturityDate  *time.Time       `json:"maturity_date,omitempty"`
	StartedAt     *time.Time       `json:"started_at,omitempty"`
	CreatedAt     time.Time        `json:"created_at"`
}

// SeedPoolCreditResult summarizes one investor pool credit.
type SeedPoolCreditResult struct {
	UserID         string  `json:"user_id"`
	InvestmentID   string  `json:"investment_id"`
	AmountUSD      float64 `json:"amount_usd"`
	ProfitUSD      float64 `json:"profit_usd"`
	ProfitNGN      float64 `json:"profit_ngn"`
	PortfolioUSD   float64 `json:"portfolio_usd"`
	AlreadyCredited bool   `json:"already_credited"`
	RewardID       string  `json:"reward_id,omitempty"`
}

// SeedPoolCreditSummary is returned by the Genesis pool seed job.
type SeedPoolCreditSummary struct {
	InvestorsTargeted int                    `json:"investors_targeted"`
	InvestorsCredited int                    `json:"investors_credited"`
	InvestorsSkipped  int                    `json:"investors_skipped"`
	TotalProfitUSD    float64                `json:"total_profit_usd"`
	ProfitPerInvestor float64                `json:"profit_per_investor"`
	PoolUSD           float64                `json:"pool_usd"`
	ExchangeRate      float64                `json:"exchange_rate"`
	Results           []SeedPoolCreditResult `json:"results"`
}

type RewardHistoryItem struct {
	ID                string    `json:"id"`
	AmountNGN         float64   `json:"amount_ngn"`
	RewardDate        string    `json:"reward_date"`
	BusinessDayNumber int       `json:"business_day_number"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
}

type PaymentHistoryItem struct {
	ID        string     `json:"id"`
	AmountNGN float64    `json:"amount_ngn"`
	AmountUSD float64    `json:"amount_usd"`
	Provider  string     `json:"provider"`
	Reference string     `json:"reference"`
	Status    string     `json:"status"`
	PaidAt    *time.Time `json:"paid_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

type WithdrawalHistoryItem struct {
	ID              string           `json:"id"`
	AmountNGN       float64          `json:"amount_ngn"`
	FeeNGN          float64          `json:"fee_ngn"`
	PenaltyNGN      float64          `json:"penalty_ngn"`
	NetAmountNGN    float64          `json:"net_amount_ngn"`
	WithdrawalType  WithdrawalType   `json:"withdrawal_type"`
	Status          WithdrawalStatus `json:"status"`
	RejectionReason *string          `json:"rejection_reason,omitempty"`
	CreatedAt       time.Time        `json:"created_at"`
	CompletedAt     *time.Time       `json:"completed_at,omitempty"`
}

type ReferralInfo struct {
	ReferralCode           string  `json:"referral_code"`
	ReferralLink           string  `json:"referral_link"`
	TotalReferrals         int     `json:"total_referrals"`
	ActiveReferrals        int     `json:"active_referrals"`
	ReferralEarningsNGN    float64 `json:"referral_earnings_ngn"`
	WithdrawableBalanceNGN float64 `json:"withdrawable_balance_ngn"`
	MinimumTarget          int     `json:"minimum_target"`
}

type EarningsChartData struct {
	DailyEarnings   []*ChartPoint `json:"daily_earnings"`
	MonthlyEarnings []*ChartPoint `json:"monthly_earnings"`
	ROIGrowth       []*ChartPoint `json:"roi_growth"`
}

type ChartPoint struct {
	Label string  `json:"label"`
	Value float64 `json:"value"`
}

// ─── Admin DTOs ───────────────────────────────────────────

type AdminEarningsDashboard struct {
	TotalInvestedUSD     float64 `json:"total_invested_usd"`
	TotalInvestedNGN     float64 `json:"total_invested_ngn"`
	TotalPaidOutNGN      float64 `json:"total_paid_out_ngn"`
	ActiveInvestors      int     `json:"active_investors"`
	TotalInvestors       int     `json:"total_investors"`
	PendingWithdrawals   int     `json:"pending_withdrawals"`
	PendingPayments      int     `json:"pending_payments"`
	TodayPayoutNGN       float64 `json:"today_payout_ngn"`
	TotalReferralPaidNGN float64 `json:"total_referral_paid_ngn"`
}
