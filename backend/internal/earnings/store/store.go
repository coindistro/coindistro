package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/coindistro/backend/internal/earnings/models"
)

// Store handles earnings investment persistence.
type Store struct {
	pool *pgxpool.Pool
}

// New creates an earnings store.
func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// GetUserEmail returns the authenticated user's email from the identity store.
// The backend never trusts a client-supplied email — this is the source of truth.
func (s *Store) GetUserEmail(ctx context.Context, userID string) (string, error) {
	var email string
	err := s.pool.QueryRow(ctx,
		`SELECT email FROM identity_users WHERE id = $1 AND deleted_at IS NULL`, userID,
	).Scan(&email)
	if err == pgx.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(email), nil
}

// ─── Settings ────────────────────────────────────────────

func (s *Store) GetSettings(ctx context.Context) (*models.InvestmentSettings, error) {
	var m models.InvestmentSettings
	err := s.pool.QueryRow(ctx, `
		SELECT id, minimum_investment_usd, daily_reward_ngn, max_business_days,
			roi_percent, referral_percent, min_referrals_for_payout,
			early_withdrawal_penalty_percent, early_withdrawal_fee_percent,
			withdrawal_processing_hours, withdrawal_interval_days, enabled,
			updated_by, created_at, updated_at
		FROM investment_settings LIMIT 1`).Scan(
		&m.ID, &m.MinimumInvestmentUSD, &m.DailyRewardNGN, &m.MaxBusinessDays,
		&m.ROIPercent, &m.ReferralPercent, &m.MinReferralsForPayout,
		&m.EarlyWithdrawalPenaltyPercent, &m.EarlyWithdrawalFeePercent,
		&m.WithdrawalProcessingHours, &m.WithdrawalIntervalDays, &m.Enabled,
		&m.UpdatedBy, &m.CreatedAt, &m.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *Store) UpdateSettings(ctx context.Context, m *models.InvestmentSettings) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE investment_settings SET
			minimum_investment_usd = $2, daily_reward_ngn = $3, max_business_days = $4,
			roi_percent = $5, referral_percent = $6, min_referrals_for_payout = $7,
			early_withdrawal_penalty_percent = $8, early_withdrawal_fee_percent = $9,
			withdrawal_processing_hours = $10, withdrawal_interval_days = $11,
			enabled = $12, updated_by = $13, updated_at = NOW()
		WHERE id = $1`,
		m.ID, m.MinimumInvestmentUSD, m.DailyRewardNGN, m.MaxBusinessDays,
		m.ROIPercent, m.ReferralPercent, m.MinReferralsForPayout,
		m.EarlyWithdrawalPenaltyPercent, m.EarlyWithdrawalFeePercent,
		m.WithdrawalProcessingHours, m.WithdrawalIntervalDays, m.Enabled, m.UpdatedBy,
	)
	return err
}

// ─── Exchange Rate ───────────────────────────────────────

func (s *Store) GetExchangeRate(ctx context.Context) (*models.ExchangeRate, error) {
	var m models.ExchangeRate
	err := s.pool.QueryRow(ctx, `
		SELECT id, usd_to_ngn, set_by, created_at, updated_at
		FROM exchange_rates ORDER BY created_at DESC LIMIT 1`).Scan(
		&m.ID, &m.USDTNGN, &m.SetBy, &m.CreatedAt, &m.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *Store) SetExchangeRate(ctx context.Context, rate float64, setBy string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO exchange_rates (id, usd_to_ngn, set_by, created_at, updated_at)
		VALUES (uuid_generate_v4(), $1, $2, NOW(), NOW())`, rate, setBy)
	return err
}

// ─── Fee Tiers ───────────────────────────────────────────

func (s *Store) GetFeeTiers(ctx context.Context) ([]*models.WithdrawalFeeTier, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, min_amount, max_amount, fee_percent, created_at, updated_at
		FROM withdrawal_fee_tiers ORDER BY min_amount ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*models.WithdrawalFeeTier
	for rows.Next() {
		var m models.WithdrawalFeeTier
		if err := rows.Scan(&m.ID, &m.MinAmount, &m.MaxAmount, &m.FeePercent, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &m)
	}
	return list, rows.Err()
}

func (s *Store) GetFeeTierForAmount(ctx context.Context, amount float64) (*models.WithdrawalFeeTier, error) {
	var m models.WithdrawalFeeTier
	err := s.pool.QueryRow(ctx, `
		SELECT id, min_amount, max_amount, fee_percent, created_at, updated_at
		FROM withdrawal_fee_tiers WHERE $1 >= min_amount AND $1 < max_amount
		ORDER BY min_amount ASC LIMIT 1`, amount).Scan(
		&m.ID, &m.MinAmount, &m.MaxAmount, &m.FeePercent, &m.CreatedAt, &m.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *Store) CreateFeeTier(ctx context.Context, m *models.WithdrawalFeeTier) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO withdrawal_fee_tiers (id, min_amount, max_amount, fee_percent, created_at, updated_at)
		VALUES (uuid_generate_v4(), $1, $2, $3, NOW(), NOW())`,
		m.MinAmount, m.MaxAmount, m.FeePercent)
	return err
}

func (s *Store) UpdateFeeTier(ctx context.Context, m *models.WithdrawalFeeTier) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE withdrawal_fee_tiers SET min_amount = $2, max_amount = $3, fee_percent = $4, updated_at = NOW()
		WHERE id = $1`, m.ID, m.MinAmount, m.MaxAmount, m.FeePercent)
	return err
}

func (s *Store) DeleteFeeTier(ctx context.Context, id string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM withdrawal_fee_tiers WHERE id = $1`, id)
	return err
}

// ─── Investments ─────────────────────────────────────────

func (s *Store) CreateInvestment(ctx context.Context, inv *models.EarningsInvestment) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO earnings_investments (
			id, user_id, amount_usd, amount_ngn, exchange_rate,
			payment_provider, payment_reference, payment_status,
			daily_reward_ngn, max_business_days, paid_business_days,
			total_earned_ngn, total_pending_ngn, status,
			maturity_date, started_at, completed_at, paused_at, cancelled_at,
			early_withdrawal_at, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22)`,
		inv.ID, inv.UserID, inv.AmountUSD, inv.AmountNGN, inv.ExchangeRate,
		inv.PaymentProvider, inv.PaymentReference, inv.PaymentStatus,
		inv.DailyRewardNGN, inv.MaxBusinessDays, inv.PaidBusinessDays,
		inv.TotalEarnedNGN, inv.TotalPendingNGN, inv.Status,
		inv.MaturityDate, inv.StartedAt, inv.CompletedAt, inv.PausedAt, inv.CancelledAt,
		inv.EarlyWithdrawalAt, inv.CreatedAt, inv.UpdatedAt,
	)
	return err
}

func (s *Store) UpdateInvestment(ctx context.Context, inv *models.EarningsInvestment) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE earnings_investments SET
			payment_status = $2, status = $3, paid_business_days = $4,
			total_earned_ngn = $5, total_pending_ngn = $6,
			started_at = $7, completed_at = $8, paused_at = $9,
			cancelled_at = $10, early_withdrawal_at = $11,
			maturity_date = $12, updated_at = NOW()
		WHERE id = $1`,
		inv.ID, inv.PaymentStatus, inv.Status, inv.PaidBusinessDays,
		inv.TotalEarnedNGN, inv.TotalPendingNGN,
		inv.StartedAt, inv.CompletedAt, inv.PausedAt, inv.CancelledAt,
		inv.EarlyWithdrawalAt, inv.MaturityDate,
	)
	return err
}

func (s *Store) GetInvestmentByID(ctx context.Context, id string) (*models.EarningsInvestment, error) {
	var inv models.EarningsInvestment
	err := s.pool.QueryRow(ctx, earningsInvestmentSelect+` WHERE i.id = $1`, id).Scan(
		earningsInvestmentScanArgs(&inv)...,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (s *Store) GetInvestmentByReference(ctx context.Context, provider, reference string) (*models.EarningsInvestment, error) {
	var inv models.EarningsInvestment
	err := s.pool.QueryRow(ctx,
		earningsInvestmentSelect+` WHERE i.payment_provider = $1 AND i.payment_reference = $2`, provider, reference).Scan(
		earningsInvestmentScanArgs(&inv)...,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (s *Store) ListUserInvestments(ctx context.Context, userID, status string, page, perPage int) ([]*models.EarningsInvestment, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	where := " WHERE i.user_id = $1"
	args := []interface{}{userID}
	if status != "" {
		where += fmt.Sprintf(" AND i.status = $%d", len(args)+1)
		args = append(args, status)
	}

	var total int
	countQ := "SELECT COUNT(*) FROM earnings_investments i" + where
	if err := s.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	args = append(args, perPage, offset)
	q := earningsInvestmentSelect + where + fmt.Sprintf(" ORDER BY i.created_at DESC LIMIT $%d OFFSET $%d", len(args)-1, len(args))
	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []*models.EarningsInvestment
	for rows.Next() {
		var inv models.EarningsInvestment
		if err := rows.Scan(earningsInvestmentScanArgs(&inv)...); err != nil {
			return nil, 0, err
		}
		list = append(list, &inv)
	}
	return list, total, rows.Err()
}

func (s *Store) ListActiveInvestments(ctx context.Context, limit int) ([]*models.EarningsInvestment, error) {
	if limit <= 0 {
		limit = 500
	}
	rows, err := s.pool.Query(ctx, earningsInvestmentSelect+` WHERE i.status = 'active' ORDER BY i.created_at ASC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*models.EarningsInvestment
	for rows.Next() {
		var inv models.EarningsInvestment
		if err := rows.Scan(earningsInvestmentScanArgs(&inv)...); err != nil {
			return nil, err
		}
		list = append(list, &inv)
	}
	return list, rows.Err()
}

// ListActiveInvestmentsByAmountUSD returns active investments at a given capital (e.g. $30 Genesis).
func (s *Store) ListActiveInvestmentsByAmountUSD(ctx context.Context, amountUSD float64, limit int) ([]*models.EarningsInvestment, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.pool.Query(ctx,
		earningsInvestmentSelect+` WHERE i.status = 'active' AND i.amount_usd = $1 ORDER BY i.created_at ASC LIMIT $2`,
		amountUSD, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*models.EarningsInvestment
	for rows.Next() {
		var inv models.EarningsInvestment
		if err := rows.Scan(earningsInvestmentScanArgs(&inv)...); err != nil {
			return nil, err
		}
		list = append(list, &inv)
	}
	return list, rows.Err()
}

// HasRewardWithStatus reports whether an investment already has a reward of the given status
// (used for idempotent pool-seed credits).
func (s *Store) HasRewardWithStatus(ctx context.Context, investmentID, status string) (bool, error) {
	var n int
	err := s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM investment_rewards WHERE investment_id = $1 AND status = $2`,
		investmentID, status,
	).Scan(&n)
	return n > 0, err
}

// CountSuccessfulReferrals counts active referred users for the referral unlock gate.
// A successful referral is an active identity user who registered with this referrer.
func (s *Store) CountSuccessfulReferrals(ctx context.Context, referrerID string) (int, error) {
	var n int
	// Prefer the referrals table when populated; also count identity_users.referred_by.
	err := s.pool.QueryRow(ctx, `
		SELECT GREATEST(
			(SELECT COUNT(*) FROM identity_users
			 WHERE referred_by = $1 AND deleted_at IS NULL AND status = 'active'),
			(SELECT COUNT(*) FROM referrals
			 WHERE referrer_id = $1 AND status IN ('active', 'converted'))
		)`, referrerID).Scan(&n)
	if err != nil {
		// Fallback if referrals table shape differs
		err2 := s.pool.QueryRow(ctx, `
			SELECT COUNT(*) FROM identity_users
			WHERE referred_by = $1 AND deleted_at IS NULL AND status = 'active'`, referrerID,
		).Scan(&n)
		return n, err2
	}
	return n, nil
}

// WalletBalances is the available/locked view for one currency wallet.
type WalletBalances struct {
	ID        string
	UserID    string
	Currency  string
	Available float64
	Locked    float64
	Total     float64
}

// GetOrCreateCurrencyWallet returns the user's wallet for a currency, creating it if needed.
// Deprecated shape kept for callers that only need id + available.
func (s *Store) GetOrCreateCurrencyWallet(ctx context.Context, userID, currency string) (walletID string, available, total float64, err error) {
	w, err := s.GetOrCreateWalletBalances(ctx, userID, currency)
	if err != nil {
		return "", 0, 0, err
	}
	return w.ID, w.Available, w.Total, nil
}

// GetOrCreateWalletBalances returns full available/locked/total for a currency wallet.
func (s *Store) GetOrCreateWalletBalances(ctx context.Context, userID, currency string) (*WalletBalances, error) {
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency == "" {
		currency = "USD"
	}
	w := &WalletBalances{UserID: userID, Currency: currency}
	err := s.pool.QueryRow(ctx, `
		SELECT id, available_balance, locked_balance, total_balance FROM wallets
		WHERE user_id = $1 AND currency = $2 LIMIT 1`, userID, currency,
	).Scan(&w.ID, &w.Available, &w.Locked, &w.Total)
	if err == nil {
		return w, nil
	}
	if err != pgx.ErrNoRows {
		// currency column may not exist — try without
		err2 := s.pool.QueryRow(ctx, `
			SELECT id, available_balance, locked_balance, total_balance FROM wallets
			WHERE user_id = $1 LIMIT 1`, userID,
		).Scan(&w.ID, &w.Available, &w.Locked, &w.Total)
		if err2 == nil {
			return w, nil
		}
		if err2 != pgx.ErrNoRows {
			return nil, err
		}
	}

	err = s.pool.QueryRow(ctx, `
		INSERT INTO wallets (id, user_id, currency, available_balance, locked_balance, staking_balance, total_balance, updated_at)
		VALUES (uuid_generate_v4(), $1, $2, 0, 0, 0, 0, NOW())
		RETURNING id, available_balance, locked_balance, total_balance`, userID, currency,
	).Scan(&w.ID, &w.Available, &w.Locked, &w.Total)
	if err != nil {
		err = s.pool.QueryRow(ctx, `
			INSERT INTO wallets (id, user_id, available_balance, locked_balance, staking_balance, total_balance, updated_at)
			VALUES (uuid_generate_v4(), $1, 0, 0, 0, 0, NOW())
			ON CONFLICT DO NOTHING
			RETURNING id, available_balance, locked_balance, total_balance`, userID,
		).Scan(&w.ID, &w.Available, &w.Locked, &w.Total)
		if err != nil {
			err = s.pool.QueryRow(ctx, `
				SELECT id, available_balance, locked_balance, total_balance FROM wallets
				WHERE user_id = $1 LIMIT 1`, userID,
			).Scan(&w.ID, &w.Available, &w.Locked, &w.Total)
		}
	}
	if err != nil {
		return nil, err
	}
	return w, nil
}

// WalletTxExists reports whether a ledger row with this reference already exists (idempotency).
func (s *Store) WalletTxExists(ctx context.Context, reference string) (bool, error) {
	var n int
	err := s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM wallet_transactions WHERE reference = $1`, reference,
	).Scan(&n)
	if err != nil {
		// table missing
		if strings.Contains(err.Error(), "wallet_transactions") || strings.Contains(err.Error(), "does not exist") {
			return false, nil
		}
		return false, err
	}
	return n > 0, nil
}

// CreditWalletAvailable credits available + total balance and records a wallet_transactions row.
// Idempotent when the same reference already exists.
func (s *Store) CreditWalletAvailable(ctx context.Context, walletID string, amount float64, txType, reference, description string) error {
	if amount <= 0 {
		return nil
	}
	if exists, err := s.WalletTxExists(ctx, reference); err != nil {
		return err
	} else if exists {
		return nil
	}
	var beforeAvail float64
	err := s.pool.QueryRow(ctx, `SELECT available_balance FROM wallets WHERE id = $1`, walletID).Scan(&beforeAvail)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `
		UPDATE wallets SET
			available_balance = available_balance + $2,
			total_balance = (available_balance + $2) + locked_balance,
			updated_at = NOW()
		WHERE id = $1`, walletID, amount)
	if err != nil {
		return err
	}
	return s.insertWalletTx(ctx, walletID, txType, amount, beforeAvail, beforeAvail+amount, reference, description)
}

// CreditWalletLocked credits locked + total balance (investment capital, profit, referral while locked).
// Idempotent by reference.
func (s *Store) CreditWalletLocked(ctx context.Context, walletID string, amount float64, txType, reference, description string) error {
	if amount <= 0 {
		return nil
	}
	if exists, err := s.WalletTxExists(ctx, reference); err != nil {
		return err
	} else if exists {
		return nil
	}
	var beforeLocked float64
	err := s.pool.QueryRow(ctx, `SELECT locked_balance FROM wallets WHERE id = $1`, walletID).Scan(&beforeLocked)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `
		UPDATE wallets SET
			locked_balance = locked_balance + $2,
			total_balance = available_balance + (locked_balance + $2),
			updated_at = NOW()
		WHERE id = $1`, walletID, amount)
	if err != nil {
		return err
	}
	return s.insertWalletTx(ctx, walletID, txType, amount, beforeLocked, beforeLocked+amount, reference, description)
}

// TransferLockedToAvailable moves amount from locked → available (one ledger unlock entry).
// Idempotent by reference. Does not change total_balance.
func (s *Store) TransferLockedToAvailable(ctx context.Context, walletID string, amount float64, reference, description string) error {
	if amount <= 0 {
		return nil
	}
	if exists, err := s.WalletTxExists(ctx, reference); err != nil {
		return err
	} else if exists {
		return nil
	}
	var locked float64
	err := s.pool.QueryRow(ctx, `SELECT locked_balance FROM wallets WHERE id = $1`, walletID).Scan(&locked)
	if err != nil {
		return err
	}
	if amount > locked {
		amount = locked
	}
	if amount <= 0 {
		return nil
	}
	tag, err := s.pool.Exec(ctx, `
		UPDATE wallets SET
			locked_balance = locked_balance - $2,
			available_balance = available_balance + $2,
			total_balance = (available_balance + $2) + (locked_balance - $2),
			updated_at = NOW()
		WHERE id = $1 AND locked_balance >= $2`, walletID, amount)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("insufficient locked balance for unlock")
	}
	return s.insertWalletTx(ctx, walletID, "unlock", amount, locked, locked-amount, reference, description)
}

func (s *Store) insertWalletTx(ctx context.Context, walletID, txType string, amount, before, after float64, reference, description string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO wallet_transactions (id, wallet_id, type, amount, balance_before, balance_after, reference, description, created_at)
		VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5, $6, $7, NOW())`,
		walletID, txType, amount, before, after, reference, description,
	)
	if err != nil && (strings.Contains(err.Error(), "wallet_transactions") || strings.Contains(err.Error(), "does not exist")) {
		return nil
	}
	return err
}

// MoveAvailableToLocked moves amount from available → locked (correct mis-credited available funds).
// Idempotent by reference.
func (s *Store) MoveAvailableToLocked(ctx context.Context, walletID string, amount float64, reference, description string) error {
	if amount <= 0 {
		return nil
	}
	if exists, err := s.WalletTxExists(ctx, reference); err != nil {
		return err
	} else if exists {
		return nil
	}
	var available float64
	err := s.pool.QueryRow(ctx, `SELECT available_balance FROM wallets WHERE id = $1`, walletID).Scan(&available)
	if err != nil {
		return err
	}
	if amount > available {
		amount = available
	}
	if amount <= 0 {
		return nil
	}
	_, err = s.pool.Exec(ctx, `
		UPDATE wallets SET
			available_balance = available_balance - $2,
			locked_balance = locked_balance + $2,
			updated_at = NOW()
		WHERE id = $1 AND available_balance >= $2`, walletID, amount)
	if err != nil {
		return err
	}
	_, _ = s.pool.Exec(ctx, `
		UPDATE wallets SET total_balance = available_balance + locked_balance, updated_at = NOW() WHERE id = $1`, walletID)
	_, err = s.pool.Exec(ctx, `
		INSERT INTO wallet_transactions (id, wallet_id, type, amount, balance_before, balance_after, reference, description, created_at)
		VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5, $6, $7, NOW())`,
		walletID, "lock_adjust", amount, available, available-amount, reference, description,
	)
	if err != nil && (strings.Contains(err.Error(), "wallet_transactions") || strings.Contains(err.Error(), "does not exist")) {
		return nil
	}
	return err
}

// SumPlatformWalletBalances returns platform-wide available + locked totals for a currency.
func (s *Store) SumPlatformWalletBalances(ctx context.Context, currency string) (available, locked, total float64, users int, err error) {
	currency = strings.ToUpper(strings.TrimSpace(currency))
	err = s.pool.QueryRow(ctx, `
		SELECT
			COALESCE(SUM(available_balance), 0),
			COALESCE(SUM(locked_balance), 0),
			COALESCE(SUM(total_balance), 0),
			COUNT(*)
		FROM wallets WHERE currency = $1`, currency,
	).Scan(&available, &locked, &total, &users)
	if err != nil {
		// fallback without currency filter
		err = s.pool.QueryRow(ctx, `
			SELECT
				COALESCE(SUM(available_balance), 0),
				COALESCE(SUM(locked_balance), 0),
				COALESCE(SUM(total_balance), 0),
				COUNT(*)
			FROM wallets`,
		).Scan(&available, &locked, &total, &users)
	}
	return
}

// CreateRewardReturning inserts a reward and returns its generated id.
func (s *Store) CreateRewardReturning(ctx context.Context, r *models.InvestmentReward) (string, error) {
	var id string
	err := s.pool.QueryRow(ctx, `
		INSERT INTO investment_rewards (id, investment_id, user_id, amount_ngn, reward_date, business_day_number, status, created_at)
		VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5, $6, NOW())
		RETURNING id`,
		r.InvestmentID, r.UserID, r.AmountNGN, r.RewardDate, r.BusinessDayNumber, r.Status,
	).Scan(&id)
	return id, err
}

// InsertActivityEvent writes a lightweight activity/audit trail row when activity_log exists.
func (s *Store) InsertActivityEvent(ctx context.Context, userID, action string, details map[string]interface{}) error {
	payload, _ := json.Marshal(details)
	if len(payload) == 0 {
		payload = []byte("{}")
	}
	// Best-effort: activity_log schema varies; ignore if unavailable.
	_, err := s.pool.Exec(ctx, `
		INSERT INTO activity_log (id, user_id, action, details, created_at)
		VALUES (uuid_generate_v4(), $1, $2, $3::jsonb, NOW())`,
		userID, action, string(payload),
	)
	if err != nil {
		return nil // non-fatal
	}
	return nil
}

func (s *Store) CountInvestmentsByStatus(ctx context.Context, status string) (int, error) {
	var n int
	q := "SELECT COUNT(*) FROM earnings_investments"
	if status != "" {
		q += " WHERE status = $1"
		err := s.pool.QueryRow(ctx, q, status).Scan(&n)
		return n, err
	}
	err := s.pool.QueryRow(ctx, q).Scan(&n)
	return n, err
}

func (s *Store) SumInvestmentsByStatus(ctx context.Context, status string) (float64, error) {
	var sum float64
	q := "SELECT COALESCE(SUM(amount_ngn), 0) FROM earnings_investments"
	if status != "" {
		q += " WHERE status = $1"
		err := s.pool.QueryRow(ctx, q, status).Scan(&sum)
		return sum, err
	}
	err := s.pool.QueryRow(ctx, q).Scan(&sum)
	return sum, err
}

const earningsInvestmentSelect = `
	SELECT i.id, i.user_id, i.amount_usd, i.amount_ngn, i.exchange_rate,
		i.payment_provider, i.payment_reference, i.payment_status,
		i.daily_reward_ngn, i.max_business_days, i.paid_business_days,
		i.total_earned_ngn, i.total_pending_ngn, i.status,
		i.maturity_date, i.started_at, i.completed_at, i.paused_at, i.cancelled_at,
		i.early_withdrawal_at, i.created_at, i.updated_at
	FROM earnings_investments i`

func earningsInvestmentScanArgs(inv *models.EarningsInvestment) []interface{} {
	return []interface{}{
		&inv.ID, &inv.UserID, &inv.AmountUSD, &inv.AmountNGN, &inv.ExchangeRate,
		&inv.PaymentProvider, &inv.PaymentReference, &inv.PaymentStatus,
		&inv.DailyRewardNGN, &inv.MaxBusinessDays, &inv.PaidBusinessDays,
		&inv.TotalEarnedNGN, &inv.TotalPendingNGN, &inv.Status,
		&inv.MaturityDate, &inv.StartedAt, &inv.CompletedAt, &inv.PausedAt, &inv.CancelledAt,
		&inv.EarlyWithdrawalAt, &inv.CreatedAt, &inv.UpdatedAt,
	}
}

// ─── Rewards ─────────────────────────────────────────────

func (s *Store) CreateReward(ctx context.Context, r *models.InvestmentReward) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO investment_rewards (id, investment_id, user_id, amount_ngn, reward_date, business_day_number, status, created_at)
		VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5, $6, NOW())`,
		r.InvestmentID, r.UserID, r.AmountNGN, r.RewardDate, r.BusinessDayNumber, r.Status)
	return err
}

func (s *Store) GetRewardByDate(ctx context.Context, investmentID string, date time.Time) (*models.InvestmentReward, error) {
	var r models.InvestmentReward
	err := s.pool.QueryRow(ctx, `
		SELECT id, investment_id, user_id, amount_ngn, reward_date, business_day_number, status, created_at
		FROM investment_rewards WHERE investment_id = $1 AND reward_date = $2`, investmentID, date).Scan(
		&r.ID, &r.InvestmentID, &r.UserID, &r.AmountNGN, &r.RewardDate, &r.BusinessDayNumber, &r.Status, &r.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *Store) ListRewardsByInvestment(ctx context.Context, investmentID string) ([]*models.InvestmentReward, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, investment_id, user_id, amount_ngn, reward_date, business_day_number, status, created_at
		FROM investment_rewards WHERE investment_id = $1 ORDER BY reward_date ASC`, investmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*models.InvestmentReward
	for rows.Next() {
		var r models.InvestmentReward
		if err := rows.Scan(&r.ID, &r.InvestmentID, &r.UserID, &r.AmountNGN, &r.RewardDate, &r.BusinessDayNumber, &r.Status, &r.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &r)
	}
	return list, rows.Err()
}

func (s *Store) ListRewardsByUser(ctx context.Context, userID string, page, perPage int) ([]*models.InvestmentReward, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	var total int
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM investment_rewards WHERE user_id = $1`, userID).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	rows, err := s.pool.Query(ctx, `
		SELECT id, investment_id, user_id, amount_ngn, reward_date, business_day_number, status, created_at
		FROM investment_rewards WHERE user_id = $1 ORDER BY reward_date DESC LIMIT $2 OFFSET $3`, userID, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []*models.InvestmentReward
	for rows.Next() {
		var r models.InvestmentReward
		if err := rows.Scan(&r.ID, &r.InvestmentID, &r.UserID, &r.AmountNGN, &r.RewardDate, &r.BusinessDayNumber, &r.Status, &r.CreatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, &r)
	}
	return list, total, rows.Err()
}

func (s *Store) SumRewardsByUser(ctx context.Context, userID string) (float64, error) {
	var sum float64
	err := s.pool.QueryRow(ctx, `SELECT COALESCE(SUM(amount_ngn), 0) FROM investment_rewards WHERE user_id = $1`, userID).Scan(&sum)
	return sum, err
}

func (s *Store) SumRewardsByInvestment(ctx context.Context, investmentID string) (float64, error) {
	var sum float64
	err := s.pool.QueryRow(ctx, `SELECT COALESCE(SUM(amount_ngn), 0) FROM investment_rewards WHERE investment_id = $1`, investmentID).Scan(&sum)
	return sum, err
}

func (s *Store) SumTodayRewardsByUser(ctx context.Context, userID string) (float64, error) {
	var sum float64
	err := s.pool.QueryRow(ctx, `SELECT COALESCE(SUM(amount_ngn), 0) FROM investment_rewards WHERE user_id = $1 AND reward_date = CURRENT_DATE`, userID).Scan(&sum)
	return sum, err
}

func (s *Store) SumMonthlyRewardsByUser(ctx context.Context, userID string) (float64, error) {
	var sum float64
	err := s.pool.QueryRow(ctx, `SELECT COALESCE(SUM(amount_ngn), 0) FROM investment_rewards WHERE user_id = $1 AND DATE_TRUNC('month', reward_date) = DATE_TRUNC('month', CURRENT_DATE)`, userID).Scan(&sum)
	return sum, err
}

func (s *Store) SumTodayRewardsAll(ctx context.Context) (float64, error) {
	var sum float64
	err := s.pool.QueryRow(ctx, `SELECT COALESCE(SUM(amount_ngn), 0) FROM investment_rewards WHERE reward_date = CURRENT_DATE`).Scan(&sum)
	return sum, err
}

func (s *Store) SumAllRewards(ctx context.Context) (float64, error) {
	var sum float64
	err := s.pool.QueryRow(ctx, `SELECT COALESCE(SUM(amount_ngn), 0) FROM investment_rewards`).Scan(&sum)
	return sum, err
}

// ─── Withdrawals ─────────────────────────────────────────

func (s *Store) CreateWithdrawal(ctx context.Context, w *models.Withdrawal) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO withdrawals (id, user_id, investment_id, amount_ngn, fee_ngn, penalty_ngn,
			net_amount_ngn, withdrawal_type, status, created_at, updated_at)
		VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())`,
		w.UserID, w.InvestmentID, w.AmountNGN, w.FeeNGN, w.PenaltyNGN,
		w.NetAmountNGN, w.WithdrawalType, w.Status)
	return err
}

func (s *Store) UpdateWithdrawal(ctx context.Context, w *models.Withdrawal) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE withdrawals SET
			status = $2, reviewed_by = $3, reviewed_at = $4,
			processed_at = $5, completed_at = $6, rejection_reason = $7,
			updated_at = NOW()
		WHERE id = $1`,
		w.ID, w.Status, w.ReviewedBy, w.ReviewedAt, w.ProcessedAt, w.CompletedAt, w.RejectionReason)
	return err
}

func (s *Store) GetWithdrawalByID(ctx context.Context, id string) (*models.Withdrawal, error) {
	var w models.Withdrawal
	err := s.pool.QueryRow(ctx, `
		SELECT id, user_id, investment_id, amount_ngn, fee_ngn, penalty_ngn,
			net_amount_ngn, withdrawal_type, status, reviewed_by, reviewed_at,
			processed_at, completed_at, rejection_reason, created_at, updated_at
		FROM withdrawals WHERE id = $1`, id).Scan(
		&w.ID, &w.UserID, &w.InvestmentID, &w.AmountNGN, &w.FeeNGN, &w.PenaltyNGN,
		&w.NetAmountNGN, &w.WithdrawalType, &w.Status, &w.ReviewedBy, &w.ReviewedAt,
		&w.ProcessedAt, &w.CompletedAt, &w.RejectionReason, &w.CreatedAt, &w.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (s *Store) ListUserWithdrawals(ctx context.Context, userID string, page, perPage int) ([]*models.Withdrawal, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	var total int
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM withdrawals WHERE user_id = $1`, userID).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	rows, err := s.pool.Query(ctx, `
		SELECT id, user_id, investment_id, amount_ngn, fee_ngn, penalty_ngn,
			net_amount_ngn, withdrawal_type, status, reviewed_by, reviewed_at,
			processed_at, completed_at, rejection_reason, created_at, updated_at
		FROM withdrawals WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, userID, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []*models.Withdrawal
	for rows.Next() {
		var w models.Withdrawal
		if err := rows.Scan(&w.ID, &w.UserID, &w.InvestmentID, &w.AmountNGN, &w.FeeNGN, &w.PenaltyNGN,
			&w.NetAmountNGN, &w.WithdrawalType, &w.Status, &w.ReviewedBy, &w.ReviewedAt,
			&w.ProcessedAt, &w.CompletedAt, &w.RejectionReason, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, &w)
	}
	return list, total, rows.Err()
}

func (s *Store) ListWithdrawalsByStatus(ctx context.Context, status string, page, perPage int) ([]*models.Withdrawal, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	where := " WHERE 1=1"
	args := []interface{}{}
	if status != "" {
		where = " WHERE status = $1"
		args = append(args, status)
	}

	var total int
	if err := s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM withdrawals"+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	args = append(args, perPage, offset)
	q := fmt.Sprintf(`
		SELECT id, user_id, investment_id, amount_ngn, fee_ngn, penalty_ngn,
			net_amount_ngn, withdrawal_type, status, reviewed_by, reviewed_at,
			processed_at, completed_at, rejection_reason, created_at, updated_at
		FROM withdrawals%s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, len(args)-1, len(args))
	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []*models.Withdrawal
	for rows.Next() {
		var w models.Withdrawal
		if err := rows.Scan(&w.ID, &w.UserID, &w.InvestmentID, &w.AmountNGN, &w.FeeNGN, &w.PenaltyNGN,
			&w.NetAmountNGN, &w.WithdrawalType, &w.Status, &w.ReviewedBy, &w.ReviewedAt,
			&w.ProcessedAt, &w.CompletedAt, &w.RejectionReason, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, &w)
	}
	return list, total, rows.Err()
}

func (s *Store) SumPendingWithdrawalsByUser(ctx context.Context, userID string) (float64, error) {
	var sum float64
	err := s.pool.QueryRow(ctx, `SELECT COALESCE(SUM(amount_ngn), 0) FROM withdrawals WHERE user_id = $1 AND status IN ('pending_review', 'approved', 'processing')`, userID).Scan(&sum)
	return sum, err
}

// GetLastWithdrawal returns the most recent withdrawal request for a user,
// used to enforce the one-withdrawal-every-7-days rule.
func (s *Store) GetLastWithdrawal(ctx context.Context, userID string) (*models.Withdrawal, error) {
	var w models.Withdrawal
	err := s.pool.QueryRow(ctx, `
		SELECT id, user_id, investment_id, amount_ngn, fee_ngn, penalty_ngn,
			net_amount_ngn, withdrawal_type, status, reviewed_by, reviewed_at,
			processed_at, completed_at, rejection_reason, created_at, updated_at
		FROM withdrawals WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1`, userID).Scan(
		&w.ID, &w.UserID, &w.InvestmentID, &w.AmountNGN, &w.FeeNGN, &w.PenaltyNGN,
		&w.NetAmountNGN, &w.WithdrawalType, &w.Status, &w.ReviewedBy, &w.ReviewedAt,
		&w.ProcessedAt, &w.CompletedAt, &w.RejectionReason, &w.CreatedAt, &w.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &w, nil
}

// ─── Payment Transactions ────────────────────────────────

func (s *Store) CreatePaymentTransaction(ctx context.Context, pt *models.EarningsPaymentTransaction) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO earnings_payment_transactions (id, user_id, investment_id, provider, reference,
			type, status, amount_ngn, amount_usd, exchange_rate, response, paid_at, created_at)
		VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW())`,
		pt.UserID, pt.InvestmentID, pt.Provider, pt.Reference, pt.Type, pt.Status,
		pt.AmountNGN, pt.AmountUSD, pt.ExchangeRate, pt.Response, pt.PaidAt)
	return err
}

func (s *Store) UpdatePaymentTransaction(ctx context.Context, id, status string, paidAt *time.Time) error {
	_, err := s.pool.Exec(ctx, `UPDATE earnings_payment_transactions SET status = $2, paid_at = $3 WHERE id = $1`, id, status, paidAt)
	return err
}

func (s *Store) GetPaymentTransactionByReference(ctx context.Context, provider, reference string) (*models.EarningsPaymentTransaction, error) {
	var pt models.EarningsPaymentTransaction
	err := s.pool.QueryRow(ctx, `
		SELECT id, user_id, investment_id, provider, reference, type, status,
			amount_ngn, amount_usd, exchange_rate, response, paid_at, created_at
		FROM earnings_payment_transactions WHERE provider = $1 AND reference = $2`, provider, reference).Scan(
		&pt.ID, &pt.UserID, &pt.InvestmentID, &pt.Provider, &pt.Reference, &pt.Type, &pt.Status,
		&pt.AmountNGN, &pt.AmountUSD, &pt.ExchangeRate, &pt.Response, &pt.PaidAt, &pt.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &pt, nil
}

func (s *Store) ListUserPaymentTransactions(ctx context.Context, userID string, page, perPage int) ([]*models.EarningsPaymentTransaction, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	var total int
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM earnings_payment_transactions WHERE user_id = $1`, userID).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	rows, err := s.pool.Query(ctx, `
		SELECT id, user_id, investment_id, provider, reference, type, status,
			amount_ngn, amount_usd, exchange_rate, response, paid_at, created_at
		FROM earnings_payment_transactions WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, userID, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []*models.EarningsPaymentTransaction
	for rows.Next() {
		var pt models.EarningsPaymentTransaction
		if err := rows.Scan(&pt.ID, &pt.UserID, &pt.InvestmentID, &pt.Provider, &pt.Reference, &pt.Type, &pt.Status,
			&pt.AmountNGN, &pt.AmountUSD, &pt.ExchangeRate, &pt.Response, &pt.PaidAt, &pt.CreatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, &pt)
	}
	return list, total, rows.Err()
}

// ─── Referral Commissions ────────────────────────────────

func (s *Store) CreateReferralCommission(ctx context.Context, rc *models.ReferralCommission) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO referral_commissions (id, referrer_id, referred_id, investment_id,
			amount_usd, amount_ngn, percent, status, created_at)
		VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5, $6, $7, NOW())`,
		rc.ReferrerID, rc.ReferredID, rc.InvestmentID,
		rc.AmountUSD, rc.AmountNGN, rc.Percent, rc.Status)
	return err
}

func (s *Store) GetReferralCommission(ctx context.Context, investmentID, referrerID string) (*models.ReferralCommission, error) {
	var rc models.ReferralCommission
	err := s.pool.QueryRow(ctx, `
		SELECT id, referrer_id, referred_id, investment_id, amount_usd, amount_ngn, percent, status, paid_at, created_at
		FROM referral_commissions WHERE investment_id = $1 AND referrer_id = $2`, investmentID, referrerID).Scan(
		&rc.ID, &rc.ReferrerID, &rc.ReferredID, &rc.InvestmentID, &rc.AmountUSD, &rc.AmountNGN, &rc.Percent, &rc.Status, &rc.PaidAt, &rc.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &rc, nil
}

func (s *Store) ListReferralCommissionsByReferrer(ctx context.Context, referrerID string) ([]*models.ReferralCommission, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, referrer_id, referred_id, investment_id, amount_usd, amount_ngn, percent, status, paid_at, created_at
		FROM referral_commissions WHERE referrer_id = $1 ORDER BY created_at DESC`, referrerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*models.ReferralCommission
	for rows.Next() {
		var rc models.ReferralCommission
		if err := rows.Scan(&rc.ID, &rc.ReferrerID, &rc.ReferredID, &rc.InvestmentID, &rc.AmountUSD, &rc.AmountNGN, &rc.Percent, &rc.Status, &rc.PaidAt, &rc.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &rc)
	}
	return list, rows.Err()
}

func (s *Store) SumReferralCommissionsByReferrer(ctx context.Context, referrerID, status string) (float64, error) {
	var sum float64
	q := "SELECT COALESCE(SUM(amount_ngn), 0) FROM referral_commissions WHERE referrer_id = $1"
	if status != "" {
		q += " AND status = $2"
		err := s.pool.QueryRow(ctx, q, referrerID, status).Scan(&sum)
		return sum, err
	}
	err := s.pool.QueryRow(ctx, q, referrerID).Scan(&sum)
	return sum, err
}

func (s *Store) CountReferralsByReferrer(ctx context.Context, referrerID string) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx, `SELECT COUNT(DISTINCT referred_id) FROM referral_commissions WHERE referrer_id = $1`, referrerID).Scan(&n)
	return n, err
}

func (s *Store) CountActiveReferralsByReferrer(ctx context.Context, referrerID string) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT rc.referred_id)
		FROM referral_commissions rc
		JOIN earnings_investments ei ON ei.user_id = rc.referred_id
		WHERE rc.referrer_id = $1 AND ei.status = 'active'`, referrerID).Scan(&n)
	return n, err
}

// ─── Notifications ───────────────────────────────────────

func (s *Store) CreateNotification(ctx context.Context, n *models.InvestmentNotification) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO investment_notifications (id, user_id, type, title, message, data, created_at)
		VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5, NOW())`,
		n.UserID, n.Type, n.Title, n.Message, n.Data)
	return err
}

func (s *Store) ListUserNotifications(ctx context.Context, userID string, page, perPage int) ([]*models.InvestmentNotification, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	var total int
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM investment_notifications WHERE user_id = $1`, userID).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	rows, err := s.pool.Query(ctx, `
		SELECT id, user_id, type, title, message, data, is_read, read_at, created_at
		FROM investment_notifications WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, userID, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []*models.InvestmentNotification
	for rows.Next() {
		var n models.InvestmentNotification
		if err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Message, &n.Data, &n.IsRead, &n.ReadAt, &n.CreatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, &n)
	}
	return list, total, rows.Err()
}

func (s *Store) MarkNotificationRead(ctx context.Context, id string) error {
	_, err := s.pool.Exec(ctx, `UPDATE investment_notifications SET is_read = true, read_at = NOW() WHERE id = $1`, id)
	return err
}

func (s *Store) MarkAllNotificationsRead(ctx context.Context, userID string) error {
	_, err := s.pool.Exec(ctx, `UPDATE investment_notifications SET is_read = true, read_at = NOW() WHERE user_id = $1 AND is_read = false`, userID)
	return err
}

func (s *Store) CountUnreadNotifications(ctx context.Context, userID string) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM investment_notifications WHERE user_id = $1 AND is_read = false`, userID).Scan(&n)
	return n, err
}

// ─── Stats ───────────────────────────────────────────────

func (s *Store) CountActiveInvestors(ctx context.Context) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx, `SELECT COUNT(DISTINCT user_id) FROM earnings_investments WHERE status = 'active'`).Scan(&n)
	return n, err
}

func (s *Store) CountTotalInvestors(ctx context.Context) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx, `SELECT COUNT(DISTINCT user_id) FROM earnings_investments`).Scan(&n)
	return n, err
}

func (s *Store) SumAllPaidOut(ctx context.Context) (float64, error) {
	var sum float64
	err := s.pool.QueryRow(ctx, `SELECT COALESCE(SUM(total_earned_ngn), 0) FROM earnings_investments WHERE status = 'completed'`).Scan(&sum)
	return sum, err
}
