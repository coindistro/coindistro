package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/coindistro/backend/internal/earn/models"
)

// Store handles Earn persistence.
type Store struct {
	pool *pgxpool.Pool
}

// New creates an Earn store.
func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func marshalMap(m map[string]interface{}) []byte {
	if m == nil {
		return []byte("{}")
	}
	b, _ := json.Marshal(m)
	return b
}

func unmarshalMap(b []byte) map[string]interface{} {
	if len(b) == 0 {
		return map[string]interface{}{}
	}
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)
	if m == nil {
		return map[string]interface{}{}
	}
	return m
}

// ─── Products ─────────────────────────────────────────

func (s *Store) CreateProduct(ctx context.Context, p *models.Product) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO earn_products (
			id, name, slug, description, category, supported_assets, duration_days,
			capacity_total, capacity_used, status, risk_level, min_allocation, max_allocation,
			reward_model, reward_apr, eligibility, rules, strategy_profiles, featured,
			metadata, starts_at, ends_at, created_by, created_at, updated_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25
		)`,
		p.ID, p.Name, p.Slug, p.Description, p.Category, p.SupportedAssets, p.DurationDays,
		p.CapacityTotal, p.CapacityUsed, p.Status, p.RiskLevel, p.MinAllocation, p.MaxAllocation,
		p.RewardModel, p.RewardAPR, marshalMap(p.Eligibility), marshalMap(p.Rules), p.StrategyProfiles, p.Featured,
		marshalMap(p.Metadata), p.StartsAt, p.EndsAt, p.CreatedBy, p.CreatedAt, p.UpdatedAt,
	)
	return err
}

func (s *Store) UpdateProduct(ctx context.Context, p *models.Product) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE earn_products SET
			name=$2, description=$3, supported_assets=$4, duration_days=$5, capacity_total=$6,
			capacity_used=$7, status=$8, risk_level=$9, min_allocation=$10, max_allocation=$11,
			reward_model=$12, reward_apr=$13, eligibility=$14, rules=$15, strategy_profiles=$16,
			featured=$17, metadata=$18, starts_at=$19, ends_at=$20, updated_at=$21
		WHERE id=$1`,
		p.ID, p.Name, p.Description, p.SupportedAssets, p.DurationDays, p.CapacityTotal,
		p.CapacityUsed, p.Status, p.RiskLevel, p.MinAllocation, p.MaxAllocation,
		p.RewardModel, p.RewardAPR, marshalMap(p.Eligibility), marshalMap(p.Rules), p.StrategyProfiles,
		p.Featured, marshalMap(p.Metadata), p.StartsAt, p.EndsAt, time.Now().UTC(),
	)
	return err
}

func (s *Store) GetProductByID(ctx context.Context, id string) (*models.Product, error) {
	return s.scanProduct(s.pool.QueryRow(ctx, productSelect+" WHERE id=$1", id))
}

func (s *Store) GetProductBySlug(ctx context.Context, slug string) (*models.Product, error) {
	return s.scanProduct(s.pool.QueryRow(ctx, productSelect+" WHERE slug=$1", slug))
}

const productSelect = `
	SELECT id, name, slug, COALESCE(description,''), category, supported_assets, duration_days,
		capacity_total, capacity_used, status, risk_level, min_allocation, max_allocation,
		reward_model, reward_apr, eligibility, rules, strategy_profiles, featured,
		metadata, starts_at, ends_at, created_by, created_at, updated_at
	FROM earn_products`

func (s *Store) scanProduct(row pgx.Row) (*models.Product, error) {
	var p models.Product
	var eligibility, rules, metadata []byte
	err := row.Scan(
		&p.ID, &p.Name, &p.Slug, &p.Description, &p.Category, &p.SupportedAssets, &p.DurationDays,
		&p.CapacityTotal, &p.CapacityUsed, &p.Status, &p.RiskLevel, &p.MinAllocation, &p.MaxAllocation,
		&p.RewardModel, &p.RewardAPR, &eligibility, &rules, &p.StrategyProfiles, &p.Featured,
		&metadata, &p.StartsAt, &p.EndsAt, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	p.Eligibility = unmarshalMap(eligibility)
	p.Rules = unmarshalMap(rules)
	p.Metadata = unmarshalMap(metadata)
	return &p, nil
}

func (s *Store) ListProducts(ctx context.Context, f models.ProductListFilter) ([]*models.Product, int, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PerPage < 1 || f.PerPage > 100 {
		f.PerPage = 20
	}
	where := " WHERE 1=1"
	args := []interface{}{}
	i := 1
	if f.Category != "" {
		where += fmt.Sprintf(" AND category=$%d", i)
		args = append(args, f.Category)
		i++
	}
	if f.Status != "" {
		where += fmt.Sprintf(" AND status=$%d", i)
		args = append(args, f.Status)
		i++
	}
	if f.Featured != nil {
		where += fmt.Sprintf(" AND featured=$%d", i)
		args = append(args, *f.Featured)
		i++
	}
	if f.Asset != "" {
		where += fmt.Sprintf(" AND $%d = ANY(supported_assets)", i)
		args = append(args, f.Asset)
		i++
	}

	var total int
	countQ := "SELECT COUNT(*) FROM earn_products" + where
	if err := s.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (f.Page - 1) * f.PerPage
	args = append(args, f.PerPage, offset)
	q := productSelect + where + fmt.Sprintf(" ORDER BY featured DESC, created_at DESC LIMIT $%d OFFSET $%d", i, i+1)
	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []*models.Product
	for rows.Next() {
		p, err := s.scanProduct(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, p)
	}
	return list, total, rows.Err()
}

// ─── Participations ───────────────────────────────────

func (s *Store) CreateParticipation(ctx context.Context, p *models.Participation) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO earn_participations (
			id, user_id, product_id, asset, allocated_amount, current_balance,
			estimated_rewards, accrued_rewards, lifetime_rewards, status, strategy_profile,
			joined_at, lock_start_at, lock_end_at, completed_at, exited_at, metadata, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)`,
		p.ID, p.UserID, p.ProductID, p.Asset, p.AllocatedAmount, p.CurrentBalance,
		p.EstimatedRewards, p.AccruedRewards, p.LifetimeRewards, p.Status, p.StrategyProfile,
		p.JoinedAt, p.LockStartAt, p.LockEndAt, p.CompletedAt, p.ExitedAt, marshalMap(p.Metadata), p.CreatedAt, p.UpdatedAt,
	)
	return err
}

func (s *Store) UpdateParticipation(ctx context.Context, p *models.Participation) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE earn_participations SET
			allocated_amount=$2, current_balance=$3, estimated_rewards=$4, accrued_rewards=$5,
			lifetime_rewards=$6, status=$7, strategy_profile=$8, lock_start_at=$9, lock_end_at=$10,
			completed_at=$11, exited_at=$12, metadata=$13, updated_at=$14
		WHERE id=$1`,
		p.ID, p.AllocatedAmount, p.CurrentBalance, p.EstimatedRewards, p.AccruedRewards,
		p.LifetimeRewards, p.Status, p.StrategyProfile, p.LockStartAt, p.LockEndAt,
		p.CompletedAt, p.ExitedAt, marshalMap(p.Metadata), time.Now().UTC(),
	)
	return err
}

func (s *Store) GetParticipationByID(ctx context.Context, id string) (*models.Participation, error) {
	return s.scanParticipation(s.pool.QueryRow(ctx, participationSelect+" WHERE id=$1", id))
}

func (s *Store) GetUserProductParticipation(ctx context.Context, userID, productID string) (*models.Participation, error) {
	return s.scanParticipation(s.pool.QueryRow(ctx,
		participationSelect+` WHERE user_id=$1 AND product_id=$2 AND status IN ('active','locked') ORDER BY joined_at DESC LIMIT 1`,
		userID, productID,
	))
}

const participationSelect = `
	SELECT id, user_id, product_id, asset, allocated_amount, current_balance,
		estimated_rewards, accrued_rewards, lifetime_rewards, status, strategy_profile,
		joined_at, lock_start_at, lock_end_at, completed_at, exited_at, metadata, created_at, updated_at
	FROM earn_participations`

func (s *Store) scanParticipation(row pgx.Row) (*models.Participation, error) {
	var p models.Participation
	var meta []byte
	err := row.Scan(
		&p.ID, &p.UserID, &p.ProductID, &p.Asset, &p.AllocatedAmount, &p.CurrentBalance,
		&p.EstimatedRewards, &p.AccruedRewards, &p.LifetimeRewards, &p.Status, &p.StrategyProfile,
		&p.JoinedAt, &p.LockStartAt, &p.LockEndAt, &p.CompletedAt, &p.ExitedAt, &meta, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	p.Metadata = unmarshalMap(meta)
	return &p, nil
}

func (s *Store) ListUserParticipations(ctx context.Context, userID, status string, page, perPage int) ([]*models.Participation, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	where := " WHERE user_id=$1"
	args := []interface{}{userID}
	if status != "" {
		where += " AND status=$2"
		args = append(args, status)
	}
	var total int
	if err := s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM earn_participations"+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * perPage
	args = append(args, perPage, offset)
	q := participationSelect + where + fmt.Sprintf(" ORDER BY joined_at DESC LIMIT $%d OFFSET $%d", len(args)-1, len(args))
	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []*models.Participation
	for rows.Next() {
		p, err := s.scanParticipation(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, p)
	}
	return list, total, rows.Err()
}

func (s *Store) ListActiveParticipations(ctx context.Context, limit int) ([]*models.Participation, error) {
	if limit <= 0 {
		limit = 500
	}
	rows, err := s.pool.Query(ctx, participationSelect+` WHERE status IN ('active','locked') ORDER BY joined_at ASC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*models.Participation
	for rows.Next() {
		p, err := s.scanParticipation(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, rows.Err()
}

func (s *Store) ListProductParticipants(ctx context.Context, productID string, page, perPage int) ([]*models.Participation, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	var total int
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM earn_participations WHERE product_id=$1`, productID).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * perPage
	rows, err := s.pool.Query(ctx, participationSelect+` WHERE product_id=$1 ORDER BY joined_at DESC LIMIT $2 OFFSET $3`, productID, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []*models.Participation
	for rows.Next() {
		p, err := s.scanParticipation(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, p)
	}
	return list, total, rows.Err()
}

func (s *Store) IncrementCapacity(ctx context.Context, productID string, amount float64) error {
	_, err := s.pool.Exec(ctx, `UPDATE earn_products SET capacity_used = capacity_used + $2, updated_at=NOW() WHERE id=$1`, productID, amount)
	return err
}

// ─── Rewards & transactions ───────────────────────────

func (s *Store) CreateReward(ctx context.Context, r *models.Reward) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO earn_rewards (
			id, user_id, product_id, participation_id, asset, amount, reward_type, status,
			description, period_start, period_end, granted_at, metadata, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		r.ID, r.UserID, r.ProductID, r.ParticipationID, r.Asset, r.Amount, r.RewardType, r.Status,
		r.Description, r.PeriodStart, r.PeriodEnd, r.GrantedAt, marshalMap(r.Metadata), r.CreatedAt, r.UpdatedAt,
	)
	return err
}

func (s *Store) UpdateReward(ctx context.Context, r *models.Reward) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE earn_rewards SET status=$2, granted_at=$3, metadata=$4, updated_at=$5 WHERE id=$1`,
		r.ID, r.Status, r.GrantedAt, marshalMap(r.Metadata), time.Now().UTC(),
	)
	return err
}

func (s *Store) ListUserRewards(ctx context.Context, userID string, page, perPage int) ([]*models.Reward, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	var total int
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM earn_rewards WHERE user_id=$1`, userID).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * perPage
	rows, err := s.pool.Query(ctx, `
		SELECT id, user_id, product_id, participation_id, asset, amount, reward_type, status,
			COALESCE(description,''), period_start, period_end, granted_at, metadata, created_at, updated_at
		FROM earn_rewards WHERE user_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, userID, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []*models.Reward
	for rows.Next() {
		var r models.Reward
		var meta []byte
		if err := rows.Scan(
			&r.ID, &r.UserID, &r.ProductID, &r.ParticipationID, &r.Asset, &r.Amount, &r.RewardType, &r.Status,
			&r.Description, &r.PeriodStart, &r.PeriodEnd, &r.GrantedAt, &meta, &r.CreatedAt, &r.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		r.Metadata = unmarshalMap(meta)
		list = append(list, &r)
	}
	return list, total, rows.Err()
}

func (s *Store) SumUserRewardsToday(ctx context.Context, userID string) (float64, error) {
	var sum float64
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount),0) FROM earn_rewards
		WHERE user_id=$1 AND status IN ('calculated','granted')
		AND created_at >= date_trunc('day', NOW() AT TIME ZONE 'UTC')`, userID).Scan(&sum)
	return sum, err
}

func (s *Store) SumUserLifetimeRewards(ctx context.Context, userID string) (float64, error) {
	var sum float64
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount),0) FROM earn_rewards
		WHERE user_id=$1 AND status IN ('calculated','granted')`, userID).Scan(&sum)
	return sum, err
}

func (s *Store) CreateTransaction(ctx context.Context, t *models.Transaction) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO earn_transactions (
			id, user_id, product_id, participation_id, type, asset, amount, balance_after,
			status, reference, description, metadata, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		t.ID, t.UserID, t.ProductID, t.ParticipationID, t.Type, t.Asset, t.Amount, t.BalanceAfter,
		t.Status, t.Reference, t.Description, marshalMap(t.Metadata), t.CreatedAt,
	)
	return err
}

func (s *Store) ListUserTransactions(ctx context.Context, userID string, page, perPage int) ([]*models.Transaction, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	var total int
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM earn_transactions WHERE user_id=$1`, userID).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * perPage
	rows, err := s.pool.Query(ctx, `
		SELECT id, user_id, product_id, participation_id, type, asset, amount, balance_after,
			status, reference, COALESCE(description,''), metadata, created_at
		FROM earn_transactions WHERE user_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, userID, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []*models.Transaction
	for rows.Next() {
		var t models.Transaction
		var meta []byte
		if err := rows.Scan(
			&t.ID, &t.UserID, &t.ProductID, &t.ParticipationID, &t.Type, &t.Asset, &t.Amount, &t.BalanceAfter,
			&t.Status, &t.Reference, &t.Description, &meta, &t.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		t.Metadata = unmarshalMap(meta)
		list = append(list, &t)
	}
	return list, total, rows.Err()
}

// ─── Portfolio / analytics ────────────────────────────

func (s *Store) GetPortfolioOverview(ctx context.Context, userID string) (*models.PortfolioOverview, error) {
	ov := &models.PortfolioOverview{
		AllocationByProduct: map[string]float64{},
		AllocationByAsset:   map[string]float64{},
	}
	rows, err := s.pool.Query(ctx, `
		SELECT product_id, asset, current_balance, estimated_rewards, lifetime_rewards, status
		FROM earn_participations WHERE user_id=$1 AND status IN ('active','locked')`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	products := map[string]struct{}{}
	for rows.Next() {
		var productID, asset, status string
		var bal, est, life float64
		if err := rows.Scan(&productID, &asset, &bal, &est, &life, &status); err != nil {
			return nil, err
		}
		ov.TotalAssetsInEarn += bal
		ov.EstimatedRewards += est
		ov.LifetimeRewards += life
		ov.AllocationByProduct[productID] += bal
		ov.AllocationByAsset[asset] += bal
		products[productID] = struct{}{}
		if status == models.ParticipationLocked {
			ov.LockedBalance += bal
		} else {
			ov.AvailableBalance += bal
		}
	}
	ov.ActiveProducts = len(products)
	todays, _ := s.SumUserRewardsToday(ctx, userID)
	ov.TodaysRewards = todays
	if life, err := s.SumUserLifetimeRewards(ctx, userID); err == nil && life > ov.LifetimeRewards {
		ov.LifetimeRewards = life
	}
	return ov, rows.Err()
}

func (s *Store) GetProductAnalytics(ctx context.Context, productID string) (*models.ProductAnalytics, error) {
	a := &models.ProductAnalytics{ProductID: productID}
	err := s.pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM earn_participations WHERE product_id=$1 AND status IN ('active','locked')),
			(SELECT COALESCE(SUM(allocated_amount),0) FROM earn_participations WHERE product_id=$1 AND status IN ('active','locked')),
			(SELECT COALESCE(SUM(amount),0) FROM earn_rewards WHERE product_id=$1 AND status IN ('calculated','granted')),
			(SELECT capacity_used FROM earn_products WHERE id=$1),
			(SELECT capacity_total FROM earn_products WHERE id=$1)
	`, productID).Scan(&a.Participants, &a.TotalAllocated, &a.TotalRewards, &a.CapacityUsed, &a.CapacityTotal)
	if err != nil {
		return nil, err
	}
	if a.Participants > 0 {
		a.AvgAllocation = a.TotalAllocated / float64(a.Participants)
	}
	if a.CapacityTotal != nil && *a.CapacityTotal > 0 {
		a.CapacityUsedPct = (a.CapacityUsed / *a.CapacityTotal) * 100
	}
	return a, nil
}

func (s *Store) CountActiveProducts(ctx context.Context) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM earn_products WHERE status='active'`).Scan(&n)
	return n, err
}

func (s *Store) CountActiveParticipants(ctx context.Context) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx, `SELECT COUNT(DISTINCT user_id) FROM earn_participations WHERE status IN ('active','locked')`).Scan(&n)
	return n, err
}

// ─── Launchpool / Learn / Referral ────────────────────

func (s *Store) CreateLaunchpool(ctx context.Context, c *models.LaunchpoolCampaign) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO earn_launchpool_campaigns (
			id, product_id, name, description, supported_assets, window_start, window_end,
			allocation_rules, reward_distribution, status, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		c.ID, c.ProductID, c.Name, c.Description, c.SupportedAssets, c.WindowStart, c.WindowEnd,
		marshalMap(c.AllocationRules), marshalMap(c.RewardDistribution), c.Status, c.CreatedAt, c.UpdatedAt,
	)
	return err
}

func (s *Store) ListLaunchpools(ctx context.Context, status string) ([]*models.LaunchpoolCampaign, error) {
	q := `SELECT id, product_id, name, COALESCE(description,''), supported_assets, window_start, window_end,
		allocation_rules, reward_distribution, status, created_at, updated_at FROM earn_launchpool_campaigns`
	args := []interface{}{}
	if status != "" {
		q += " WHERE status=$1"
		args = append(args, status)
	}
	q += " ORDER BY window_start DESC"
	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*models.LaunchpoolCampaign
	for rows.Next() {
		var c models.LaunchpoolCampaign
		var ar, rd []byte
		if err := rows.Scan(&c.ID, &c.ProductID, &c.Name, &c.Description, &c.SupportedAssets, &c.WindowStart, &c.WindowEnd,
			&ar, &rd, &c.Status, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		c.AllocationRules = unmarshalMap(ar)
		c.RewardDistribution = unmarshalMap(rd)
		list = append(list, &c)
	}
	return list, rows.Err()
}

func (s *Store) CreateLearnCampaign(ctx context.Context, c *models.LearnCampaign) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO earn_learn_campaigns (
			id, product_id, name, description, academy_course_id, reward_asset, reward_amount,
			status, starts_at, ends_at, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		c.ID, c.ProductID, c.Name, c.Description, c.AcademyCourseID, c.RewardAsset, c.RewardAmount,
		c.Status, c.StartsAt, c.EndsAt, c.CreatedAt, c.UpdatedAt,
	)
	return err
}

func (s *Store) ListLearnCampaigns(ctx context.Context, status string) ([]*models.LearnCampaign, error) {
	q := `SELECT id, product_id, name, COALESCE(description,''), academy_course_id, reward_asset, reward_amount,
		status, starts_at, ends_at, created_at, updated_at FROM earn_learn_campaigns`
	args := []interface{}{}
	if status != "" {
		q += " WHERE status=$1"
		args = append(args, status)
	}
	q += " ORDER BY created_at DESC"
	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*models.LearnCampaign
	for rows.Next() {
		var c models.LearnCampaign
		if err := rows.Scan(&c.ID, &c.ProductID, &c.Name, &c.Description, &c.AcademyCourseID, &c.RewardAsset, &c.RewardAmount,
			&c.Status, &c.StartsAt, &c.EndsAt, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &c)
	}
	return list, rows.Err()
}

func (s *Store) GetLearnCampaign(ctx context.Context, id string) (*models.LearnCampaign, error) {
	var c models.LearnCampaign
	err := s.pool.QueryRow(ctx, `
		SELECT id, product_id, name, COALESCE(description,''), academy_course_id, reward_asset, reward_amount,
			status, starts_at, ends_at, created_at, updated_at FROM earn_learn_campaigns WHERE id=$1`, id).Scan(
		&c.ID, &c.ProductID, &c.Name, &c.Description, &c.AcademyCourseID, &c.RewardAsset, &c.RewardAmount,
		&c.Status, &c.StartsAt, &c.EndsAt, &c.CreatedAt, &c.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *Store) CreateLearnCompletion(ctx context.Context, c *models.LearnCompletion) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO earn_learn_completions (
			id, campaign_id, user_id, completed_at, reward_eligible, reward_granted, reward_id, metadata, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		c.ID, c.CampaignID, c.UserID, c.CompletedAt, c.RewardEligible, c.RewardGranted, c.RewardID, marshalMap(c.Metadata), c.CreatedAt,
	)
	return err
}

func (s *Store) GetLearnCompletion(ctx context.Context, campaignID, userID string) (*models.LearnCompletion, error) {
	var c models.LearnCompletion
	var meta []byte
	err := s.pool.QueryRow(ctx, `
		SELECT id, campaign_id, user_id, completed_at, reward_eligible, reward_granted, reward_id, metadata, created_at
		FROM earn_learn_completions WHERE campaign_id=$1 AND user_id=$2`, campaignID, userID).Scan(
		&c.ID, &c.CampaignID, &c.UserID, &c.CompletedAt, &c.RewardEligible, &c.RewardGranted, &c.RewardID, &meta, &c.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	c.Metadata = unmarshalMap(meta)
	return &c, nil
}

func (s *Store) ListReferralMilestones(ctx context.Context) ([]*models.ReferralMilestone, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, COALESCE(description,''), required_referrals, reward_asset, reward_amount, status, metadata, created_at, updated_at
		FROM earn_referral_milestones WHERE status='active' ORDER BY required_referrals ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*models.ReferralMilestone
	for rows.Next() {
		var m models.ReferralMilestone
		var meta []byte
		if err := rows.Scan(&m.ID, &m.Name, &m.Description, &m.RequiredReferrals, &m.RewardAsset, &m.RewardAmount, &m.Status, &meta, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		m.Metadata = unmarshalMap(meta)
		list = append(list, &m)
	}
	return list, rows.Err()
}

func (s *Store) ListReferralClaims(ctx context.Context, userID string) ([]*models.ReferralRewardClaim, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, user_id, milestone_id, amount, asset, status, granted_at, metadata, created_at
		FROM earn_referral_reward_claims WHERE user_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*models.ReferralRewardClaim
	for rows.Next() {
		var c models.ReferralRewardClaim
		var meta []byte
		if err := rows.Scan(&c.ID, &c.UserID, &c.MilestoneID, &c.Amount, &c.Asset, &c.Status, &c.GrantedAt, &meta, &c.CreatedAt); err != nil {
			return nil, err
		}
		c.Metadata = unmarshalMap(meta)
		list = append(list, &c)
	}
	return list, rows.Err()
}

func (s *Store) CreatePerformanceSnapshot(ctx context.Context, productID string, date time.Time, participants int, allocated, rewards, pct float64) error {
	id := fmt.Sprintf("%s-%s", productID, date.Format("2006-01-02"))
	// Use uuid from caller ideally; generate simple unique via product+date uniqueness
	_, err := s.pool.Exec(ctx, `
		INSERT INTO earn_performance_snapshots (id, product_id, snapshot_date, participants, total_allocated, total_rewards, capacity_used_pct, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (product_id, snapshot_date) DO UPDATE SET
			participants=EXCLUDED.participants, total_allocated=EXCLUDED.total_allocated,
			total_rewards=EXCLUDED.total_rewards, capacity_used_pct=EXCLUDED.capacity_used_pct`,
		productID, date, participants, allocated, rewards, pct,
	)
	_ = id
	return err
}

func (s *Store) ListProductsByStatus(ctx context.Context, status string) ([]*models.Product, error) {
	list, _, err := s.ListProducts(ctx, models.ProductListFilter{Status: status, Page: 1, PerPage: 500})
	return list, err
}
