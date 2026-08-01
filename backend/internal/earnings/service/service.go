package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/coindistro/backend/internal/audit"
	"github.com/coindistro/backend/internal/earnings/errors"
	"github.com/coindistro/backend/internal/earnings/models"
	"github.com/coindistro/backend/internal/earnings/store"
	apperrors "github.com/coindistro/backend/internal/errors"
	"github.com/coindistro/backend/internal/events"
	uuidlib "github.com/coindistro/backend/internal/uuid"
	"github.com/coindistro/backend/internal/workers"
)

// Service implements earnings investment business logic.
type Service struct {
	store       *store.Store
	eventBus    *events.InMemoryBus
	jobRegistry *workers.Registry
	workerPool  *workers.Pool
	auditLogger *audit.Logger
	logger      *zap.Logger
	cfg         Config
}

// Config holds the earnings service configuration.
type Config struct {
	BaseURL               string
	PaystackSecretKey     string
	PaystackPublicKey     string
	FlutterwaveSecretKey  string
	FlutterwavePublicKey  string
	FlutterwaveSecretHash string
	AppURL                string
}

// New creates the earnings service.
func New(
	st *store.Store,
	eventBus *events.InMemoryBus,
	jobRegistry *workers.Registry,
	workerPool *workers.Pool,
	auditLogger *audit.Logger,
	logger *zap.Logger,
	cfg Config,
) *Service {
	svc := &Service{
		store:       st,
		eventBus:    eventBus,
		jobRegistry: jobRegistry,
		workerPool:  workerPool,
		auditLogger: auditLogger,
		logger:      logger,
		cfg:         cfg,
	}
	svc.registerWorkers()
	return svc
}

// ─── Settings ────────────────────────────────────────────

func defaultInvestmentSettings() *models.InvestmentSettings {
	return &models.InvestmentSettings{
		MinimumInvestmentUSD:          30,
		DailyRewardNGN:                650,
		MaxBusinessDays:               20,
		ROIPercent:                    30,
		ReferralPercent:               10,
		MinReferralsForPayout:         5,
		EarlyWithdrawalPenaltyPercent: 15,
		EarlyWithdrawalFeePercent:     5,
		WithdrawalProcessingHours:     24,
		Enabled:                       true,
	}
}

func defaultExchangeRate() *models.ExchangeRate {
	return &models.ExchangeRate{USDTNGN: 1400}
}

func (s *Service) GetSettings(ctx context.Context) (*models.InvestmentSettings, error) {
	settings, err := s.store.GetSettings(ctx)
	if err != nil {
		return nil, err
	}
	if settings == nil {
		return defaultInvestmentSettings(), nil
	}
	return settings, nil
}

func (s *Service) UpdateSettings(ctx context.Context, req *models.AdminUpdateSettingsRequest, actorID string) (*models.InvestmentSettings, error) {
	settings, err := s.store.GetSettings(ctx)
	if err != nil {
		return nil, err
	}
	if settings == nil {
		return nil, errors.ErrSettingsNotFound
	}

	if req.MinimumInvestmentUSD != nil {
		settings.MinimumInvestmentUSD = *req.MinimumInvestmentUSD
	}
	if req.DailyRewardNGN != nil {
		settings.DailyRewardNGN = *req.DailyRewardNGN
	}
	if req.MaxBusinessDays != nil {
		settings.MaxBusinessDays = *req.MaxBusinessDays
	}
	if req.ROIPercent != nil {
		settings.ROIPercent = *req.ROIPercent
	}
	if req.ReferralPercent != nil {
		settings.ReferralPercent = *req.ReferralPercent
	}
	if req.MinReferralsForPayout != nil {
		settings.MinReferralsForPayout = *req.MinReferralsForPayout
	}
	if req.EarlyWithdrawalPenaltyPercent != nil {
		settings.EarlyWithdrawalPenaltyPercent = *req.EarlyWithdrawalPenaltyPercent
	}
	if req.EarlyWithdrawalFeePercent != nil {
		settings.EarlyWithdrawalFeePercent = *req.EarlyWithdrawalFeePercent
	}
	if req.WithdrawalProcessingHours != nil {
		settings.WithdrawalProcessingHours = *req.WithdrawalProcessingHours
	}
	if req.Enabled != nil {
		settings.Enabled = *req.Enabled
	}
	settings.UpdatedBy = &actorID

	if err := s.store.UpdateSettings(ctx, settings); err != nil {
		return nil, err
	}

	s.audit(ctx, actorID, audit.ActionAdminAction, "investment_settings", settings.ID, map[string]interface{}{
		"daily_reward_ngn": settings.DailyRewardNGN,
		"enabled":          settings.Enabled,
	})
	return settings, nil
}

// ─── Exchange Rate ───────────────────────────────────────

func (s *Service) GetExchangeRate(ctx context.Context) (*models.ExchangeRate, error) {
	rate, err := s.store.GetExchangeRate(ctx)
	if err != nil {
		return nil, err
	}
	if rate == nil {
		return defaultExchangeRate(), nil
	}
	return rate, nil
}

func (s *Service) SetExchangeRate(ctx context.Context, rate float64, actorID string) (*models.ExchangeRate, error) {
	if err := s.store.SetExchangeRate(ctx, rate, actorID); err != nil {
		return nil, err
	}
	return s.GetExchangeRate(ctx)
}

// ─── Fee Tiers ───────────────────────────────────────────

func (s *Service) GetFeeTiers(ctx context.Context) ([]*models.WithdrawalFeeTier, error) {
	return s.store.GetFeeTiers(ctx)
}

func (s *Service) CreateFeeTier(ctx context.Context, req *models.AdminUpdateFeeTierRequest) (*models.WithdrawalFeeTier, error) {
	m := &models.WithdrawalFeeTier{
		MinAmount:  req.MinAmount,
		MaxAmount:  req.MaxAmount,
		FeePercent: req.FeePercent,
	}
	if err := s.store.CreateFeeTier(ctx, m); err != nil {
		return nil, err
	}
	tiers, err := s.store.GetFeeTiers(ctx)
	if err != nil {
		return nil, err
	}
	if len(tiers) > 0 {
		return tiers[len(tiers)-1], nil
	}
	return m, nil
}

func (s *Service) UpdateFeeTier(ctx context.Context, id string, req *models.AdminUpdateFeeTierRequest) (*models.WithdrawalFeeTier, error) {
	m := &models.WithdrawalFeeTier{
		ID:         id,
		MinAmount:  req.MinAmount,
		MaxAmount:  req.MaxAmount,
		FeePercent: req.FeePercent,
	}
	if err := s.store.UpdateFeeTier(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *Service) DeleteFeeTier(ctx context.Context, id string) error {
	return s.store.DeleteFeeTier(ctx, id)
}

// ─── Payment Initialization ──────────────────────────────

func (s *Service) InitPaystackPayment(ctx context.Context, userID string, req *models.InitEarningsPaymentRequest) (*models.InitEarningsPaymentResponse, error) {
	req.Normalize()
	settings, rate, err := s.validateInvestmentRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	amountNGN := req.AmountUSD * rate.USDTNGN
	reference := fmt.Sprintf("EARN-PS-%s-%d", uuidlib.NewString()[:8], time.Now().Unix())

	// Create pending payment transaction
	pt := &models.EarningsPaymentTransaction{
		UserID:       userID,
		Provider:     "paystack",
		Reference:    reference,
		Type:         "investment",
		Status:       "pending",
		AmountNGN:    amountNGN,
		AmountUSD:    req.AmountUSD,
		ExchangeRate: rate.USDTNGN,
		PaidAt:       nil,
	}
	if err := s.store.CreatePaymentTransaction(ctx, pt); err != nil {
		return nil, err
	}

	// Call Paystack initialize API
	authURL, accessCode, err := s.callPaystackInitialize(ctx, amountNGN, req.Currency, reference, userID)
	if err != nil {
		return nil, err
	}

	// Create pending investment record
	now := time.Now().UTC()
	totalPending := float64(settings.MaxBusinessDays) * settings.DailyRewardNGN
	inv := &models.EarningsInvestment{
		ID:               uuidlib.NewString(),
		UserID:           userID,
		AmountUSD:        req.AmountUSD,
		AmountNGN:        amountNGN,
		ExchangeRate:     rate.USDTNGN,
		PaymentProvider:  "paystack",
		PaymentReference: reference,
		PaymentStatus:    "pending",
		DailyRewardNGN:   settings.DailyRewardNGN,
		MaxBusinessDays:  settings.MaxBusinessDays,
		PaidBusinessDays: 0,
		TotalEarnedNGN:   0,
		TotalPendingNGN:  totalPending,
		Status:           models.InvestmentStatusPendingPayment,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := s.store.CreateInvestment(ctx, inv); err != nil {
		return nil, err
	}

	s.audit(ctx, userID, audit.ActionDeposit, "earnings_investment", inv.ID, map[string]interface{}{
		"provider": "paystack", "reference": reference, "amount_usd": req.AmountUSD,
	})

	return &models.InitEarningsPaymentResponse{
		AuthorizationURL: authURL,
		Reference:        reference,
		AccessCode:       accessCode,
	}, nil
}

func (s *Service) InitFlutterwavePayment(ctx context.Context, userID string, req *models.InitEarningsPaymentRequest) (*models.InitEarningsPaymentResponse, error) {
	req.Normalize()
	settings, rate, err := s.validateInvestmentRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	amountNGN := req.AmountUSD * rate.USDTNGN
	reference := fmt.Sprintf("EARN-FW-%s-%d", uuidlib.NewString()[:8], time.Now().Unix())

	// Create pending payment transaction
	pt := &models.EarningsPaymentTransaction{
		UserID:       userID,
		Provider:     "flutterwave",
		Reference:    reference,
		Type:         "investment",
		Status:       "pending",
		AmountNGN:    amountNGN,
		AmountUSD:    req.AmountUSD,
		ExchangeRate: rate.USDTNGN,
	}
	if err := s.store.CreatePaymentTransaction(ctx, pt); err != nil {
		return nil, err
	}

	// Call Flutterwave initialize API
	authURL, err := s.callFlutterwaveInitialize(ctx, amountNGN, req.Currency, reference, userID)
	if err != nil {
		return nil, err
	}

	// Create pending investment record
	now := time.Now().UTC()
	totalPending := float64(settings.MaxBusinessDays) * settings.DailyRewardNGN
	inv := &models.EarningsInvestment{
		ID:               uuidlib.NewString(),
		UserID:           userID,
		AmountUSD:        req.AmountUSD,
		AmountNGN:        amountNGN,
		ExchangeRate:     rate.USDTNGN,
		PaymentProvider:  "flutterwave",
		PaymentReference: reference,
		PaymentStatus:    "pending",
		DailyRewardNGN:   settings.DailyRewardNGN,
		MaxBusinessDays:  settings.MaxBusinessDays,
		PaidBusinessDays: 0,
		TotalEarnedNGN:   0,
		TotalPendingNGN:  totalPending,
		Status:           models.InvestmentStatusPendingPayment,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := s.store.CreateInvestment(ctx, inv); err != nil {
		return nil, err
	}

	s.audit(ctx, userID, audit.ActionDeposit, "earnings_investment", inv.ID, map[string]interface{}{
		"provider": "flutterwave", "reference": reference, "amount_usd": req.AmountUSD,
	})

	return &models.InitEarningsPaymentResponse{
		AuthorizationURL: authURL,
		Reference:        reference,
	}, nil
}

func (s *Service) validateInvestmentRequest(ctx context.Context, req *models.InitEarningsPaymentRequest) (*models.InvestmentSettings, *models.ExchangeRate, error) {
	// Check system is enabled
	settings, err := s.store.GetSettings(ctx)
	if err != nil {
		return nil, nil, err
	}
	if settings == nil {
		settings = defaultInvestmentSettings()
	}
	if !settings.Enabled {
		return nil, nil, errors.ErrSystemDisabled
	}

	// Validate minimum investment
	if req.AmountUSD < settings.MinimumInvestmentUSD {
		return nil, nil, errors.ErrMinimumInvestmentNotMet
	}

	// Get exchange rate
	rate, err := s.store.GetExchangeRate(ctx)
	if err != nil {
		return nil, nil, err
	}
	if rate == nil {
		rate = defaultExchangeRate()
	}

	return settings, rate, nil
}

// ─── Webhook Processing ─────────────────────────────────

func (s *Service) ProcessPaystackWebhook(ctx context.Context, payload []byte, signature string) error {
	if !s.verifyPaystackSignature(payload, signature) {
		return errors.ErrInvalidSignature
	}

	var event struct {
		Event string `json:"event"`
		Data  struct {
			Reference string  `json:"reference"`
			Status    string  `json:"status"`
			Amount    float64 `json:"amount"`
			Currency  string  `json:"currency"`
			PaidAt    string  `json:"paid_at"`
			ID        int     `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(payload, &event); err != nil {
		return apperrors.ErrBadRequest
	}

	if event.Event != "charge.success" {
		return nil
	}

	eventID := fmt.Sprintf("paystack-earnings-%d", event.Data.ID)
	reference := event.Data.Reference

	// Deduplicate by checking if already processed
	existing, err := s.store.GetPaymentTransactionByReference(ctx, "paystack", reference)
	if err != nil {
		return err
	}
	if existing != nil && existing.Status == "completed" {
		return errors.ErrPaymentAlreadyProcessed
	}

	// Verify payment with Paystack API
	verified, err := s.verifyPaystackTransaction(ctx, reference)
	if err != nil {
		return err
	}
	if !verified {
		return errors.ErrPaymentVerificationFailed
	}

	// Process the investment
	amountPaid := event.Data.Amount / 100
	if err := s.processSuccessfulPayment(ctx, "paystack", reference, amountPaid, event.Data.Currency); err != nil {
		// Log but don't return error for duplicate processing
		if err == errors.ErrPaymentAlreadyProcessed {
			return nil
		}
		return err
	}

	// Record webhook event for deduplication
	_ = s.store.CreatePaymentTransaction(ctx, &models.EarningsPaymentTransaction{
		UserID:    "",
		Provider:  "paystack",
		Reference: eventID,
		Type:      "webhook",
		Status:    "processed",
	})

	return nil
}

func (s *Service) ProcessFlutterwaveWebhook(ctx context.Context, payload []byte, signature string) error {
	if !s.verifyFlutterwaveSignature(payload, signature) {
		return errors.ErrInvalidSignature
	}

	var event struct {
		Event string `json:"event"`
		Data  struct {
			ID            int     `json:"id"`
			Reference     string  `json:"tx_ref"`
			Status        string  `json:"status"`
			Amount        float64 `json:"amount"`
			Currency      string  `json:"currency"`
			ChargedAmount float64 `json:"charged_amount"`
		} `json:"data"`
	}
	if err := json.Unmarshal(payload, &event); err != nil {
		return apperrors.ErrBadRequest
	}

	if event.Event != "charge.completed" && event.Event != "transfer.completed" {
		return nil
	}
	if event.Data.Status != "successful" {
		return nil
	}

	eventID := fmt.Sprintf("flutterwave-earnings-%d", event.Data.ID)
	reference := event.Data.Reference

	existing, err := s.store.GetPaymentTransactionByReference(ctx, "flutterwave", reference)
	if err != nil {
		return err
	}
	if existing != nil && existing.Status == "completed" {
		return errors.ErrPaymentAlreadyProcessed
	}

	verified, err := s.verifyFlutterwaveTransaction(ctx, reference)
	if err != nil {
		return err
	}
	if !verified {
		return errors.ErrPaymentVerificationFailed
	}

	amount := event.Data.Amount
	if event.Data.ChargedAmount > 0 {
		amount = event.Data.ChargedAmount
	}
	if err := s.processSuccessfulPayment(ctx, "flutterwave", reference, amount, event.Data.Currency); err != nil {
		if err == errors.ErrPaymentAlreadyProcessed {
			return nil
		}
		return err
	}

	_ = s.store.CreatePaymentTransaction(ctx, &models.EarningsPaymentTransaction{
		UserID:    "",
		Provider:  "flutterwave",
		Reference: eventID,
		Type:      "webhook",
		Status:    "processed",
	})

	return nil
}

func (s *Service) processSuccessfulPayment(ctx context.Context, provider, reference string, amountPaid float64, currency string) error {
	s.logger.Info("payment verified",
		zap.String("provider", provider),
		zap.String("reference", reference),
		zap.Float64("amount", amountPaid),
		zap.String("currency", currency),
	)

	inv, err := s.store.GetInvestmentByReference(ctx, provider, reference)
	if err != nil {
		return err
	}
	if inv == nil {
		return errors.ErrInvestmentNotFound
	}

	if inv.Status != models.InvestmentStatusPendingPayment {
		return errors.ErrPaymentAlreadyProcessed
	}

	// Verify payment amount (allow for gateway fees)
	diff := math.Abs(inv.AmountNGN - amountPaid)
	if diff > 100 {
		s.logger.Warn("payment amount mismatch",
			zap.String("reference", reference),
			zap.Float64("expected", inv.AmountNGN),
			zap.Float64("received", amountPaid),
		)
		return errors.ErrPaymentVerificationFailed
	}

	now := time.Now().UTC()

	// Calculate maturity date (20 business days from now)
	maturityDate := s.calculateMaturityDate(now, inv.MaxBusinessDays)

	// Update payment transaction
	pt, err := s.store.GetPaymentTransactionByReference(ctx, provider, reference)
	if err == nil && pt != nil {
		_ = s.store.UpdatePaymentTransaction(ctx, pt.ID, "completed", &now)
	}

	// Update investment to active
	inv.PaymentStatus = "completed"
	inv.Status = models.InvestmentStatusActive
	inv.StartedAt = &now
	inv.MaturityDate = &maturityDate
	inv.UpdatedAt = now

	if err := s.store.UpdateInvestment(ctx, inv); err != nil {
		return err
	}

	// Process referral commission for the referred user
	s.processReferralCommission(ctx, inv)

	// Create notification
	s.createNotification(ctx, inv.UserID, "payment_confirmed",
		"Payment Confirmed",
		fmt.Sprintf("Your investment of $%.2f has been confirmed and is now active.", inv.AmountUSD),
		map[string]interface{}{"investment_id": inv.ID, "amount_usd": inv.AmountUSD})

	s.audit(ctx, inv.UserID, audit.ActionDeposit, "earnings_investment", inv.ID, map[string]interface{}{
		"provider": provider, "reference": reference, "amount_usd": inv.AmountUSD,
	})

	return nil
}

// ─── Daily Reward Processing ─────────────────────────────

// ProcessDailyRewards checks all active investments and credits daily rewards for business days.
func (s *Service) ProcessDailyRewards(ctx context.Context) error {
	today := time.Now().UTC()
	if !s.isBusinessDay(today) {
		s.logger.Info("today is not a business day, skipping rewards",
			zap.String("date", today.Format("2006-01-02")),
			zap.String("weekday", today.Weekday().String()),
		)
		return nil
	}

	investments, err := s.store.ListActiveInvestments(ctx, 500)
	if err != nil {
		return err
	}

	s.logger.Info("daily reward check started",
		zap.Int("active_investments", len(investments)),
		zap.String("date", today.Format("2006-01-02")),
	)

	for _, inv := range investments {
		if err := s.processDailyReward(ctx, inv, today); err != nil {
			s.logger.Error("failed to process daily reward",
				zap.String("investment_id", inv.ID),
				zap.Error(err),
			)
			continue
		}
	}
	return nil
}

func (s *Service) processDailyReward(ctx context.Context, inv *models.EarningsInvestment, today time.Time) error {
	// Check if already rewarded today
	existing, err := s.store.GetRewardByDate(ctx, inv.ID, today)
	if err != nil {
		return err
	}
	if existing != nil {
		return nil // Already rewarded today, skip
	}

	// Check if max business days reached
	if inv.PaidBusinessDays >= inv.MaxBusinessDays {
		// Mark as completed
		now := time.Now().UTC()
		inv.Status = models.InvestmentStatusCompleted
		inv.CompletedAt = &now
		inv.UpdatedAt = now
		if err := s.store.UpdateInvestment(ctx, inv); err != nil {
			return err
		}

		s.createNotification(ctx, inv.UserID, "investment_matured",
			"Investment Matured",
			fmt.Sprintf("Your investment of $%.2f has matured. You earned a total of ₦%.2f.", inv.AmountUSD, inv.TotalEarnedNGN),
			map[string]interface{}{"investment_id": inv.ID, "total_earned": inv.TotalEarnedNGN})

		return nil
	}

	nextBusinessDay := inv.PaidBusinessDays + 1
	amount := inv.DailyRewardNGN

	// Create reward record
	reward := &models.InvestmentReward{
		InvestmentID:      inv.ID,
		UserID:            inv.UserID,
		AmountNGN:         amount,
		RewardDate:        today,
		BusinessDayNumber: nextBusinessDay,
		Status:            "credited",
	}
	if err := s.store.CreateReward(ctx, reward); err != nil {
		return err
	}

	// Update investment totals
	inv.PaidBusinessDays = nextBusinessDay
	inv.TotalEarnedNGN += amount
	inv.TotalPendingNGN = float64(inv.MaxBusinessDays-nextBusinessDay) * inv.DailyRewardNGN
	inv.UpdatedAt = time.Now().UTC()

	if err := s.store.UpdateInvestment(ctx, inv); err != nil {
		return err
	}

	s.logger.Info("daily reward credited",
		zap.String("investment_id", inv.ID),
		zap.String("user_id", inv.UserID),
		zap.Float64("amount", amount),
		zap.Int("day", nextBusinessDay),
		zap.Int("total_days", inv.MaxBusinessDays),
	)

	s.createNotification(ctx, inv.UserID, "daily_reward_credited",
		"Daily Reward Credited",
		fmt.Sprintf("₦%.2f has been credited to your investment. Day %d of %d.", amount, nextBusinessDay, inv.MaxBusinessDays),
		map[string]interface{}{"investment_id": inv.ID, "amount": amount, "day": nextBusinessDay})

	return nil
}

// ─── Withdrawal Processing ───────────────────────────────

func (s *Service) RequestWithdrawal(ctx context.Context, userID string, req *models.WithdrawalRequest) (*models.Withdrawal, error) {
	settings, err := s.store.GetSettings(ctx)
	if err != nil {
		return nil, err
	}
	if settings == nil {
		return nil, errors.ErrSettingsNotFound
	}

	var withdrawal *models.Withdrawal

	if req.InvestmentID != nil && *req.InvestmentID != "" {
		// Investment-specific withdrawal (early or normal)
		inv, err := s.store.GetInvestmentByID(ctx, *req.InvestmentID)
		if err != nil {
			return nil, err
		}
		if inv == nil || inv.UserID != userID {
			return nil, errors.ErrInvestmentNotFound
		}

		if inv.Status == models.InvestmentStatusActive {
			// Check if matured
			if inv.MaturityDate != nil && time.Now().UTC().After(*inv.MaturityDate) {
				// Normal withdrawal (capital + profits - fee)
				withdrawal, err = s.processNormalWithdrawal(ctx, inv, settings)
			} else {
				// Early withdrawal (penalty + fee)
				withdrawal, err = s.processEarlyWithdrawal(ctx, inv, settings)
			}
		} else {
			return nil, errors.ErrInvestmentNotActive
		}
	} else {
		// Earnings-only withdrawal
		withdrawal, err = s.processEarningsWithdrawal(ctx, userID, req.AmountNGN, settings)
	}

	if err != nil {
		return nil, err
	}

	if err := s.store.CreateWithdrawal(ctx, withdrawal); err != nil {
		return nil, err
	}

	return withdrawal, nil
}

func (s *Service) processEarningsWithdrawal(ctx context.Context, userID string, amount float64, settings *models.InvestmentSettings) (*models.Withdrawal, error) {
	// Check total earnings
	totalEarnings, err := s.store.SumRewardsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Check pending withdrawals
	pendingWithdrawals, err := s.store.SumPendingWithdrawalsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	availableBalance := totalEarnings - pendingWithdrawals
	if amount > availableBalance {
		return nil, errors.ErrInsufficientBalance
	}

	// Calculate fee
	fee := s.calculateWithdrawalFee(amount)
	netAmount := amount - fee

	return &models.Withdrawal{
		UserID:         userID,
		AmountNGN:      amount,
		FeeNGN:         fee,
		PenaltyNGN:     0,
		NetAmountNGN:   netAmount,
		WithdrawalType: models.WithdrawalTypeEarnings,
		Status:         models.WithdrawalStatusPendingReview,
	}, nil
}

func (s *Service) processEarlyWithdrawal(ctx context.Context, inv *models.EarningsInvestment, settings *models.InvestmentSettings) (*models.Withdrawal, error) {
	// Calculate penalty and fee
	penaltyPercent := settings.EarlyWithdrawalPenaltyPercent / 100.0
	feePercent := settings.EarlyWithdrawalFeePercent / 100.0

	capital := inv.AmountNGN
	earned := inv.TotalEarnedNGN
	totalAmount := capital + earned

	penalty := totalAmount * penaltyPercent
	fee := totalAmount * feePercent
	netAmount := totalAmount - penalty - fee

	return &models.Withdrawal{
		UserID:         inv.UserID,
		InvestmentID:   &inv.ID,
		AmountNGN:      totalAmount,
		FeeNGN:         fee,
		PenaltyNGN:     penalty,
		NetAmountNGN:   netAmount,
		WithdrawalType: models.WithdrawalTypeEarly,
		Status:         models.WithdrawalStatusPendingReview,
	}, nil
}

func (s *Service) processNormalWithdrawal(ctx context.Context, inv *models.EarningsInvestment, settings *models.InvestmentSettings) (*models.Withdrawal, error) {
	capital := inv.AmountNGN
	earned := inv.TotalEarnedNGN
	totalAmount := capital + earned

	// Calculate fee
	fee := s.calculateWithdrawalFee(totalAmount)
	netAmount := totalAmount - fee

	return &models.Withdrawal{
		UserID:         inv.UserID,
		InvestmentID:   &inv.ID,
		AmountNGN:      totalAmount,
		FeeNGN:         fee,
		PenaltyNGN:     0,
		NetAmountNGN:   netAmount,
		WithdrawalType: models.WithdrawalTypeNormal,
		Status:         models.WithdrawalStatusPendingReview,
	}, nil
}

func (s *Service) calculateWithdrawalFee(amount float64) float64 {
	tier, err := s.store.GetFeeTierForAmount(context.Background(), amount)
	if err != nil || tier == nil {
		// Default fee if no tier found
		return amount * 0.03
	}
	return amount * (tier.FeePercent / 100.0)
}

// ─── Admin Withdrawal Actions ────────────────────────────

func (s *Service) AdminProcessWithdrawal(ctx context.Context, withdrawalID string, req *models.AdminWithdrawalActionRequest, adminID string) (*models.Withdrawal, error) {
	w, err := s.store.GetWithdrawalByID(ctx, withdrawalID)
	if err != nil {
		return nil, err
	}
	if w == nil {
		return nil, errors.ErrWithdrawalNotFound
	}
	if w.Status != models.WithdrawalStatusPendingReview {
		return nil, errors.ErrWithdrawalAlreadyProcessed
	}

	now := time.Now().UTC()

	switch req.Action {
	case "approve":
		w.Status = models.WithdrawalStatusApproved
		w.ReviewedBy = &adminID
		w.ReviewedAt = &now
		w.ProcessedAt = &now
		w.CompletedAt = &now

		// If early withdrawal, update investment status
		if w.InvestmentID != nil && w.WithdrawalType == models.WithdrawalTypeEarly {
			inv, err := s.store.GetInvestmentByID(ctx, *w.InvestmentID)
			if err == nil && inv != nil {
				inv.Status = models.InvestmentStatusEarlyWithdrawal
				inv.EarlyWithdrawalAt = &now
				inv.UpdatedAt = now
				_ = s.store.UpdateInvestment(ctx, inv)
			}
		}

		// If normal withdrawal, complete the investment
		if w.InvestmentID != nil && w.WithdrawalType == models.WithdrawalTypeNormal {
			inv, err := s.store.GetInvestmentByID(ctx, *w.InvestmentID)
			if err == nil && inv != nil {
				inv.Status = models.InvestmentStatusCompleted
				inv.CompletedAt = &now
				inv.UpdatedAt = now
				_ = s.store.UpdateInvestment(ctx, inv)
			}
		}

		s.createNotification(ctx, w.UserID, "withdrawal_approved",
			"Withdrawal Approved",
			fmt.Sprintf("Your withdrawal of ₦%.2f has been approved and will be processed shortly.", w.NetAmountNGN),
			map[string]interface{}{"withdrawal_id": w.ID, "amount": w.NetAmountNGN})

	case "reject":
		w.Status = models.WithdrawalStatusRejected
		w.ReviewedBy = &adminID
		w.ReviewedAt = &now
		if req.RejectionReason != nil {
			w.RejectionReason = req.RejectionReason
		}

		s.createNotification(ctx, w.UserID, "withdrawal_rejected",
			"Withdrawal Rejected",
			fmt.Sprintf("Your withdrawal request of ₦%.2f has been rejected. Reason: %s", w.AmountNGN, *w.RejectionReason),
			map[string]interface{}{"withdrawal_id": w.ID, "reason": *w.RejectionReason})
	}

	if err := s.store.UpdateWithdrawal(ctx, w); err != nil {
		return nil, err
	}

	return w, nil
}

// ─── Referral Commission ─────────────────────────────────

func (s *Service) processReferralCommission(ctx context.Context, inv *models.EarningsInvestment) {
	// Get the user's referral code to find who referred them
	// This relies on the existing identity service's referral system
	// For now, we check if the user was referred by someone
	// The referral code is stored in the identity_users table

	// We'll use the event bus to notify the referral system
	if s.eventBus != nil {
		s.publish("referral.investment_activated", map[string]interface{}{
			"user_id":       inv.UserID,
			"investment_id": inv.ID,
			"amount_usd":    inv.AmountUSD,
		})
	}
}

func (s *Service) ProcessReferralCommission(ctx context.Context, referrerID, referredID, investmentID string, amountUSD float64, percent float64) error {
	// Check duplicate
	existing, err := s.store.GetReferralCommission(ctx, investmentID, referrerID)
	if err != nil {
		return err
	}
	if existing != nil {
		return errors.ErrDuplicateReferralCommission
	}

	// Get exchange rate
	rate, err := s.store.GetExchangeRate(ctx)
	if err != nil || rate == nil {
		return errors.ErrExchangeRateNotFound
	}

	commissionUSD := amountUSD * (percent / 100.0)
	commissionNGN := commissionUSD * rate.USDTNGN

	rc := &models.ReferralCommission{
		ReferrerID:   referrerID,
		ReferredID:   referredID,
		InvestmentID: investmentID,
		AmountUSD:    commissionUSD,
		AmountNGN:    commissionNGN,
		Percent:      percent,
		Status:       "paid",
		PaidAt:       nil,
	}

	if err := s.store.CreateReferralCommission(ctx, rc); err != nil {
		return err
	}

	s.createNotification(ctx, referrerID, "referral_commission_received",
		"Referral Commission Received",
		fmt.Sprintf("You earned ₦%.2f from a referral investment.", commissionNGN),
		map[string]interface{}{"commission_id": rc.ID, "amount": commissionNGN})

	return nil
}

// ─── Investments ─────────────────────────────────────────

func (s *Service) ListInvestments(ctx context.Context, userID, status string, page, perPage int) ([]*models.EarningsSummary, int, error) {
	investments, total, err := s.store.ListUserInvestments(ctx, userID, status, page, perPage)
	if err != nil {
		return nil, 0, err
	}
	summaries := make([]*models.EarningsSummary, 0, len(investments))
	for _, inv := range investments {
		summaries = append(summaries, toEarningsSummary(inv))
	}
	return summaries, total, nil
}

func (s *Service) GetInvestment(ctx context.Context, userID, id string) (*models.EarningsInvestment, error) {
	inv, err := s.store.GetInvestmentByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if inv == nil || inv.UserID != userID {
		return nil, errors.ErrInvestmentNotFound
	}
	return inv, nil
}

func toEarningsSummary(inv *models.EarningsInvestment) *models.EarningsSummary {
	summary := &models.EarningsSummary{
		ID:               inv.ID,
		AmountUSD:        inv.AmountUSD,
		AmountNGN:        inv.AmountNGN,
		ExchangeRate:     inv.ExchangeRate,
		DailyRewardNGN:   inv.DailyRewardNGN,
		PaidBusinessDays: inv.PaidBusinessDays,
		MaxBusinessDays:  inv.MaxBusinessDays,
		TotalEarnedNGN:   inv.TotalEarnedNGN,
		TotalPendingNGN:  inv.TotalPendingNGN,
		Status:           inv.Status,
		MaturityDate:     inv.MaturityDate,
		StartedAt:        inv.StartedAt,
		CreatedAt:        inv.CreatedAt,
	}
	if inv.MaturityDate != nil && inv.Status == models.InvestmentStatusActive {
		remaining := time.Until(*inv.MaturityDate)
		if remaining > 0 {
			summary.RemainingDays = int(remaining.Hours() / 24)
		}
	}
	if inv.MaxBusinessDays > 0 {
		summary.ProgressPct = math.Min((float64(inv.PaidBusinessDays)/float64(inv.MaxBusinessDays))*100, 100)
	}
	return summary
}

// ─── Dashboard ───────────────────────────────────────────

func (s *Service) GetDashboard(ctx context.Context, userID string) (*models.EarningsDashboard, error) {
	investments, _, err := s.store.ListUserInvestments(ctx, userID, "", 1, 500)
	if err != nil {
		return nil, err
	}

	rate, err := s.store.GetExchangeRate(ctx)
	if err != nil {
		rate = nil
	}

	exchangeRate := 1400.0
	if rate != nil {
		exchangeRate = rate.USDTNGN
	}

	todayEarnings, _ := s.store.SumTodayRewardsByUser(ctx, userID)
	monthlyEarnings, _ := s.store.SumMonthlyRewardsByUser(ctx, userID)
	totalEarnings, _ := s.store.SumRewardsByUser(ctx, userID)
	pendingWithdrawals, _ := s.store.SumPendingWithdrawalsByUser(ctx, userID)
	referralEarnings, _ := s.store.SumReferralCommissionsByReferrer(ctx, userID, "paid")

	dash := &models.EarningsDashboard{
		TotalInvestedUSD:     0,
		TotalInvestedNGN:     0,
		ExchangeRate:         exchangeRate,
		TodayEarningsNGN:     todayEarnings,
		MonthlyEarningsNGN:   monthlyEarnings,
		AvailableBalanceNGN:  totalEarnings - pendingWithdrawals,
		PendingWithdrawalNGN: pendingWithdrawals,
		ReferralEarningsNGN:  referralEarnings,
	}

	var summaries []*models.EarningsSummary

	for _, inv := range investments {
		dash.TotalInvestedUSD += inv.AmountUSD
		dash.TotalInvestedNGN += inv.AmountNGN

		switch inv.Status {
		case models.InvestmentStatusActive:
			dash.ActiveInvestments++
		case models.InvestmentStatusCompleted:
			dash.CompletedInvestments++
		}

		summary := &models.EarningsSummary{
			ID:               inv.ID,
			AmountUSD:        inv.AmountUSD,
			AmountNGN:        inv.AmountNGN,
			ExchangeRate:     inv.ExchangeRate,
			DailyRewardNGN:   inv.DailyRewardNGN,
			PaidBusinessDays: inv.PaidBusinessDays,
			MaxBusinessDays:  inv.MaxBusinessDays,
			TotalEarnedNGN:   inv.TotalEarnedNGN,
			TotalPendingNGN:  inv.TotalPendingNGN,
			Status:           inv.Status,
			MaturityDate:     inv.MaturityDate,
			StartedAt:        inv.StartedAt,
			CreatedAt:        inv.CreatedAt,
		}

		if inv.MaturityDate != nil && inv.Status == models.InvestmentStatusActive {
			remaining := time.Until(*inv.MaturityDate)
			if remaining > 0 {
				summary.RemainingDays = int(remaining.Hours() / 24)
			}
			if inv.MaxBusinessDays > 0 {
				summary.ProgressPct = math.Min((float64(inv.PaidBusinessDays)/float64(inv.MaxBusinessDays))*100, 100)
			}
		}

		if inv.Status == models.InvestmentStatusCompleted {
			summary.ProgressPct = 100
		}

		summaries = append(summaries, summary)
	}

	dash.Investments = summaries

	// Get referral info
	referralInfo := &models.ReferralInfo{
		ReferralCode:           "",
		ReferralLink:           "",
		TotalReferrals:         0,
		ActiveReferrals:        0,
		ReferralEarningsNGN:    referralEarnings,
		WithdrawableBalanceNGN: referralEarnings,
		MinimumTarget:          5,
	}

	totalRefs, _ := s.store.CountReferralsByReferrer(ctx, userID)
	activeRefs, _ := s.store.CountActiveReferralsByReferrer(ctx, userID)
	referralInfo.TotalReferrals = totalRefs
	referralInfo.ActiveReferrals = activeRefs

	dash.ReferralInfo = referralInfo

	return dash, nil
}

func (s *Service) GetInvestmentRewards(ctx context.Context, userID, investmentID string, page, perPage int) ([]*models.RewardHistoryItem, int, error) {
	if investmentID != "" {
		// Verify ownership
		inv, err := s.store.GetInvestmentByID(ctx, investmentID)
		if err != nil {
			return nil, 0, err
		}
		if inv == nil || inv.UserID != userID {
			return nil, 0, errors.ErrInvestmentNotFound
		}
		rewards, err := s.store.ListRewardsByInvestment(ctx, investmentID)
		if err != nil {
			return nil, 0, err
		}
		items := make([]*models.RewardHistoryItem, len(rewards))
		for i, r := range rewards {
			items[i] = &models.RewardHistoryItem{
				ID:                r.ID,
				AmountNGN:         r.AmountNGN,
				RewardDate:        r.RewardDate.Format("2006-01-02"),
				BusinessDayNumber: r.BusinessDayNumber,
				Status:            r.Status,
				CreatedAt:         r.CreatedAt,
			}
		}
		return items, len(items), nil
	}

	rewards, total, err := s.store.ListRewardsByUser(ctx, userID, page, perPage)
	if err != nil {
		return nil, 0, err
	}
	items := make([]*models.RewardHistoryItem, len(rewards))
	for i, r := range rewards {
		items[i] = &models.RewardHistoryItem{
			ID:                r.ID,
			AmountNGN:         r.AmountNGN,
			RewardDate:        r.RewardDate.Format("2006-01-02"),
			BusinessDayNumber: r.BusinessDayNumber,
			Status:            r.Status,
			CreatedAt:         r.CreatedAt,
		}
	}
	return items, total, nil
}

func (s *Service) GetPaymentHistory(ctx context.Context, userID string, page, perPage int) ([]*models.PaymentHistoryItem, int, error) {
	transactions, total, err := s.store.ListUserPaymentTransactions(ctx, userID, page, perPage)
	if err != nil {
		return nil, 0, err
	}
	items := make([]*models.PaymentHistoryItem, len(transactions))
	for i, t := range transactions {
		items[i] = &models.PaymentHistoryItem{
			ID:        t.ID,
			AmountNGN: t.AmountNGN,
			AmountUSD: t.AmountUSD,
			Provider:  t.Provider,
			Reference: t.Reference,
			Status:    t.Status,
			PaidAt:    t.PaidAt,
			CreatedAt: t.CreatedAt,
		}
	}
	return items, total, nil
}

func (s *Service) GetWithdrawalHistory(ctx context.Context, userID string, page, perPage int) ([]*models.WithdrawalHistoryItem, int, error) {
	withdrawals, total, err := s.store.ListUserWithdrawals(ctx, userID, page, perPage)
	if err != nil {
		return nil, 0, err
	}
	items := make([]*models.WithdrawalHistoryItem, len(withdrawals))
	for i, w := range withdrawals {
		items[i] = &models.WithdrawalHistoryItem{
			ID:              w.ID,
			AmountNGN:       w.AmountNGN,
			FeeNGN:          w.FeeNGN,
			PenaltyNGN:      w.PenaltyNGN,
			NetAmountNGN:    w.NetAmountNGN,
			WithdrawalType:  w.WithdrawalType,
			Status:          w.Status,
			RejectionReason: w.RejectionReason,
			CreatedAt:       w.CreatedAt,
			CompletedAt:     w.CompletedAt,
		}
	}
	return items, total, nil
}

func (s *Service) GetNotifications(ctx context.Context, userID string, page, perPage int) ([]*models.InvestmentNotification, int, error) {
	return s.store.ListUserNotifications(ctx, userID, page, perPage)
}

func (s *Service) MarkNotificationRead(ctx context.Context, userID, notificationID string) error {
	return s.store.MarkNotificationRead(ctx, notificationID)
}

func (s *Service) MarkAllNotificationsRead(ctx context.Context, userID string) error {
	return s.store.MarkAllNotificationsRead(ctx, userID)
}

func (s *Service) GetUnreadNotificationCount(ctx context.Context, userID string) (int, error) {
	return s.store.CountUnreadNotifications(ctx, userID)
}

// ─── Admin Dashboard ─────────────────────────────────────

func (s *Service) AdminGetDashboard(ctx context.Context) (*models.AdminEarningsDashboard, error) {
	totalInvested, _ := s.store.SumInvestmentsByStatus(ctx, "")
	totalPaid, _ := s.store.SumAllPaidOut(ctx)
	activeInvestors, _ := s.store.CountActiveInvestors(ctx)
	totalInvestors, _ := s.store.CountTotalInvestors(ctx)
	pendingWithdrawals, _ := s.store.CountInvestmentsByStatus(ctx, "pending_payment")
	todayPayout, _ := s.store.SumTodayRewardsAll(ctx)
	referralPaid, _ := s.store.SumReferralCommissionsByReferrer(ctx, "", "paid")

	// Count pending withdrawals
	var pendingCount int
	withdrawals, _, _ := s.store.ListWithdrawalsByStatus(ctx, "pending_review", 1, 1)
	if withdrawals != nil {
		pendingCount = len(withdrawals)
	}

	return &models.AdminEarningsDashboard{
		TotalInvestedNGN:     totalInvested,
		TotalPaidOutNGN:      totalPaid,
		ActiveInvestors:      activeInvestors,
		TotalInvestors:       totalInvestors,
		PendingWithdrawals:   pendingCount,
		PendingPayments:      pendingWithdrawals,
		TodayPayoutNGN:       todayPayout,
		TotalReferralPaidNGN: referralPaid,
	}, nil
}

func (s *Service) AdminListWithdrawals(ctx context.Context, status string, page, perPage int) ([]*models.Withdrawal, int, error) {
	return s.store.ListWithdrawalsByStatus(ctx, status, page, perPage)
}

func (s *Service) AdminListInvestments(ctx context.Context, status string, page, perPage int) ([]*models.EarningsInvestment, int, error) {
	return s.store.ListUserInvestments(ctx, "", status, page, perPage)
}

// ─── Helpers ─────────────────────────────────────────────

func (s *Service) isBusinessDay(t time.Time) bool {
	weekday := t.Weekday()
	return weekday != time.Saturday && weekday != time.Sunday
}

func (s *Service) calculateMaturityDate(start time.Time, businessDays int) time.Time {
	current := start
	count := 0
	for count < businessDays {
		current = current.AddDate(0, 0, 1)
		if s.isBusinessDay(current) {
			count++
		}
	}
	return current
}

func (s *Service) createNotification(ctx context.Context, userID, notifType, title, message string, data map[string]interface{}) {
	n := &models.InvestmentNotification{
		UserID:  userID,
		Type:    notifType,
		Title:   title,
		Message: message,
		Data:    data,
	}
	if err := s.store.CreateNotification(ctx, n); err != nil {
		s.logger.Error("failed to create notification",
			zap.String("user_id", userID),
			zap.String("type", notifType),
			zap.Error(err),
		)
	}
}

// ─── Payment Gateway API Calls ───────────────────────────

func (s *Service) callPaystackInitialize(ctx context.Context, amount float64, currency, reference, userID string) (string, string, error) {
	if s.cfg.PaystackSecretKey == "" {
		return "", "", errors.ErrGatewayNotConfigured
	}

	amountInKobo := int64(amount * 100)
	payload := map[string]interface{}{
		"amount":       amountInKobo,
		"currency":     currency,
		"reference":    reference,
		"callback_url": fmt.Sprintf("%s/investments/verify", s.cfg.BaseURL),
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST", "https://api.paystack.co/transaction/initialize", strings.NewReader(string(body)))
	req.Header.Set("Authorization", "Bearer "+s.cfg.PaystackSecretKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("paystack init failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
		Data    struct {
			AuthorizationURL string `json:"authorization_url"`
			AccessCode       string `json:"access_code"`
			Reference        string `json:"reference"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", fmt.Errorf("paystack init decode failed: %w", err)
	}
	if !result.Status {
		return "", "", fmt.Errorf("paystack init failed: %s", result.Message)
	}

	return result.Data.AuthorizationURL, result.Data.AccessCode, nil
}

func (s *Service) verifyPaystackTransaction(ctx context.Context, reference string) (bool, error) {
	if s.cfg.PaystackSecretKey == "" {
		return false, errors.ErrGatewayNotConfigured
	}

	req, _ := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.paystack.co/transaction/verify/%s", reference), nil)
	req.Header.Set("Authorization", "Bearer "+s.cfg.PaystackSecretKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("paystack verify failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
		Data    struct {
			Status string `json:"status"`
			Amount int    `json:"amount"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("paystack verify decode failed: %w", err)
	}

	return result.Status && result.Data.Status == "success", nil
}

func (s *Service) verifyPaystackSignature(payload []byte, signature string) bool {
	if s.cfg.PaystackSecretKey == "" {
		return false
	}
	h := sha512.New()
	h.Write(payload)
	expected := hex.EncodeToString(h.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

func (s *Service) callFlutterwaveInitialize(ctx context.Context, amount float64, currency, reference, userID string) (string, error) {
	if s.cfg.FlutterwaveSecretKey == "" {
		return "", errors.ErrGatewayNotConfigured
	}

	payload := map[string]interface{}{
		"tx_ref":       reference,
		"amount":       amount,
		"currency":     currency,
		"redirect_url": fmt.Sprintf("%s/investments/verify", s.cfg.BaseURL),
		"customer": map[string]string{
			"email": "",
		},
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST", "https://api.flutterwave.com/v3/payments", strings.NewReader(string(body)))
	req.Header.Set("Authorization", "Bearer "+s.cfg.FlutterwaveSecretKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("flutterwave init failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Data    struct {
			Link string `json:"link"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("flutterwave init decode failed: %w", err)
	}
	if result.Status != "success" {
		return "", fmt.Errorf("flutterwave init failed: %s", result.Message)
	}

	return result.Data.Link, nil
}

func (s *Service) verifyFlutterwaveTransaction(ctx context.Context, reference string) (bool, error) {
	if s.cfg.FlutterwaveSecretKey == "" {
		return false, errors.ErrGatewayNotConfigured
	}

	req, _ := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.flutterwave.com/v3/transactions/verify_by_reference?tx_ref=%s", reference), nil)
	req.Header.Set("Authorization", "Bearer "+s.cfg.FlutterwaveSecretKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("flutterwave verify failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Data    struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("flutterwave verify decode failed: %w", err)
	}

	return result.Status == "success" && result.Data.Status == "successful", nil
}

func (s *Service) verifyFlutterwaveSignature(payload []byte, signature string) bool {
	if s.cfg.FlutterwaveSecretHash == "" {
		return false
	}
	h := hmac.New(sha256.New, []byte(s.cfg.FlutterwaveSecretHash))
	h.Write(payload)
	expected := hex.EncodeToString(h.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

// ─── Workers ─────────────────────────────────────────────

func (s *Service) registerWorkers() {
	if s.jobRegistry == nil {
		return
	}
	s.jobRegistry.Register("earnings.daily_rewards", func(ctx context.Context, job workers.Job) error {
		return s.ProcessDailyRewards(ctx)
	})
}

// ─── Helpers ─────────────────────────────────────────────

func (s *Service) publish(eventType string, data map[string]interface{}) {
	if s.eventBus == nil {
		return
	}
	_ = s.eventBus.Publish(context.Background(), events.NewEvent(eventType, "earnings-service", data))
}

func (s *Service) audit(ctx context.Context, actorID string, action audit.Action, entityType string, entityID string, meta map[string]interface{}) {
	if s.auditLogger == nil {
		return
	}
	ev := audit.NewEvent(actorID, action).
		WithUserID(actorID).
		WithEntity(audit.EntityType(entityType), entityID).
		WithOutcome("success")
	if meta != nil {
		ev = ev.WithMetadata(meta)
	}
	_ = s.auditLogger.Record(ctx, ev.Build())
}
