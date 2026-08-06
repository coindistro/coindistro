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

func (s *Service) hasStore() bool {
	return s != nil && s.store != nil
}

// Config holds the earnings service configuration.
// Paystack keys must be supplied from PAYSTACK_* environment variables so
// merchant accounts can be swapped without code changes.
type Config struct {
	BaseURL               string
	PaystackSecretKey     string
	PaystackPublicKey     string
	PaystackCallbackURL   string
	PaystackWebhookSecret string
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
		MinimumInvestmentUSD:          10,
		DailyRewardNGN:                126,
		MaxBusinessDays:               20,
		ROIPercent:                    18,
		ReferralPercent:               10,
		MinReferralsForPayout:         5,
		EarlyWithdrawalPenaltyPercent: 15,
		EarlyWithdrawalFeePercent:     5,
		WithdrawalProcessingHours:     24,
		WithdrawalIntervalDays:        7,
		Enabled:                       true,
	}
}

func defaultExchangeRate() *models.ExchangeRate {
	return &models.ExchangeRate{USDTNGN: 1400}
}

func (s *Service) GetSettings(ctx context.Context) (*models.InvestmentSettings, error) {
	if !s.hasStore() {
		return defaultInvestmentSettings(), nil
	}
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
	if req.WithdrawalIntervalDays != nil {
		settings.WithdrawalIntervalDays = *req.WithdrawalIntervalDays
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
	if !s.hasStore() {
		return defaultExchangeRate(), nil
	}
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
	if s.cfg.PaystackSecretKey == "" {
		s.logger.Error("Initializing Paystack transaction failed: gateway not configured")
		return nil, errors.ErrGatewayNotConfigured
	}
	if !s.hasStore() {
		s.logger.Error("Initializing Paystack transaction failed: database store unavailable")
		return nil, errors.ErrGatewayNotConfigured
	}
	settings, rate, err := s.validateInvestmentRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	amountNGN := req.AmountUSD * rate.USDTNGN
	reference := fmt.Sprintf("EARN-PS-%s-%d", uuidlib.NewString()[:8], time.Now().Unix())

	s.logger.Info("Initializing Paystack transaction",
		zap.String("user_id", userID),
		zap.String("reference", reference),
		zap.Float64("amount_usd", req.AmountUSD),
		zap.Float64("amount_ngn", amountNGN),
		zap.String("currency", req.Currency),
	)

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

	// Load the authenticated user's email from the identity store.
	// Never trust a client-supplied email — the backend is the source of truth.
	email, err := s.store.GetUserEmail(ctx, userID)
	if err != nil {
		return nil, err
	}
	if email == "" {
		return nil, errors.ErrEmailRequired
	}

	// Call Paystack initialize API (secret key from PAYSTACK_SECRET_KEY only).
	// CRITICAL: Do NOT create investment or credit wallet here — only after
	// webhook/verify reports Paystack status == "success".
	_ = settings // validated above; amounts stored on payment transaction
	authURL, accessCode, err := s.callPaystackInitialize(ctx, amountNGN, req.Currency, reference, userID, email)
	if err != nil {
		s.logger.Error("Paystack initialize API failed",
			zap.String("reference", reference),
			zap.Error(err),
		)
		return nil, err
	}

	s.audit(ctx, userID, audit.ActionDeposit, "payment_transaction", reference, map[string]interface{}{
		"provider": "paystack", "reference": reference, "amount_usd": req.AmountUSD,
		"stage": "initialized", "wallet_credited": false, "investment_created": false,
	})

	s.logger.Info("Paystack checkout ready (no investment until payment success)",
		zap.String("reference", reference),
		zap.String("customer_email", email),
		zap.Float64("amount_ngn", amountNGN),
		zap.String("callback_url", s.paystackCallbackURL()),
	)

	return &models.InitEarningsPaymentResponse{
		AuthorizationURL: authURL,
		Reference:        reference,
		AccessCode:       accessCode,
	}, nil
}

// VerifyPaystackPayment confirms a checkout after the user returns from Paystack
// (or if the webhook is delayed). It re-verifies with Paystack's API — never trusts the client.
// Investment + wallet credit happen only after verified success.
func (s *Service) VerifyPaystackPayment(ctx context.Context, userID, reference string) (*models.EarningsInvestment, error) {
	reference = strings.TrimSpace(reference)
	if reference == "" {
		return nil, apperrors.ErrBadRequest
	}
	if !s.hasStore() {
		return nil, errors.ErrGatewayNotConfigured
	}
	if s.cfg.PaystackSecretKey == "" {
		return nil, errors.ErrGatewayNotConfigured
	}

	// Already activated?
	if inv, err := s.store.GetInvestmentByReference(ctx, "paystack", reference); err == nil && inv != nil {
		if inv.UserID != userID {
			return nil, errors.ErrInvestmentNotFound
		}
		if inv.Status == models.InvestmentStatusActive || inv.PaymentStatus == "completed" {
			s.logger.Info("Paystack payment already verified",
				zap.String("reference", reference),
				zap.String("investment_id", inv.ID),
			)
			return inv, nil
		}
	}

	// Must have a pending payment intent from init (no investment yet is OK).
	pt, err := s.store.GetPaymentTransactionByReference(ctx, "paystack", reference)
	if err != nil {
		return nil, err
	}
	if pt == nil || pt.UserID != userID {
		return nil, errors.ErrPaymentVerificationFailed
	}

	verified, err := s.verifyPaystackTransaction(ctx, reference)
	if err != nil {
		return nil, err
	}
	if !verified {
		s.logger.Info("Paystack verification not successful — no investment/wallet credit",
			zap.String("reference", reference),
			zap.String("user_id", userID),
		)
		return nil, errors.ErrPaymentVerificationFailed
	}

	s.logger.Info("Verification succeeded",
		zap.String("reference", reference),
		zap.String("user_id", userID),
	)

	if err := s.processSuccessfulPayment(ctx, "paystack", reference, pt.AmountNGN, "NGN"); err != nil {
		if err == errors.ErrPaymentAlreadyProcessed {
			return s.store.GetInvestmentByReference(ctx, "paystack", reference)
		}
		return nil, err
	}

	return s.store.GetInvestmentByReference(ctx, "paystack", reference)
}

func (s *Service) InitFlutterwavePayment(ctx context.Context, userID string, req *models.InitEarningsPaymentRequest) (*models.InitEarningsPaymentResponse, error) {
	req.Normalize()
	if !s.hasStore() {
		reference := fmt.Sprintf("EARN-FW-%s-%d", uuidlib.NewString()[:8], time.Now().Unix())
		return &models.InitEarningsPaymentResponse{
			AuthorizationURL: fmt.Sprintf("%s/checkout/flutterwave/%s", strings.TrimRight(s.cfg.BaseURL, "/"), reference),
			Reference:        reference,
		}, nil
	}
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

	// Load the authenticated user's email from the identity store.
	// Never trust a client-supplied email — the backend is the source of truth.
	email, err := s.store.GetUserEmail(ctx, userID)
	if err != nil {
		return nil, err
	}
	if email == "" {
		return nil, errors.ErrEmailRequired
	}

	// Call Flutterwave initialize API — no investment/wallet credit until success.
	_ = settings
	authURL, err := s.callFlutterwaveInitialize(ctx, amountNGN, req.Currency, reference, userID)
	if err != nil {
		return nil, err
	}

	s.audit(ctx, userID, audit.ActionDeposit, "payment_transaction", reference, map[string]interface{}{
		"provider": "flutterwave", "reference": reference, "amount_usd": req.AmountUSD,
		"stage": "initialized", "wallet_credited": false, "investment_created": false,
	})

	return &models.InitEarningsPaymentResponse{
		AuthorizationURL: authURL,
		Reference:        reference,
	}, nil
}

func (s *Service) validateInvestmentRequest(ctx context.Context, req *models.InitEarningsPaymentRequest) (*models.InvestmentSettings, *models.ExchangeRate, error) {
	if !s.hasStore() {
		return defaultInvestmentSettings(), defaultExchangeRate(), nil
	}

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
	s.logger.Info("Webhook received", zap.String("provider", "paystack"))

	if !s.verifyPaystackSignature(payload, signature) {
		s.logger.Warn("Paystack webhook signature verification failed")
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
		s.logger.Info("Paystack webhook ignored (non-success event)",
			zap.String("event", event.Event),
		)
		return nil
	}

	eventID := fmt.Sprintf("paystack-earnings-%d", event.Data.ID)
	reference := event.Data.Reference

	s.logger.Info("Processing Paystack charge.success",
		zap.String("reference", reference),
		zap.String("event_id", eventID),
	)

	// Deduplicate by checking if already processed
	existing, err := s.store.GetPaymentTransactionByReference(ctx, "paystack", reference)
	if err != nil {
		return err
	}
	if existing != nil && existing.Status == "completed" {
		s.logger.Info("Duplicate webhook ignored (already completed)",
			zap.String("reference", reference),
		)
		return nil
	}

	// Verify payment with Paystack API (never trust webhook body alone)
	verified, err := s.verifyPaystackTransaction(ctx, reference)
	if err != nil {
		return err
	}
	if !verified {
		return errors.ErrPaymentVerificationFailed
	}
	s.logger.Info("Verification succeeded", zap.String("reference", reference))

	// Process the investment (amount from Paystack is in kobo)
	amountPaid := event.Data.Amount / 100
	if err := s.processSuccessfulPayment(ctx, "paystack", reference, amountPaid, event.Data.Currency); err != nil {
		if err == errors.ErrPaymentAlreadyProcessed {
			s.logger.Info("Duplicate webhook ignored (investment already active)",
				zap.String("reference", reference),
			)
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

	s.logger.Info("Investment activated",
		zap.String("reference", reference),
		zap.String("provider", "paystack"),
	)

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

// processSuccessfulPayment runs ONLY after Paystack/Flutterwave verification succeeds.
// It creates the investment (if not already present), activates it, credits wallet, and audits.
// Cancelled/abandoned/failed payments never reach this path.
func (s *Service) processSuccessfulPayment(ctx context.Context, provider, reference string, amountPaid float64, currency string) error {
	s.logger.Info("payment verified — creating investment and crediting wallet",
		zap.String("provider", provider),
		zap.String("reference", reference),
		zap.Float64("amount", amountPaid),
		zap.String("currency", currency),
	)

	pt, err := s.store.GetPaymentTransactionByReference(ctx, provider, reference)
	if err != nil {
		return err
	}
	if pt == nil {
		return errors.ErrPaymentVerificationFailed
	}
	if pt.Status == "completed" {
		// Idempotent: already processed
		if inv, _ := s.store.GetInvestmentByReference(ctx, provider, reference); inv != nil {
			return errors.ErrPaymentAlreadyProcessed
		}
	}

	// Amount check against payment intent
	diff := math.Abs(pt.AmountNGN - amountPaid)
	if amountPaid > 0 && diff > 100 {
		s.logger.Warn("payment amount mismatch",
			zap.String("reference", reference),
			zap.Float64("expected", pt.AmountNGN),
			zap.Float64("received", amountPaid),
		)
		return errors.ErrPaymentVerificationFailed
	}

	now := time.Now().UTC()
	settings, _ := s.store.GetSettings(ctx)
	if settings == nil {
		settings = defaultInvestmentSettings()
	}

	inv, err := s.store.GetInvestmentByReference(ctx, provider, reference)
	if err != nil {
		return err
	}

	// Create investment only after verified success (deferred from init).
	if inv == nil {
		totalPending := float64(settings.MaxBusinessDays) * settings.DailyRewardNGN
		// Prefer ROI-derived daily reward when ROI is set
		daily := settings.DailyRewardNGN
		if settings.ROIPercent > 0 && settings.MaxBusinessDays > 0 {
			daily = (pt.AmountNGN * (settings.ROIPercent / 100.0)) / float64(settings.MaxBusinessDays)
			totalPending = daily * float64(settings.MaxBusinessDays)
		}
		maturityDate := s.calculateMaturityDate(now, settings.MaxBusinessDays)
		inv = &models.EarningsInvestment{
			ID:               uuidlib.NewString(),
			UserID:           pt.UserID,
			AmountUSD:        pt.AmountUSD,
			AmountNGN:        pt.AmountNGN,
			ExchangeRate:     pt.ExchangeRate,
			PaymentProvider:  provider,
			PaymentReference: reference,
			PaymentStatus:    "completed",
			DailyRewardNGN:   daily,
			MaxBusinessDays:  settings.MaxBusinessDays,
			PaidBusinessDays: 0,
			TotalEarnedNGN:   0,
			TotalPendingNGN:  totalPending,
			Status:           models.InvestmentStatusActive,
			IsDemo:           false,
			MaturityDate:     &maturityDate,
			StartedAt:        &now,
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		if err := s.store.CreateInvestment(ctx, inv); err != nil {
			return err
		}
		s.logger.Info("Investment created after verified payment",
			zap.String("investment_id", inv.ID),
			zap.String("reference", reference),
		)
	} else {
		if inv.Status != models.InvestmentStatusPendingPayment && inv.PaymentStatus == "completed" {
			return errors.ErrPaymentAlreadyProcessed
		}
		if inv.IsDemo {
			// Never promote demo via payment path
			return errors.ErrPaymentAlreadyProcessed
		}
		maturityDate := s.calculateMaturityDate(now, inv.MaxBusinessDays)
		inv.PaymentStatus = "completed"
		inv.Status = models.InvestmentStatusActive
		inv.StartedAt = &now
		inv.MaturityDate = &maturityDate
		inv.UpdatedAt = now
		if err := s.store.UpdateInvestment(ctx, inv); err != nil {
			return err
		}
	}

	// Mark payment transaction completed
	_ = s.store.UpdatePaymentTransaction(ctx, pt.ID, "completed", &now)

	// Lock investment capital in the investor USD wallet (idempotent).
	if err := s.creditLockedWallet(ctx, inv.UserID, inv.AmountUSD, "investment",
		fmt.Sprintf("CAPITAL-%s", inv.ID),
		fmt.Sprintf("Locked investment capital $%.2f (%s)", inv.AmountUSD, reference)); err != nil {
		s.logger.Warn("failed to lock investment capital in wallet",
			zap.String("investment_id", inv.ID),
			zap.Error(err),
		)
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
		"stage": "completed", "wallet_credited": true, "investment_created": true,
	})

	s.logger.Info("Investment activated",
		zap.String("investment_id", inv.ID),
		zap.String("user_id", inv.UserID),
		zap.String("provider", provider),
		zap.String("reference", reference),
		zap.Float64("amount_usd", inv.AmountUSD),
	)
	s.logger.Info("ROI schedule started",
		zap.String("investment_id", inv.ID),
		zap.Int("max_business_days", inv.MaxBusinessDays),
		zap.Float64("daily_reward_ngn", inv.DailyRewardNGN),
	)

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

	// Credit daily reward into locked wallet (USD), unlock later with referrals.
	rate, _ := s.store.GetExchangeRate(ctx)
	exchangeRate := 1400.0
	if rate != nil && rate.USDTNGN > 0 {
		exchangeRate = rate.USDTNGN
	}
	rewardUSD := amount / exchangeRate
	_ = s.creditLockedWallet(ctx, inv.UserID, rewardUSD, "roi",
		fmt.Sprintf("DAILY-%s-%d", inv.ID, nextBusinessDay),
		fmt.Sprintf("Locked daily reward day %d: ₦%.2f (~$%.2f)", nextBusinessDay, amount, rewardUSD))
	_, _ = s.UnlockEligibleEarnings(ctx, inv.UserID)

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

// ─── Wallet locked / available (USD investor ledger) ───────

// InvestorWalletCurrency is the unit used for Available vs Locked portfolio balances.
const InvestorWalletCurrency = "USD"

// creditLockedWallet posts to locked_balance with idempotent reference.
func (s *Service) creditLockedWallet(ctx context.Context, userID string, amount float64, txType, reference, description string) error {
	if amount <= 0 || !s.hasStore() {
		return nil
	}
	w, err := s.store.GetOrCreateWalletBalances(ctx, userID, InvestorWalletCurrency)
	if err != nil {
		return err
	}
	return s.store.CreditWalletLocked(ctx, w.ID, amount, txType, reference, description)
}

// UnlockEligibleEarnings moves profit + referral earnings from locked → available
// when successful_referrals >= min. Capital remains locked. Idempotent.
func (s *Service) UnlockEligibleEarnings(ctx context.Context, userID string) (unlocked float64, err error) {
	if !s.hasStore() {
		return 0, nil
	}
	settings, _ := s.store.GetSettings(ctx)
	minRefs := DefaultMinReferralsForWithdraw
	if settings != nil && settings.MinReferralsForPayout > 0 {
		minRefs = settings.MinReferralsForPayout
	}
	refs, err := s.store.CountSuccessfulReferrals(ctx, userID)
	if err != nil {
		return 0, err
	}
	if refs < minRefs {
		return 0, nil
	}

	// Eligible earnings = investment rewards + paid referral commissions.
	// Capital (active investment principal) stays locked.
	rewards, err := s.store.SumRewardsByUser(ctx, userID)
	if err != nil {
		return 0, err
	}
	// Convert NGN rewards to USD for the USD wallet.
	rate, _ := s.store.GetExchangeRate(ctx)
	exchangeRate := 1400.0
	if rate != nil && rate.USDTNGN > 0 {
		exchangeRate = rate.USDTNGN
	}
	rewardUSD := 0.0
	if exchangeRate > 0 {
		rewardUSD = rewards / exchangeRate
	}
	// Also sum investment.TotalEarned if rewards are in USD terms from pool seeds...
	// Pool seeds store NGN (profit * rate). Capital is credited in USD.
	// Referral commissions are NGN.
	referralNGN, _ := s.store.SumReferralCommissionsByReferrer(ctx, userID, "")
	referralUSD := 0.0
	if exchangeRate > 0 {
		referralUSD = referralNGN / exchangeRate
	}
	eligible := rewardUSD + referralUSD
	if eligible <= 0 {
		return 0, nil
	}

	w, err := s.store.GetOrCreateWalletBalances(ctx, userID, InvestorWalletCurrency)
	if err != nil {
		return 0, err
	}
	// Cap unlock to locked minus capital still active.
	capitalUSD, _ := s.sumActiveCapitalUSD(ctx, userID)
	maxUnlock := w.Locked - capitalUSD
	if maxUnlock < 0 {
		maxUnlock = 0
	}
	if eligible > maxUnlock {
		// Prefer unlocking whatever is above capital
		eligible = maxUnlock
	}
	if eligible <= 0 {
		// If capital not posted yet but profits are in locked, unlock min(eligible, locked)
		if w.Locked > 0 && capitalUSD <= 0 {
			eligible = rewardUSD + referralUSD
			if eligible > w.Locked {
				eligible = w.Locked
			}
		} else {
			return 0, nil
		}
	}

	ref := fmt.Sprintf("UNLOCK-EARNINGS-%s", userID)
	desc := fmt.Sprintf("Unlock investment profit and referral earnings after %d successful referrals", refs)
	if err := s.store.TransferLockedToAvailable(ctx, w.ID, eligible, ref, desc); err != nil {
		return 0, err
	}
	s.audit(ctx, userID, audit.ActionEarningsCredited, "wallet", w.ID, map[string]interface{}{
		"type":            "unlock_earnings",
		"amount_usd":      eligible,
		"successful_refs": refs,
	})
	s.createNotification(ctx, userID, "withdrawals_unlocked",
		"Withdrawals Unlocked",
		fmt.Sprintf("$%.2f moved from Locked to Available. You can now request withdrawals.", eligible),
		map[string]interface{}{"amount_usd": eligible, "referrals": refs})
	_ = s.store.InsertActivityEvent(ctx, userID, "wallet.earnings_unlocked", map[string]interface{}{
		"amount_usd": fmt.Sprintf("%.2f", eligible),
		"referrals":  fmt.Sprintf("%d", refs),
	})
	s.logger.Info("Eligible earnings unlocked to available",
		zap.String("user_id", userID),
		zap.Float64("amount_usd", eligible),
		zap.Int("referrals", refs),
	)
	return eligible, nil
}

func (s *Service) sumActiveCapitalUSD(ctx context.Context, userID string) (float64, error) {
	investments, _, err := s.store.ListUserInvestments(ctx, userID, "active", 1, 500)
	if err != nil {
		return 0, err
	}
	var sum float64
	for _, inv := range investments {
		sum += inv.AmountUSD
	}
	return sum, nil
}

// SyncInvestorWallet ensures capital and profit sit in Locked (or Available after unlock)
// using ledger-backed idempotent wallet transactions. No raw balance overwrites.
func (s *Service) SyncInvestorWallet(ctx context.Context, userID string) error {
	if !s.hasStore() {
		return nil
	}
	// 1) Ensure active capital is locked
	investments, _, err := s.store.ListUserInvestments(ctx, userID, "", 1, 500)
	if err != nil {
		return err
	}
	for _, inv := range investments {
		if inv.Status != models.InvestmentStatusActive && inv.Status != models.InvestmentStatusCompleted {
			continue
		}
		if inv.Status == models.InvestmentStatusActive {
			ref := fmt.Sprintf("CAPITAL-%s", inv.ID)
			_ = s.creditLockedWallet(ctx, userID, inv.AmountUSD, "investment", ref,
				fmt.Sprintf("Locked investment capital $%.2f", inv.AmountUSD))
		}
	}

	// 2) Ensure profit is locked (or was unlocked)
	rate, _ := s.store.GetExchangeRate(ctx)
	exchangeRate := 1400.0
	if rate != nil && rate.USDTNGN > 0 {
		exchangeRate = rate.USDTNGN
	}
	// Per-investment earned → locked credit
	for _, inv := range investments {
		if inv.TotalEarnedNGN <= 0 {
			continue
		}
		profitUSD := inv.TotalEarnedNGN / exchangeRate
		ref := fmt.Sprintf("PROFIT-%s", inv.ID)
		// If old seed credited to available with POOL- prefix, move to locked first.
		w, werr := s.store.GetOrCreateWalletBalances(ctx, userID, InvestorWalletCurrency)
		if werr == nil {
			oldRef := fmt.Sprintf("POOL-%s", inv.ID)
			if len(inv.ID) >= 8 {
				oldRef = fmt.Sprintf("POOL-%s", inv.ID[:8])
			}
			// Correct mis-posted available profit into locked (idempotent adjust).
			_ = s.store.MoveAvailableToLocked(ctx, w.ID, profitUSD,
				fmt.Sprintf("RELOCK-%s", inv.ID),
				"Move previously available profit into locked balance")
			_ = oldRef
		}
		_ = s.creditLockedWallet(ctx, userID, profitUSD, "roi", ref,
			fmt.Sprintf("Locked investment profit $%.2f", profitUSD))
	}

	// 3) Attempt unlock if referrals satisfied
	_, _ = s.UnlockEligibleEarnings(ctx, userID)
	return nil
}

// SyncAllInvestorWallets walks active investors and syncs locked/available wallets.
func (s *Service) SyncAllInvestorWallets(ctx context.Context) (int, error) {
	if !s.hasStore() {
		return 0, nil
	}
	investments, err := s.store.ListActiveInvestments(ctx, 1000)
	if err != nil {
		return 0, err
	}
	seen := map[string]struct{}{}
	n := 0
	for _, inv := range investments {
		if _, ok := seen[inv.UserID]; ok {
			continue
		}
		seen[inv.UserID] = struct{}{}
		if err := s.SyncInvestorWallet(ctx, inv.UserID); err != nil {
			s.logger.Warn("wallet sync failed", zap.String("user_id", inv.UserID), zap.Error(err))
			continue
		}
		n++
	}
	return n, nil
}

// PlatformReconciliation summarizes available vs locked across investor wallets.
func (s *Service) PlatformReconciliation(ctx context.Context) (map[string]interface{}, error) {
	avail, locked, total, users, err := s.store.SumPlatformWalletBalances(ctx, InvestorWalletCurrency)
	if err != nil {
		return nil, err
	}
	// Count locked vs eligible
	investments, _ := s.store.ListActiveInvestments(ctx, 2000)
	lockedUsers := 0
	eligibleUsers := 0
	seen := map[string]struct{}{}
	settings, _ := s.store.GetSettings(ctx)
	minRefs := DefaultMinReferralsForWithdraw
	if settings != nil && settings.MinReferralsForPayout > 0 {
		minRefs = settings.MinReferralsForPayout
	}
	for _, inv := range investments {
		if _, ok := seen[inv.UserID]; ok {
			continue
		}
		seen[inv.UserID] = struct{}{}
		refs, _ := s.store.CountSuccessfulReferrals(ctx, inv.UserID)
		if refs >= minRefs {
			eligibleUsers++
		} else {
			lockedUsers++
		}
	}
	return map[string]interface{}{
		"currency":                      InvestorWalletCurrency,
		"total_available_balance":       avail,
		"total_locked_balance":          locked,
		"total_portfolio_value":         total,
		"wallet_rows":                   users,
		"users_with_locked_withdrawals": lockedUsers,
		"users_eligible_to_withdraw":    eligibleUsers,
		"min_referrals_required":        minRefs,
	}, nil
}

// ─── Genesis pool profit credit (auditable earnings engine) ─

const (
	// PoolSeedRewardStatus marks one-time pool profit credits for idempotency.
	PoolSeedRewardStatus = "pool_seed"
	// GenesisPoolInvestmentUSD is the capital tier for the $30 Genesis pool.
	GenesisPoolInvestmentUSD = 30.0
	// GenesisPoolTotalUSD is the total profit pool to distribute.
	GenesisPoolTotalUSD = 500.0
	// GenesisPoolInvestors is the number of investors sharing the pool.
	GenesisPoolInvestors = 8
	// GenesisPoolProfitPerInvestorUSD = 500 / 8
	GenesisPoolProfitPerInvestorUSD = 62.5
	// DefaultMinReferralsForWithdraw unlocks withdrawals after N successful referrals.
	DefaultMinReferralsForWithdraw = 5
)

// SeedGenesisPoolProfits credits each active $30 Genesis investor their equal
// share of the $500 profit pool ($62.50) through the normal earnings path:
// investment_rewards ledger, investment totals, wallet transaction, audit,
// notification, and activity event. Idempotent per investment.
func (s *Service) SeedGenesisPoolProfits(ctx context.Context) (*models.SeedPoolCreditSummary, error) {
	return s.SeedPoolProfits(ctx, GenesisPoolInvestmentUSD, GenesisPoolProfitPerInvestorUSD, GenesisPoolInvestors, GenesisPoolTotalUSD)
}

// SeedPoolProfits distributes a fixed profit_usd to each active investment matching investmentUSD.
func (s *Service) SeedPoolProfits(
	ctx context.Context,
	investmentUSD, profitPerInvestorUSD float64,
	maxInvestors int,
	poolUSD float64,
) (*models.SeedPoolCreditSummary, error) {
	if !s.hasStore() {
		return nil, errors.ErrSettingsNotFound
	}
	if profitPerInvestorUSD <= 0 || investmentUSD <= 0 {
		return nil, errors.ErrInvalidAmount
	}

	rate, err := s.store.GetExchangeRate(ctx)
	if err != nil {
		return nil, err
	}
	exchangeRate := 1400.0
	if rate != nil && rate.USDTNGN > 0 {
		exchangeRate = rate.USDTNGN
	}

	investments, err := s.store.ListActiveInvestmentsByAmountUSD(ctx, investmentUSD, maxInvestors)
	if err != nil {
		return nil, err
	}

	summary := &models.SeedPoolCreditSummary{
		InvestorsTargeted: len(investments),
		ProfitPerInvestor: profitPerInvestorUSD,
		PoolUSD:           poolUSD,
		ExchangeRate:      exchangeRate,
		Results:           make([]models.SeedPoolCreditResult, 0, len(investments)),
	}

	profitNGN := profitPerInvestorUSD * exchangeRate
	today := time.Now().UTC().Truncate(24 * time.Hour)

	for _, inv := range investments {
		result := models.SeedPoolCreditResult{
			UserID:       inv.UserID,
			InvestmentID: inv.ID,
			AmountUSD:    inv.AmountUSD,
			ProfitUSD:    profitPerInvestorUSD,
			ProfitNGN:    profitNGN,
			PortfolioUSD: inv.AmountUSD + profitPerInvestorUSD,
		}

		exists, err := s.store.HasRewardWithStatus(ctx, inv.ID, PoolSeedRewardStatus)
		if err != nil {
			return nil, err
		}
		if exists {
			result.AlreadyCredited = true
			summary.InvestorsSkipped++
			summary.Results = append(summary.Results, result)
			continue
		}

		// Avoid unique (investment_id, reward_date) collision with a daily reward.
		rewardDate := today
		if existing, _ := s.store.GetRewardByDate(ctx, inv.ID, rewardDate); existing != nil {
			rewardDate = today.AddDate(0, 0, 1)
		}

		// Business day number uses a high sentinel so it does not collide with 1..20 daily days.
		businessDay := inv.PaidBusinessDays + 100
		if businessDay < 100 {
			businessDay = 100
		}

		reward := &models.InvestmentReward{
			InvestmentID:      inv.ID,
			UserID:            inv.UserID,
			AmountNGN:         profitNGN,
			RewardDate:        rewardDate,
			BusinessDayNumber: businessDay,
			Status:            PoolSeedRewardStatus,
		}
		rewardID, err := s.store.CreateRewardReturning(ctx, reward)
		if err != nil {
			// Fallback to non-returning insert
			if err2 := s.store.CreateReward(ctx, reward); err2 != nil {
				s.logger.Error("failed to create pool reward",
					zap.String("investment_id", inv.ID),
					zap.Error(err2),
				)
				return nil, err2
			}
			rewardID = ""
		}
		result.RewardID = rewardID

		// Update investment earned totals (earnings engine source of truth).
		inv.TotalEarnedNGN += profitNGN
		inv.UpdatedAt = time.Now().UTC()
		if err := s.store.UpdateInvestment(ctx, inv); err != nil {
			return nil, err
		}

		// Lock investment capital + profit in USD wallet (visible, not withdrawable).
		// Capital
		_ = s.creditLockedWallet(ctx, inv.UserID, inv.AmountUSD, "investment",
			fmt.Sprintf("CAPITAL-%s", inv.ID),
			fmt.Sprintf("Locked investment capital $%.2f", inv.AmountUSD))
		// Profit (locked until referrals unlock)
		if err := s.creditLockedWallet(ctx, inv.UserID, profitPerInvestorUSD, "roi",
			fmt.Sprintf("PROFIT-%s", inv.ID),
			fmt.Sprintf("Locked Genesis pool profit $%.2f (₦%.2f)", profitPerInvestorUSD, profitNGN)); err != nil {
			s.logger.Warn("locked wallet profit credit failed (reward ledger still recorded)",
				zap.String("user_id", inv.UserID),
				zap.Error(err),
			)
		} else {
			s.logger.Info("Wallet locked profit credited",
				zap.String("user_id", inv.UserID),
				zap.Float64("profit_usd", profitPerInvestorUSD),
				zap.Float64("locked_portfolio_usd", inv.AmountUSD+profitPerInvestorUSD),
			)
		}
		// Best-effort: if referrals already met, unlock earnings now.
		_, _ = s.UnlockEligibleEarnings(ctx, inv.UserID)

		// Audit log
		s.audit(ctx, inv.UserID, audit.ActionEarningsCredited, "earnings_investment", inv.ID, map[string]interface{}{
			"type":           "genesis_pool_seed",
			"profit_usd":     profitPerInvestorUSD,
			"profit_ngn":     profitNGN,
			"exchange_rate":  exchangeRate,
			"portfolio_usd":  inv.AmountUSD + profitPerInvestorUSD,
			"reward_status":  PoolSeedRewardStatus,
			"reward_id":      rewardID,
			"investment_usd": inv.AmountUSD,
		})

		// Earning history notification / activity feed
		s.createNotification(ctx, inv.UserID, "pool_profit_credited",
			"Genesis Profit Credited",
			fmt.Sprintf("You earned $%.2f (₦%.2f) profit on your $%.2f Genesis investment. Portfolio value: $%.2f.",
				profitPerInvestorUSD, profitNGN, inv.AmountUSD, inv.AmountUSD+profitPerInvestorUSD),
			map[string]interface{}{
				"investment_id": inv.ID,
				"profit_usd":    profitPerInvestorUSD,
				"profit_ngn":    profitNGN,
				"portfolio_usd": inv.AmountUSD + profitPerInvestorUSD,
			})

		_ = s.store.InsertActivityEvent(ctx, inv.UserID, "earnings.pool_credited", map[string]interface{}{
			"investment_id": inv.ID,
			"profit_usd":    fmt.Sprintf("%.2f", profitPerInvestorUSD),
			"profit_ngn":    fmt.Sprintf("%.2f", profitNGN),
		})

		s.publish("earnings.pool_credited", map[string]interface{}{
			"user_id":       inv.UserID,
			"investment_id": inv.ID,
			"profit_usd":    profitPerInvestorUSD,
			"profit_ngn":    profitNGN,
		})

		s.logger.Info("Genesis pool profit credited",
			zap.String("user_id", inv.UserID),
			zap.String("investment_id", inv.ID),
			zap.Float64("profit_usd", profitPerInvestorUSD),
			zap.Float64("profit_ngn", profitNGN),
			zap.Float64("portfolio_usd", inv.AmountUSD+profitPerInvestorUSD),
		)

		summary.InvestorsCredited++
		summary.TotalProfitUSD += profitPerInvestorUSD
		summary.Results = append(summary.Results, result)
	}

	return summary, nil
}

// ─── Withdrawal Processing ───────────────────────────────

// defaultWithdrawalInterval is the fallback lock between withdrawal requests (7 days).
const defaultWithdrawalInterval = 7 * 24 * time.Hour

func (s *Service) RequestWithdrawal(ctx context.Context, userID string, req *models.WithdrawalRequest) (*models.Withdrawal, error) {
	settings, err := s.store.GetSettings(ctx)
	if err != nil {
		return nil, err
	}
	if settings == nil {
		return nil, errors.ErrSettingsNotFound
	}

	// Referral unlock gate: withdrawals_unlocked = successful_referrals >= min (default 5).
	minRefs := settings.MinReferralsForPayout
	if minRefs <= 0 {
		minRefs = DefaultMinReferralsForWithdraw
	}
	activeRefs, err := s.store.CountSuccessfulReferrals(ctx, userID)
	if err != nil {
		return nil, err
	}
	if activeRefs < minRefs {
		return nil, errors.ErrReferralsRequired
	}

	// Enforce the one-withdrawal-every-N-days rule (default 7).
	interval := defaultWithdrawalInterval
	if settings.WithdrawalIntervalDays > 0 {
		interval = time.Duration(settings.WithdrawalIntervalDays) * 24 * time.Hour
	}
	last, err := s.store.GetLastWithdrawal(ctx, userID)
	if err != nil {
		return nil, err
	}
	if last != nil {
		elapsed := time.Since(last.CreatedAt)
		if elapsed < interval {
			return nil, errors.ErrWithdrawalLocked
		}
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
	// Ensure unlock has been applied when eligible.
	_, _ = s.UnlockEligibleEarnings(ctx, userID)

	rate, _ := s.store.GetExchangeRate(ctx)
	exchangeRate := 1400.0
	if rate != nil && rate.USDTNGN > 0 {
		exchangeRate = rate.USDTNGN
	}

	// Withdrawable cash lives in wallet available balance (USD), converted to NGN for request amounts.
	w, err := s.store.GetOrCreateWalletBalances(ctx, userID, InvestorWalletCurrency)
	if err != nil {
		return nil, err
	}
	availableNGN := w.Available * exchangeRate

	// Also respect earnings ledger - pending withdrawals reduce free cash
	pendingWithdrawals, err := s.store.SumPendingWithdrawalsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	availableBalance := availableNGN - pendingWithdrawals
	if availableBalance < 0 {
		availableBalance = 0
	}
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
	if !s.hasStore() {
		return &models.EarningsDashboard{
			TotalInvestedUSD:     0,
			TotalInvestedNGN:     0,
			ExchangeRate:         1600,
			TodayEarningsNGN:     0,
			MonthlyEarningsNGN:   0,
			AvailableBalanceNGN:  0,
			PendingWithdrawalNGN: 0,
			ReferralEarningsNGN:  0,
			Investments:          []*models.EarningsSummary{},
		}, nil
	}

	// Keep wallet locked/available in sync with investments & earnings (idempotent).
	_ = s.SyncInvestorWallet(ctx, userID)

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

	// Load the last withdrawal to power the weekly withdrawal countdown.
	var lastWithdrawalAt *time.Time
	if last, err := s.store.GetLastWithdrawal(ctx, userID); err == nil && last != nil {
		lastWithdrawalAt = &last.CreatedAt
	}

	settings, _ := s.store.GetSettings(ctx)
	minRefs := DefaultMinReferralsForWithdraw
	if settings != nil && settings.MinReferralsForPayout > 0 {
		minRefs = settings.MinReferralsForPayout
	}
	successfulRefs, _ := s.store.CountSuccessfulReferrals(ctx, userID)
	remainingRefs := minRefs - successfulRefs
	if remainingRefs < 0 {
		remainingRefs = 0
	}
	unlocked := successfulRefs >= minRefs
	lockMsg := ""
	if !unlocked {
		lockMsg = fmt.Sprintf("Complete %d successful referrals to unlock your earnings.", minRefs)
	}

	dash := &models.EarningsDashboard{
		TotalInvestedUSD:      0,
		TotalInvestedNGN:      0,
		ExchangeRate:          exchangeRate,
		TodayEarningsNGN:      todayEarnings,
		MonthlyEarningsNGN:    monthlyEarnings,
		AvailableBalanceNGN:   totalEarnings - pendingWithdrawals,
		PendingWithdrawalNGN:  pendingWithdrawals,
		ReferralEarningsNGN:   referralEarnings,
		LastWithdrawalAt:      lastWithdrawalAt,
		WithdrawalsUnlocked:   unlocked,
		WithdrawalLockMessage: lockMsg,
		ActiveReferrals:       successfulRefs,
		MinReferralsRequired:  minRefs,
		RemainingReferrals:    remainingRefs,
	}

	var summaries []*models.EarningsSummary
	totalProfitNGN := 0.0
	var lockedUSD, lockedNGN float64

	for _, inv := range investments {
		dash.TotalInvestedUSD += inv.AmountUSD
		dash.TotalInvestedNGN += inv.AmountNGN
		totalProfitNGN += inv.TotalEarnedNGN

		switch inv.Status {
		case models.InvestmentStatusActive:
			dash.ActiveInvestments++
			// Capital remains locked while the plan is active.
			lockedUSD += inv.AmountUSD
			lockedNGN += inv.AmountNGN
		case models.InvestmentStatusCompleted:
			dash.CompletedInvestments++
		case models.InvestmentStatusPendingPayment:
			// Pending payments are not yet locked capital.
		}

		rateInv := inv.ExchangeRate
		if rateInv <= 0 {
			rateInv = exchangeRate
		}
		earnedUSD := 0.0
		if rateInv > 0 {
			earnedUSD = inv.TotalEarnedNGN / rateInv
		}
		roiPct := 0.0
		if inv.AmountUSD > 0 {
			roiPct = (earnedUSD / inv.AmountUSD) * 100
		}

		summary := &models.EarningsSummary{
			ID:                inv.ID,
			AmountUSD:         inv.AmountUSD,
			AmountNGN:         inv.AmountNGN,
			ExchangeRate:      inv.ExchangeRate,
			DailyRewardNGN:    inv.DailyRewardNGN,
			PaidBusinessDays:  inv.PaidBusinessDays,
			MaxBusinessDays:   inv.MaxBusinessDays,
			TotalEarnedNGN:    inv.TotalEarnedNGN,
			TotalEarnedUSD:    earnedUSD,
			TotalPendingNGN:   inv.TotalPendingNGN,
			PortfolioValueUSD: inv.AmountUSD + earnedUSD,
			PortfolioValueNGN: inv.AmountNGN + inv.TotalEarnedNGN,
			ROIPercentage:     math.Round(roiPct*100) / 100,
			Status:            inv.Status,
			MaturityDate:      inv.MaturityDate,
			StartedAt:         inv.StartedAt,
			CreatedAt:         inv.CreatedAt,
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
	dash.TotalProfitNGN = totalProfitNGN
	if exchangeRate > 0 {
		dash.TotalProfitUSD = totalProfitNGN / exchangeRate
	}
	// Prefer sum of rewards ledger if investment totals lag
	if totalEarnings > totalProfitNGN {
		dash.TotalProfitNGN = totalEarnings
		if exchangeRate > 0 {
			dash.TotalProfitUSD = totalEarnings / exchangeRate
		}
	}
	// Explicit portfolio aliases for clients (Bybit-style position view).
	dash.CapitalInvestedUSD = dash.TotalInvestedUSD
	dash.CapitalInvestedNGN = dash.TotalInvestedNGN
	dash.ProfitEarnedUSD = dash.TotalProfitUSD
	dash.ProfitEarnedNGN = dash.TotalProfitNGN

	// Wallet Available / Locked from USD investor wallet (source of truth for UI).
	// Available = unlocked earnings. Locked = capital + still-locked earnings/referrals.
	dash.ReferralEarningsUSD = 0
	if exchangeRate > 0 {
		dash.ReferralEarningsUSD = referralEarnings / exchangeRate
	}

	walletUSD, werr := s.store.GetOrCreateWalletBalances(ctx, userID, InvestorWalletCurrency)
	if werr == nil && walletUSD != nil && (walletUSD.Available > 0 || walletUSD.Locked > 0 || walletUSD.Total > 0) {
		dash.AvailableBalanceUSD = walletUSD.Available
		dash.LockedBalanceUSD = walletUSD.Locked
		dash.AvailableBalanceNGN = walletUSD.Available * exchangeRate
		dash.LockedBalanceNGN = walletUSD.Locked * exchangeRate
		// Portfolio = available + locked
		dash.PortfolioValueUSD = walletUSD.Available + walletUSD.Locked
		dash.PortfolioValueNGN = dash.PortfolioValueUSD * exchangeRate
		if unlocked {
			dash.WithdrawableBalanceUSD = walletUSD.Available
			dash.WithdrawableBalanceNGN = walletUSD.Available * exchangeRate
		} else {
			dash.WithdrawableBalanceUSD = 0
			dash.WithdrawableBalanceNGN = 0
		}
	} else {
		// Fallback when wallet not yet synced
		if unlocked {
			dash.AvailableBalanceUSD = dash.TotalProfitUSD
			dash.LockedBalanceUSD = lockedUSD
			dash.WithdrawableBalanceUSD = dash.TotalProfitUSD
			dash.WithdrawableBalanceNGN = dash.TotalProfitNGN - pendingWithdrawals
			if dash.WithdrawableBalanceNGN < 0 {
				dash.WithdrawableBalanceNGN = 0
			}
		} else {
			dash.AvailableBalanceUSD = 0
			dash.LockedBalanceUSD = lockedUSD + dash.TotalProfitUSD + dash.ReferralEarningsUSD
			dash.WithdrawableBalanceUSD = 0
			dash.WithdrawableBalanceNGN = 0
		}
		dash.AvailableBalanceNGN = dash.AvailableBalanceUSD * exchangeRate
		dash.LockedBalanceNGN = dash.LockedBalanceUSD * exchangeRate
		dash.PortfolioValueUSD = dash.AvailableBalanceUSD + dash.LockedBalanceUSD
		dash.PortfolioValueNGN = dash.PortfolioValueUSD * exchangeRate
	}
	if dash.TotalInvestedUSD > 0 {
		dash.ROIPercentage = math.Round((dash.TotalProfitUSD/dash.TotalInvestedUSD)*10000) / 100
	}

	// Get referral info
	referralInfo := &models.ReferralInfo{
		ReferralCode:           "",
		ReferralLink:           "",
		TotalReferrals:         0,
		ActiveReferrals:        successfulRefs,
		ReferralEarningsNGN:    referralEarnings,
		WithdrawableBalanceNGN: 0,
		MinimumTarget:          minRefs,
	}
	if unlocked {
		referralInfo.WithdrawableBalanceNGN = dash.AvailableBalanceNGN
	}

	totalRefs, _ := s.store.CountReferralsByReferrer(ctx, userID)
	referralInfo.TotalReferrals = totalRefs
	if referralInfo.TotalReferrals < successfulRefs {
		referralInfo.TotalReferrals = successfulRefs
	}

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

// ─── Demo Seeding (development only) ─────────────────────

// SeedDemoInvestments creates a demo Genesis investment for every active user
// that has no real (non-demo) investment yet. It is idempotent: running it
// multiple times never duplicates investments or wallet balances.
//
// Demo plan: $30 capital, 19% ROI, 20 business days, 10% referral bonus.
// All demo records are flagged is_demo = true so they can be purged later.
func (s *Service) SeedDemoInvestments(ctx context.Context) (int, error) {
	if !s.hasStore() {
		return 0, nil
	}

	const (
		demoAmountUSD   = 30.0
		demoROIPercent  = 19.0
		demoWorkingDays = 20
	)

	userIDs, err := s.store.ListIdentityUserIDs(ctx, 10000)
	if err != nil {
		return 0, err
	}

	seeded := 0
	for _, userID := range userIDs {
		// Skip users who already have a real investment or a demo investment.
		hasReal, err := s.store.UserHasRealInvestment(ctx, userID)
		if err != nil {
			return seeded, err
		}
		if hasReal {
			continue
		}
		hasDemo, err := s.store.UserHasDemoInvestment(ctx, userID)
		if err != nil {
			return seeded, err
		}
		if hasDemo {
			continue
		}

		// Exchange rate for NGN display.
		rate := 1400.0
		if r, err := s.store.GetExchangeRate(ctx); err == nil && r != nil {
			rate = r.USDTNGN
		}

		amountNGN := demoAmountUSD * rate
		dailyReward := (amountNGN * (demoROIPercent / 100.0)) / demoWorkingDays
		totalPending := dailyReward * demoWorkingDays

		now := time.Now().UTC()
		maturityDate := s.calculateMaturityDate(now, demoWorkingDays)

		inv := &models.EarningsInvestment{
			ID:               uuidlib.NewString(),
			UserID:           userID,
			AmountUSD:        demoAmountUSD,
			AmountNGN:        amountNGN,
			ExchangeRate:     rate,
			PaymentProvider:  "demo",
			PaymentReference: fmt.Sprintf("DEMO-GENESIS-%s", uuidlib.NewString()[:8]),
			PaymentStatus:    "completed",
			DailyRewardNGN:   dailyReward,
			MaxBusinessDays:  demoWorkingDays,
			PaidBusinessDays: 0,
			TotalEarnedNGN:   0,
			TotalPendingNGN:  totalPending,
			Status:           models.InvestmentStatusActive,
			IsDemo:           true,
			PlanName:         "Seed Plan",
			MaturityDate:     &maturityDate,
			StartedAt:        &now,
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		if err := s.store.CreateInvestment(ctx, inv); err != nil {
			return seeded, err
		}

		// Credit the user's USD wallet with the demo available balance ($30).
		wallet, err := s.store.GetOrCreateWalletBalances(ctx, userID, "USD")
		if err != nil {
			return seeded, err
		}
		if err := s.store.CreditWalletAvailable(ctx, wallet.ID, demoAmountUSD, "demo_seed",
			fmt.Sprintf("DEMO-SEED-%s", inv.ID),
			fmt.Sprintf("Demo Genesis seed balance $%.2f", demoAmountUSD)); err != nil {
			return seeded, err
		}

		seeded++
	}

	s.logger.Info("demo investments seeded",
		zap.Int("seeded", seeded),
		zap.Int("users_checked", len(userIDs)),
	)
	return seeded, nil
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

func (s *Service) paystackCallbackURL() string {
	// Must come from PAYSTACK_CALLBACK_URL (e.g. https://coindistro-hazel.vercel.app/app/earn).
	if cb := strings.TrimSpace(s.cfg.PaystackCallbackURL); cb != "" {
		return cb
	}
	// Defensive fallback only — startup validation requires PAYSTACK_CALLBACK_URL.
	base := strings.TrimRight(s.cfg.AppURL, "/")
	if base == "" {
		base = strings.TrimRight(s.cfg.BaseURL, "/")
	}
	return base + "/app/earn"
}

func (s *Service) callPaystackInitialize(ctx context.Context, amount float64, currency, reference, userID, email string) (string, string, error) {
	if s.cfg.PaystackSecretKey == "" {
		return "", "", errors.ErrGatewayNotConfigured
	}

	amountInKobo := int64(amount * 100)
	payload := map[string]interface{}{
		"amount":       amountInKobo,
		"currency":     currency,
		"email":        email,
		"reference":    reference,
		"callback_url": s.paystackCallbackURL(),
		"metadata": map[string]interface{}{
			"user_id":     userID,
			"platform":    "coindistro",
			"payment_for": "earnings_investment",
		},
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
	// Prefer PAYSTACK_WEBHOOK_SECRET; fall back to PAYSTACK_SECRET_KEY.
	secret := strings.TrimSpace(s.cfg.PaystackWebhookSecret)
	if secret == "" {
		secret = s.cfg.PaystackSecretKey
	}
	if secret == "" || signature == "" {
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
