package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/coindistro/backend/internal/investments/models"
)

// Store handles investment persistence.
type Store struct {
	pool *pgxpool.Pool
}

// New creates an investment store.
func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// ─── Investment Plans ─────────────────────────────────

func (s *Store) CreatePlan(ctx context.Context, p *models.InvestmentPlan) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO investment_plans (id, name, description, minimum_amount, maximum_amount, currency, roi_percent, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		p.ID, p.Name, p.Description, p.MinimumAmount, p.MaximumAmount, p.Currency, p.ROIPercent, p.Enabled, p.CreatedAt, p.UpdatedAt,
	)
	return err
}

func (s *Store) UpdatePlan(ctx context.Context, p *models.InvestmentPlan) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE investment_plans SET
			name = $2, description = $3, minimum_amount = $4, maximum_amount = $5,
			currency = $6, roi_percent = $7, enabled = $8, updated_at = $9
		WHERE id = $1`,
		p.ID, p.Name, p.Description, p.MinimumAmount, p.MaximumAmount, p.Currency, p.ROIPercent, p.Enabled, time.Now().UTC(),
	)
	return err
}

func (s *Store) GetPlanByID(ctx context.Context, id string) (*models.InvestmentPlan, error) {
	return s.scanPlan(s.pool.QueryRow(ctx, planSelect+" WHERE id = $1", id))
}

func (s *Store) GetPlanByName(ctx context.Context, name string) (*models.InvestmentPlan, error) {
	return s.scanPlan(s.pool.QueryRow(ctx, planSelect+" WHERE name = $1", name))
}

func (s *Store) ListPlans(ctx context.Context, onlyEnabled bool) ([]*models.InvestmentPlan, error) {
	q := planSelect + " ORDER BY minimum_amount ASC"
	if onlyEnabled {
		q = planSelect + " WHERE enabled = true ORDER BY minimum_amount ASC"
	}
	rows, err := s.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*models.InvestmentPlan
	for rows.Next() {
		p, err := s.scanPlan(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, rows.Err()
}

func (s *Store) DeletePlan(ctx context.Context, id string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM investment_plans WHERE id = $1`, id)
	return err
}

const planSelect = `
	SELECT id, name, COALESCE(description, ''), minimum_amount, maximum_amount, currency, roi_percent, enabled, created_at, updated_at
	FROM investment_plans`

func (s *Store) scanPlan(row pgx.Row) (*models.InvestmentPlan, error) {
	var p models.InvestmentPlan
	err := row.Scan(&p.ID, &p.Name, &p.Description, &p.MinimumAmount, &p.MaximumAmount, &p.Currency, &p.ROIPercent, &p.Enabled, &p.CreatedAt, &p.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// ─── Investments ──────────────────────────────────────

func (s *Store) CreateInvestment(ctx context.Context, inv *models.Investment) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO investments (
			id, user_id, plan_id, payment_provider, payment_reference, payment_status,
			amount_paid, currency, exchange_rate, cdt_price, allocated_cdt, roi_percent, roi_cdt,
			lock_period_days, status, started_at, matures_at, completed_at, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20)`,
		inv.ID, inv.UserID, inv.PlanID, inv.PaymentProvider, inv.PaymentReference, inv.PaymentStatus,
		inv.AmountPaid, inv.Currency, inv.ExchangeRate, inv.CDTPrice, inv.AllocatedCDT, inv.ROIPercent, inv.ROICDT,
		inv.LockPeriodDays, inv.Status, inv.StartedAt, inv.MaturesAt, inv.CompletedAt, inv.CreatedAt, inv.UpdatedAt,
	)
	return err
}

func (s *Store) UpdateInvestment(ctx context.Context, inv *models.Investment) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE investments SET
			payment_status = $2, status = $3, started_at = $4, matures_at = $5,
			completed_at = $6, roi_cdt = $7, updated_at = $8
		WHERE id = $1`,
		inv.ID, inv.PaymentStatus, inv.Status, inv.StartedAt, inv.MaturesAt, inv.CompletedAt, inv.ROICDT, time.Now().UTC(),
	)
	return err
}

func (s *Store) GetInvestmentByID(ctx context.Context, id string) (*models.Investment, error) {
	return s.scanInvestment(s.pool.QueryRow(ctx, investmentSelect+" WHERE i.id = $1", id))
}

func (s *Store) GetInvestmentByReference(ctx context.Context, provider, reference string) (*models.Investment, error) {
	return s.scanInvestment(s.pool.QueryRow(ctx,
		investmentSelect+" WHERE i.payment_provider = $1 AND i.payment_reference = $2", provider, reference))
}

func (s *Store) ListUserInvestments(ctx context.Context, userID string, status string, page, perPage int) ([]*models.Investment, int, error) {
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
	countQ := "SELECT COUNT(*) FROM investments i" + where
	if err := s.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	args = append(args, perPage, offset)
	q := investmentSelect + where + fmt.Sprintf(" ORDER BY i.created_at DESC LIMIT $%d OFFSET $%d", len(args)-1, len(args))
	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []*models.Investment
	for rows.Next() {
		inv, err := s.scanInvestment(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, inv)
	}
	return list, total, rows.Err()
}

func (s *Store) ListInvestmentsByStatus(ctx context.Context, status string, limit int) ([]*models.Investment, error) {
	if limit <= 0 {
		limit = 500
	}
	rows, err := s.pool.Query(ctx, investmentSelect+` WHERE i.status = $1 ORDER BY i.created_at ASC LIMIT $2`, status, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*models.Investment
	for rows.Next() {
		inv, err := s.scanInvestment(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, inv)
	}
	return list, rows.Err()
}

func (s *Store) ListMaturedInvestments(ctx context.Context, limit int) ([]*models.Investment, error) {
	if limit <= 0 {
		limit = 500
	}
	rows, err := s.pool.Query(ctx, investmentSelect+` WHERE i.status = 'active' AND i.matures_at <= NOW() ORDER BY i.matures_at ASC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*models.Investment
	for rows.Next() {
		inv, err := s.scanInvestment(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, inv)
	}
	return list, rows.Err()
}

func (s *Store) ListAllInvestments(ctx context.Context, status string, page, perPage int) ([]*models.Investment, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	where := " WHERE 1=1"
	args := []interface{}{}
	if status != "" {
		where += " AND i.status = $1"
		args = append(args, status)
	}

	var total int
	countQ := "SELECT COUNT(*) FROM investments i" + where
	if err := s.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	args = append(args, perPage, offset)
	q := investmentSelect + where + fmt.Sprintf(" ORDER BY i.created_at DESC LIMIT $%d OFFSET $%d", len(args)-1, len(args))
	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []*models.Investment
	for rows.Next() {
		inv, err := s.scanInvestment(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, inv)
	}
	return list, total, rows.Err()
}

func (s *Store) CountInvestmentsByStatus(ctx context.Context, status string) (int, error) {
	var n int
	q := "SELECT COUNT(*) FROM investments"
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
	q := "SELECT COALESCE(SUM(amount_paid), 0) FROM investments"
	if status != "" {
		q += " WHERE status = $1"
		err := s.pool.QueryRow(ctx, q, status).Scan(&sum)
		return sum, err
	}
	err := s.pool.QueryRow(ctx, q).Scan(&sum)
	return sum, err
}

func (s *Store) SumAllocatedCDTByStatus(ctx context.Context, status string) (float64, error) {
	var sum float64
	q := "SELECT COALESCE(SUM(allocated_cdt), 0) FROM investments"
	if status != "" {
		q += " WHERE status = $1"
		err := s.pool.QueryRow(ctx, q, status).Scan(&sum)
		return sum, err
	}
	err := s.pool.QueryRow(ctx, q).Scan(&sum)
	return sum, err
}

func (s *Store) CountDistinctUsersWithInvestments(ctx context.Context) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx, `SELECT COUNT(DISTINCT user_id) FROM investments`).Scan(&n)
	return n, err
}

const investmentSelect = `
	SELECT i.id, i.user_id, i.plan_id, i.payment_provider, i.payment_reference, i.payment_status,
		i.amount_paid, i.currency, i.exchange_rate, i.cdt_price, i.allocated_cdt, i.roi_percent, i.roi_cdt,
		i.lock_period_days, i.status, i.started_at, i.matures_at, i.completed_at, i.created_at, i.updated_at,
		p.name, p.description, p.minimum_amount, p.maximum_amount, p.currency, p.roi_percent, p.enabled
	FROM investments i
	LEFT JOIN investment_plans p ON i.plan_id = p.id`

func (s *Store) scanInvestment(row pgx.Row) (*models.Investment, error) {
	var inv models.Investment
	var plan models.InvestmentPlan
	err := row.Scan(
		&inv.ID, &inv.UserID, &inv.PlanID, &inv.PaymentProvider, &inv.PaymentReference, &inv.PaymentStatus,
		&inv.AmountPaid, &inv.Currency, &inv.ExchangeRate, &inv.CDTPrice, &inv.AllocatedCDT, &inv.ROIPercent, &inv.ROICDT,
		&inv.LockPeriodDays, &inv.Status, &inv.StartedAt, &inv.MaturesAt, &inv.CompletedAt, &inv.CreatedAt, &inv.UpdatedAt,
		&plan.Name, &plan.Description, &plan.MinimumAmount, &plan.MaximumAmount, &plan.Currency, &plan.ROIPercent, &plan.Enabled,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	plan.ID = inv.PlanID
	inv.Plan = &plan
	return &inv, nil
}

// ─── Payment Transactions ─────────────────────────────

func (s *Store) CreatePaymentTransaction(ctx context.Context, pt *models.PaymentTransaction) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO payment_transactions (id, user_id, provider, reference, status, amount, currency, response, paid_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		pt.ID, pt.UserID, pt.Provider, pt.Reference, pt.Status, pt.Amount, pt.Currency, pt.Response, pt.PaidAt, pt.CreatedAt,
	)
	return err
}

func (s *Store) UpdatePaymentTransaction(ctx context.Context, id, status string, paidAt *time.Time) error {
	_, err := s.pool.Exec(ctx, `UPDATE payment_transactions SET status = $2, paid_at = $3 WHERE id = $1`, id, status, paidAt)
	return err
}

func (s *Store) GetPaymentTransactionByReference(ctx context.Context, provider, reference string) (*models.PaymentTransaction, error) {
	var pt models.PaymentTransaction
	err := s.pool.QueryRow(ctx, `
		SELECT id, user_id, provider, reference, status, amount, currency, response, paid_at, created_at
		FROM payment_transactions WHERE provider = $1 AND reference = $2`, provider, reference).Scan(
		&pt.ID, &pt.UserID, &pt.Provider, &pt.Reference, &pt.Status, &pt.Amount, &pt.Currency, &pt.Response, &pt.PaidAt, &pt.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &pt, nil
}

func (s *Store) ListPaymentTransactions(ctx context.Context, status string, page, perPage int) ([]*models.PaymentTransaction, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	where := " WHERE 1=1"
	args := []interface{}{}
	if status != "" {
		where += " AND status = $1"
		args = append(args, status)
	}

	var total int
	if err := s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM payment_transactions"+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	args = append(args, perPage, offset)
	q := `SELECT id, user_id, provider, reference, status, amount, currency, response, paid_at, created_at
		FROM payment_transactions` + where + fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", len(args)-1, len(args))
	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []*models.PaymentTransaction
	for rows.Next() {
		var pt models.PaymentTransaction
		if err := rows.Scan(&pt.ID, &pt.UserID, &pt.Provider, &pt.Reference, &pt.Status, &pt.Amount, &pt.Currency, &pt.Response, &pt.PaidAt, &pt.CreatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, &pt)
	}
	return list, total, rows.Err()
}

// ─── Wallets ──────────────────────────────────────────

func (s *Store) GetOrCreateWallet(ctx context.Context, userID string) (*models.Wallet, error) {
	// Try to get existing wallet
	w, err := s.scanWallet(s.pool.QueryRow(ctx, walletSelect+" WHERE user_id = $1", userID))
	if err != nil {
		return nil, err
	}
	if w != nil {
		return w, nil
	}

	// Create new wallet
	_, err = s.pool.Exec(ctx, `
		INSERT INTO wallets (id, user_id, available_balance, locked_balance, staking_balance, total_balance, updated_at)
		VALUES (uuid_generate_v4(), $1, 0, 0, 0, 0, NOW())`, userID)
	if err != nil {
		return nil, err
	}

	return s.scanWallet(s.pool.QueryRow(ctx, walletSelect+" WHERE user_id = $1", userID))
}

func (s *Store) GetWalletByUserID(ctx context.Context, userID string) (*models.Wallet, error) {
	return s.scanWallet(s.pool.QueryRow(ctx, walletSelect+" WHERE user_id = $1", userID))
}

func (s *Store) GetWalletByID(ctx context.Context, id string) (*models.Wallet, error) {
	return s.scanWallet(s.pool.QueryRow(ctx, walletSelect+" WHERE id = $1", id))
}

func (s *Store) LockWalletBalance(ctx context.Context, walletID string, amount float64) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE wallets SET
			available_balance = available_balance - $2,
			locked_balance = locked_balance + $2,
			total_balance = total_balance,
			updated_at = NOW()
		WHERE id = $1 AND available_balance >= $2`, walletID, amount)
	return err
}

func (s *Store) UnlockWalletBalance(ctx context.Context, walletID string, amount float64) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE wallets SET
			locked_balance = locked_balance - $2,
			available_balance = available_balance + $2,
			total_balance = total_balance,
			updated_at = NOW()
		WHERE id = $1 AND locked_balance >= $2`, walletID, amount)
	return err
}

func (s *Store) CreditWalletAvailable(ctx context.Context, walletID string, amount float64) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE wallets SET
			available_balance = available_balance + $2,
			total_balance = total_balance + $2,
			updated_at = NOW()
		WHERE id = $1`, walletID, amount)
	return err
}

func (s *Store) CreditWalletLocked(ctx context.Context, walletID string, amount float64) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE wallets SET
			locked_balance = locked_balance + $2,
			total_balance = total_balance + $2,
			updated_at = NOW()
		WHERE id = $1`, walletID, amount)
	return err
}

func (s *Store) ListWallets(ctx context.Context, page, perPage int) ([]*models.Wallet, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	var total int
	if err := s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM wallets").Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	rows, err := s.pool.Query(ctx, walletSelect+" ORDER BY total_balance DESC LIMIT $1 OFFSET $2", perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []*models.Wallet
	for rows.Next() {
		w, err := s.scanWallet(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, w)
	}
	return list, total, rows.Err()
}

const walletSelect = `
	SELECT id, user_id, available_balance, locked_balance, staking_balance, total_balance, updated_at
	FROM wallets`

func (s *Store) scanWallet(row pgx.Row) (*models.Wallet, error) {
	var w models.Wallet
	err := row.Scan(&w.ID, &w.UserID, &w.AvailableBalance, &w.LockedBalance, &w.StakingBalance, &w.TotalBalance, &w.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &w, nil
}

// ─── Wallet Transactions ──────────────────────────────

func (s *Store) CreateWalletTransaction(ctx context.Context, wt *models.WalletTransaction) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO wallet_transactions (id, wallet_id, type, amount, balance_before, balance_after, reference, description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		wt.ID, wt.WalletID, wt.Type, wt.Amount, wt.BalanceBefore, wt.BalanceAfter, wt.Reference, wt.Description, wt.CreatedAt,
	)
	return err
}

func (s *Store) ListWalletTransactions(ctx context.Context, walletID string, page, perPage int) ([]*models.WalletTransaction, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	var total int
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM wallet_transactions WHERE wallet_id = $1`, walletID).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	rows, err := s.pool.Query(ctx, `
		SELECT id, wallet_id, type, amount, balance_before, balance_after, reference, COALESCE(description, ''), created_at
		FROM wallet_transactions WHERE wallet_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, walletID, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []*models.WalletTransaction
	for rows.Next() {
		var wt models.WalletTransaction
		if err := rows.Scan(&wt.ID, &wt.WalletID, &wt.Type, &wt.Amount, &wt.BalanceBefore, &wt.BalanceAfter, &wt.Reference, &wt.Description, &wt.CreatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, &wt)
	}
	return list, total, rows.Err()
}

// ─── Pricing ──────────────────────────────────────────

func (s *Store) GetCurrentPricing(ctx context.Context) (*models.Pricing, error) {
	var p models.Pricing
	err := s.pool.QueryRow(ctx, `
		SELECT id, price_ngn, COALESCE(set_by::text, ''), created_at, updated_at
		FROM cdt_pricing ORDER BY created_at DESC LIMIT 1`).Scan(&p.ID, &p.PriceNGN, &p.SetBy, &p.CreatedAt, &p.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Store) SetPricing(ctx context.Context, price float64, setBy string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO cdt_pricing (id, price_ngn, set_by, created_at, updated_at)
		VALUES (uuid_generate_v4(), $1, $2, NOW(), NOW())`, price, setBy)
	return err
}

func (s *Store) ListPricingHistory(ctx context.Context, limit int) ([]*models.Pricing, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id, price_ngn, COALESCE(set_by::text, ''), created_at, updated_at
		FROM cdt_pricing ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*models.Pricing
	for rows.Next() {
		var p models.Pricing
		if err := rows.Scan(&p.ID, &p.PriceNGN, &p.SetBy, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &p)
	}
	return list, rows.Err()
}

// ─── Webhook Events ───────────────────────────────────

func (s *Store) CreateWebhookEvent(ctx context.Context, provider, eventID, reference string, payload []byte) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO webhook_events (id, provider, event_id, reference, status, payload, created_at)
		VALUES (uuid_generate_v4(), $1, $2, $3, 'received', $4, NOW())
		ON CONFLICT (provider, event_id) DO NOTHING`, provider, eventID, reference, payload)
	return err
}

func (s *Store) MarkWebhookProcessed(ctx context.Context, provider, eventID string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE webhook_events SET status = 'processed', processed_at = NOW()
		WHERE provider = $1 AND event_id = $2`, provider, eventID)
	return err
}

func (s *Store) IsWebhookProcessed(ctx context.Context, provider, eventID string) (bool, error) {
	var status string
	err := s.pool.QueryRow(ctx, `SELECT status FROM webhook_events WHERE provider = $1 AND event_id = $2`, provider, eventID).Scan(&status)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return status == "processed", nil
}

func (s *Store) ListWebhookEvents(ctx context.Context, provider, status string, page, perPage int) ([]map[string]interface{}, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	where := " WHERE 1=1"
	args := []interface{}{}
	if provider != "" {
		where += " AND provider = $1"
		args = append(args, provider)
	}
	if status != "" {
		where += fmt.Sprintf(" AND status = $%d", len(args)+1)
		args = append(args, status)
	}

	var total int
	if err := s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM webhook_events"+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	args = append(args, perPage, offset)
	q := `SELECT id, provider, event_id, COALESCE(reference, ''), status, created_at, processed_at
		FROM webhook_events` + where + fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", len(args)-1, len(args))
	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []map[string]interface{}
	for rows.Next() {
		var id, provider, eventID, reference, status string
		var createdAt time.Time
		var processedAt *time.Time
		if err := rows.Scan(&id, &provider, &eventID, &reference, &status, &createdAt, &processedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, map[string]interface{}{
			"id":           id,
			"provider":     provider,
			"event_id":     eventID,
			"reference":    reference,
			"status":       status,
			"created_at":   createdAt,
			"processed_at": processedAt,
		})
	}
	return list, total, rows.Err()
}
