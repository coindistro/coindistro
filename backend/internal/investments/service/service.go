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
	apperrors "github.com/coindistro/backend/internal/errors"
	"github.com/coindistro/backend/internal/events"
	"github.com/coindistro/backend/internal/investments/errors"
	"github.com/coindistro/backend/internal/investments/models"
	"github.com/coindistro/backend/internal/investments/store"
	"github.com/coindistro/backend/internal/metrics"
	uuidlib "github.com/coindistro/backend/internal/uuid"
	"github.com/coindistro/backend/internal/workers"
)

// Service implements investment business logic.
type Service struct {
	store       *store.Store
	eventBus    *events.InMemoryBus
	jobRegistry *workers.Registry
	workerPool  *workers.Pool
	auditLogger *audit.Logger
	promMetrics *metrics.Metrics
	logger      *zap.Logger
	cfg         Config
}

func (s *Service) hasStore() bool {
	return s != nil && s.store != nil
}

// Config holds the investment service configuration.
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
}

// New creates the investment service.
func New(
	st *store.Store,
	eventBus *events.InMemoryBus,
	jobRegistry *workers.Registry,
	workerPool *workers.Pool,
	auditLogger *audit.Logger,
	promMetrics *metrics.Metrics,
	logger *zap.Logger,
	cfg Config,
) *Service {
	svc := &Service{
		store:       st,
		eventBus:    eventBus,
		jobRegistry: jobRegistry,
		workerPool:  workerPool,
		auditLogger: auditLogger,
		promMetrics: promMetrics,
		logger:      logger,
		cfg:         cfg,
	}
	svc.registerWorkers()
	return svc
}

// ─── Plan Management ──────────────────────────────────

func (s *Service) ListPlans(ctx context.Context, onlyEnabled bool) ([]*models.InvestmentPlan, error) {
	if !s.hasStore() {
		return []*models.InvestmentPlan{}, nil
	}
	return s.store.ListPlans(ctx, onlyEnabled)
}

func (s *Service) GetPlan(ctx context.Context, id string) (*models.InvestmentPlan, error) {
	if !s.hasStore() {
		return nil, errors.ErrPlanNotFound
	}
	p, err := s.store.GetPlanByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, errors.ErrPlanNotFound
	}
	return p, nil
}

func (s *Service) CreatePlan(ctx context.Context, req *models.CreatePlanRequest, actorID string) (*models.InvestmentPlan, error) {
	existing, _ := s.store.GetPlanByName(ctx, req.Name)
	if existing != nil {
		return nil, errors.ErrPlanNameExists
	}

	now := time.Now().UTC()
	p := &models.InvestmentPlan{
		ID:            uuidlib.NewString(),
		Name:          req.Name,
		Description:   req.Description,
		MinimumAmount: req.MinimumAmount,
		MaximumAmount: req.MaximumAmount,
		Currency:      strings.ToUpper(req.Currency),
		ROIPercent:    req.ROIPercent,
		Enabled:       req.Enabled,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.store.CreatePlan(ctx, p); err != nil {
		return nil, err
	}

	s.audit(ctx, actorID, audit.ActionAdminAction, "investment_plan", p.ID, map[string]interface{}{
		"name": p.Name, "roi_percent": p.ROIPercent,
	})
	return p, nil
}

func (s *Service) UpdatePlan(ctx context.Context, id string, req *models.UpdatePlanRequest, actorID string) (*models.InvestmentPlan, error) {
	p, err := s.GetPlan(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		existing, _ := s.store.GetPlanByName(ctx, *req.Name)
		if existing != nil && existing.ID != id {
			return nil, errors.ErrPlanNameExists
		}
		p.Name = *req.Name
	}
	if req.Description != nil {
		p.Description = *req.Description
	}
	if req.MinimumAmount != nil {
		p.MinimumAmount = *req.MinimumAmount
	}
	if req.MaximumAmount != nil {
		p.MaximumAmount = *req.MaximumAmount
	}
	if req.Currency != nil {
		p.Currency = strings.ToUpper(*req.Currency)
	}
	if req.ROIPercent != nil {
		p.ROIPercent = *req.ROIPercent
	}
	if req.Enabled != nil {
		p.Enabled = *req.Enabled
	}
	p.UpdatedAt = time.Now().UTC()

	if err := s.store.UpdatePlan(ctx, p); err != nil {
		return nil, err
	}

	s.audit(ctx, actorID, audit.ActionAdminAction, "investment_plan", p.ID, map[string]interface{}{
		"name": p.Name, "enabled": p.Enabled,
	})
	return p, nil
}

func (s *Service) DeletePlan(ctx context.Context, id string, actorID string) error {
	p, err := s.GetPlan(ctx, id)
	if err != nil {
		return err
	}
	if err := s.store.DeletePlan(ctx, id); err != nil {
		return err
	}
	s.audit(ctx, actorID, audit.ActionAdminAction, "investment_plan", id, map[string]interface{}{
		"name": p.Name,
	})
	return nil
}

// ─── Pricing ──────────────────────────────────────────

func (s *Service) GetCurrentPricing(ctx context.Context) (*models.Pricing, error) {
	p, err := s.store.GetCurrentPricing(ctx)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, errors.ErrPricingNotFound
	}
	return p, nil
}

func (s *Service) SetPricing(ctx context.Context, price float64, actorID string) (*models.Pricing, error) {
	if err := s.store.SetPricing(ctx, price, actorID); err != nil {
		return nil, err
	}
	p, err := s.store.GetCurrentPricing(ctx)
	if err != nil {
		return nil, err
	}
	s.audit(ctx, actorID, audit.ActionAdminAction, "cdt_pricing", p.ID, map[string]interface{}{
		"price_ngn": price,
	})
	return p, nil
}

func (s *Service) GetPricingHistory(ctx context.Context) ([]*models.Pricing, error) {
	return s.store.ListPricingHistory(ctx, 20)
}

// ─── Payment Initialization ───────────────────────────

func (s *Service) InitPaystackPayment(ctx context.Context, userID string, req *models.InitPaymentRequest) (*models.InitPaymentResponse, error) {
	if s.cfg.PaystackSecretKey == "" {
		return nil, errors.ErrGatewayNotConfigured
	}
	if !s.hasStore() {
		return nil, errors.ErrGatewayNotConfigured
	}

	plan, pricing, err := s.validateInvestmentRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	allocatedCDT := req.Amount / pricing.PriceNGN
	reference := fmt.Sprintf("CDT-PS-%s-%d", uuidlib.NewString()[:8], time.Now().Unix())

	// Serialize investment params to store on the payment transaction so they
	// can be used to create the investment AFTER payment verification succeeds.
	now := time.Now().UTC()
	roiCDT := allocatedCDT * (plan.ROIPercent / 100.0)
	params := models.InvestmentParams{
		PlanID:         plan.ID,
		PlanName:       plan.Name,
		ROIPercent:     plan.ROIPercent,
		LockPeriodDays: req.LockPeriodDays,
		AllocatedCDT:   allocatedCDT,
		ROICDT:         roiCDT,
		CDTPrice:       pricing.PriceNGN,
	}
	paramsJSON, _ := json.Marshal(params)

	// Create pending payment transaction (NO investment created yet).
	// CRITICAL: Investment + wallet credit happen only after Paystack
	// verifies status == "success" via webhook or VerifyPaystackPayment.
	pt := &models.PaymentTransaction{
		ID:        uuidlib.NewString(),
		UserID:    userID,
		Provider:  "paystack",
		Reference: reference,
		Status:    "pending",
		Amount:    req.Amount,
		Currency:  req.Currency,
		Response:  paramsJSON,
		CreatedAt: now,
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

	// Call Paystack initialize API (secret key from PAYSTACK_SECRET_KEY).
	authURL, accessCode, err := s.callPaystackInitialize(ctx, req.Amount, req.Currency, reference, userID, email)
	if err != nil {
		return nil, err
	}

	s.audit(ctx, userID, audit.ActionDeposit, "payment_transaction", reference, map[string]interface{}{
		"provider": "paystack", "reference": reference, "amount": req.Amount,
		"stage": "initialized", "wallet_credited": false, "investment_created": false,
	})

	return &models.InitPaymentResponse{
		AuthorizationURL: authURL,
		Reference:        reference,
		AccessCode:       accessCode,
	}, nil
}

// VerifyPaystackPayment confirms a checkout after the user returns from Paystack
// (or if the webhook is delayed). It re-verifies with Paystack's API — never trusts the client.
// Investment + wallet credit happen only after verified success.
func (s *Service) VerifyPaystackPayment(ctx context.Context, userID, reference string) (*models.Investment, error) {
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

	// Already activated (webhook may have won the race)?
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

	// Verify with Paystack API — never trust the client's claim of success.
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

	s.logger.Info("Paystack verification succeeded",
		zap.String("reference", reference),
		zap.String("user_id", userID),
	)

	if err := s.processSuccessfulPayment(ctx, "paystack", reference, pt.Amount, pt.Currency); err != nil {
		if err == errors.ErrPaymentAlreadyProcessed {
			return s.store.GetInvestmentByReference(ctx, "paystack", reference)
		}
		return nil, err
	}

	return s.store.GetInvestmentByReference(ctx, "paystack", reference)
}

func (s *Service) InitFlutterwavePayment(ctx context.Context, userID string, req *models.InitPaymentRequest) (*models.InitPaymentResponse, error) {
	if !s.hasStore() {
		reference := fmt.Sprintf("CDT-FW-%s-%d", uuidlib.NewString()[:8], time.Now().Unix())
		return &models.InitPaymentResponse{
			AuthorizationURL: fmt.Sprintf("%s/checkout/flutterwave/%s", strings.TrimRight(s.cfg.BaseURL, "/"), reference),
			Reference:        reference,
		}, nil
	}

	plan, pricing, err := s.validateInvestmentRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	allocatedCDT := req.Amount / pricing.PriceNGN
	reference := fmt.Sprintf("CDT-FW-%s-%d", uuidlib.NewString()[:8], time.Now().Unix())

	// Serialize investment params to store on the payment transaction so they
	// can be used to create the investment AFTER payment verification succeeds.
	now := time.Now().UTC()
	roiCDT := allocatedCDT * (plan.ROIPercent / 100.0)
	params := models.InvestmentParams{
		PlanID:         plan.ID,
		PlanName:       plan.Name,
		ROIPercent:     plan.ROIPercent,
		LockPeriodDays: req.LockPeriodDays,
		AllocatedCDT:   allocatedCDT,
		ROICDT:         roiCDT,
		CDTPrice:       pricing.PriceNGN,
	}
	paramsJSON, _ := json.Marshal(params)

	// Create pending payment transaction (NO investment created yet).
	// CRITICAL: Investment + wallet credit happen only after Flutterwave
	// verifies status == "successful" via webhook.
	pt := &models.PaymentTransaction{
		ID:        uuidlib.NewString(),
		UserID:    userID,
		Provider:  "flutterwave",
		Reference: reference,
		Status:    "pending",
		Amount:    req.Amount,
		Currency:  req.Currency,
		Response:  paramsJSON,
		CreatedAt: now,
	}
	if err := s.store.CreatePaymentTransaction(ctx, pt); err != nil {
		return nil, err
	}

	// Call Flutterwave initialize API
	authURL, err := s.callFlutterwaveInitialize(ctx, req.Amount, req.Currency, reference, userID)
	if err != nil {
		return nil, err
	}

	s.audit(ctx, userID, audit.ActionDeposit, "payment_transaction", reference, map[string]interface{}{
		"provider": "flutterwave", "reference": reference, "amount": req.Amount,
		"stage": "initialized", "wallet_credited": false, "investment_created": false,
	})

	return &models.InitPaymentResponse{
		AuthorizationURL: authURL,
		Reference:        reference,
	}, nil
}

func (s *Service) validateInvestmentRequest(ctx context.Context, req *models.InitPaymentRequest) (*models.InvestmentPlan, *models.Pricing, error) {
	if !s.hasStore() {
		plan := &models.InvestmentPlan{
			ID:            "fallback-plan",
			Name:          "CoinDistro Plan",
			MinimumAmount: 30,
			MaximumAmount: 1000000,
			Currency:      "USD",
			ROIPercent:    30,
			Enabled:       true,
		}
		pricing := &models.Pricing{PriceNGN: 1600}
		return plan, pricing, nil
	}

	// Validate lock period
	validPeriod := false
	for _, d := range models.SupportedLockPeriods {
		if d == req.LockPeriodDays {
			validPeriod = true
			break
		}
	}
	if !validPeriod {
		return nil, nil, errors.ErrInvalidLockPeriod
	}

	// Get plan
	plan, err := s.store.GetPlanByID(ctx, req.PlanID)
	if err != nil {
		return nil, nil, err
	}
	if plan == nil {
		return nil, nil, errors.ErrPlanNotFound
	}
	if !plan.Enabled {
		return nil, nil, errors.ErrPlanDisabled
	}

	// Validate amount
	if req.Amount < plan.MinimumAmount || req.Amount > plan.MaximumAmount {
		return nil, nil, errors.ErrInvalidAmount
	}

	// Get current pricing
	pricing, err := s.store.GetCurrentPricing(ctx)
	if err != nil {
		return nil, nil, err
	}
	if pricing == nil {
		return nil, nil, errors.ErrPricingNotFound
	}

	return plan, pricing, nil
}

// ─── Webhook Processing ───────────────────────────────

// ProcessPaystackWebhook processes an incoming Paystack webhook event.
func (s *Service) ProcessPaystackWebhook(ctx context.Context, payload []byte, signature string) error {
	// Verify signature
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

	// Only process successful charge events
	if event.Event != "charge.success" {
		return nil // Ignore other events
	}

	eventID := fmt.Sprintf("paystack-%d", event.Data.ID)
	reference := event.Data.Reference

	claimed, err := s.store.RecordPaymentWebhook(ctx, "paystack", eventID, reference, signature, payload)
	if err != nil {
		return err
	}
	if !claimed {
		return errors.ErrDuplicateWebhook
	}

	// Verify payment with Paystack API
	verified, err := s.verifyPaystackTransaction(ctx, reference)
	if err != nil {
		return err
	}
	if !verified {
		return errors.ErrPaymentVerificationFailed
	}

	// Process the investment — amount from Paystack is in kobo
	if err := s.processSuccessfulPayment(ctx, "paystack", reference, event.Data.Amount/100, event.Data.Currency); err != nil {
		return err
	}

	// Mark webhook as processed
	return s.store.MarkPaymentWebhookProcessed(ctx, "paystack", eventID)
}

// ProcessFlutterwaveWebhook processes an incoming Flutterwave webhook event.
func (s *Service) ProcessFlutterwaveWebhook(ctx context.Context, payload []byte, signature string) error {
	// Verify signature
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

	// Only process successful charge events
	if event.Event != "charge.completed" && event.Event != "transfer.completed" {
		return nil
	}
	if event.Data.Status != "successful" {
		return nil
	}

	eventID := fmt.Sprintf("flutterwave-%d", event.Data.ID)
	reference := event.Data.Reference

	claimed, err := s.store.RecordPaymentWebhook(ctx, "flutterwave", eventID, reference, signature, payload)
	if err != nil {
		return err
	}
	if !claimed {
		return errors.ErrDuplicateWebhook
	}

	// Verify payment with Flutterwave API
	verified, err := s.verifyFlutterwaveTransaction(ctx, reference)
	if err != nil {
		return err
	}
	if !verified {
		return errors.ErrPaymentVerificationFailed
	}

	// Process the investment
	amount := event.Data.Amount
	if event.Data.ChargedAmount > 0 {
		amount = event.Data.ChargedAmount
	}
	if err := s.processSuccessfulPayment(ctx, "flutterwave", reference, amount, event.Data.Currency); err != nil {
		return err
	}

	// Mark webhook as processed
	return s.store.MarkPaymentWebhookProcessed(ctx, "flutterwave", eventID)
}

// ─── Payment Success Processing ───────────────────────────

// processSuccessfulPayment runs ONLY after Paystack/Flutterwave verification succeeds.
// It creates the investment (if not already present), activates it, credits wallet,
// and records transaction history. Cancelled/abandoned/failed payments never
// reach this path.
func (s *Service) processSuccessfulPayment(ctx context.Context, provider, reference string, amountPaid float64, currency string) error {
	s.logger.Info("payment verified — creating investment and crediting wallet",
		zap.String("provider", provider),
		zap.String("reference", reference),
		zap.Float64("amount", amountPaid),
		zap.String("currency", currency),
	)

	// Look up the payment transaction created at init time.
	pt, err := s.store.GetPaymentTransactionByReference(ctx, provider, reference)
	if err != nil {
		return err
	}
	if pt == nil {
		return errors.ErrPaymentVerificationFailed
	}

	// Idempotency: if already completed and investment exists, skip.
	if pt.Status == "completed" {
		if inv, _ := s.store.GetInvestmentByReference(ctx, provider, reference); inv != nil {
			if inv.Status != models.InvestmentStatusPending {
				return errors.ErrPaymentAlreadyProcessed
			}
		}
	}

	// Verify payment amount matches (allow for gateway fees).
	diff := math.Abs(pt.Amount - amountPaid)
	if amountPaid > 0 && diff > 100 {
		s.logger.Warn("payment amount mismatch",
			zap.String("reference", reference),
			zap.Float64("expected", pt.Amount),
			zap.Float64("received", amountPaid),
		)
		return errors.ErrPaymentVerificationFailed
	}

	now := time.Now().UTC()

	// Check if investment already exists (e.g. webhook and verify both fired).
	inv, err := s.store.GetInvestmentByReference(ctx, provider, reference)
	if err != nil {
		return err
	}

	if inv == nil {
		// Deferred investment creation — only now, after verified success.
		// Parse the params stored on the payment transaction at init time.
		params := models.InvestmentParams{}
		if len(pt.Response) > 0 {
			_ = json.Unmarshal(pt.Response, &params)
		}

		maturesAt := now.AddDate(0, 0, params.LockPeriodDays)
		inv = &models.Investment{
			ID:               uuidlib.NewString(),
			UserID:           pt.UserID,
			PlanID:           params.PlanID,
			PaymentProvider:  provider,
			PaymentReference: reference,
			PaymentStatus:    "completed",
			AmountPaid:       pt.Amount,
			Currency:         pt.Currency,
			ExchangeRate:     1.0,
			CDTPrice:         params.CDTPrice,
			AllocatedCDT:     params.AllocatedCDT,
			ROIPercent:       params.ROIPercent,
			ROICDT:           params.ROICDT,
			LockPeriodDays:   params.LockPeriodDays,
			Status:           models.InvestmentStatusActive,
			StartedAt:        &now,
			MaturesAt:        &maturesAt,
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		// Persisted atomically with payment completion and wallet credit below.
		s.logger.Info("Investment created after verified payment",
			zap.String("investment_id", inv.ID),
			zap.String("reference", reference),
		)
	} else {
		// Investment already exists (e.g. created by an earlier code path).
		// Activate it.
		if inv.Status != models.InvestmentStatusPending {
			return errors.ErrPaymentAlreadyProcessed
		}
		inv.PaymentStatus = "completed"
		inv.Status = models.InvestmentStatusActive
		inv.StartedAt = &now
		maturesAt := now.AddDate(0, 0, inv.LockPeriodDays)
		inv.MaturesAt = &maturesAt
		inv.UpdatedAt = now
		// Persisted atomically with payment completion and wallet credit below.
	}

	alreadyProcessed, err := s.store.FinalizeSuccessfulPayment(ctx, pt, inv)
	if err != nil {
		return err
	}
	if alreadyProcessed {
		return errors.ErrPaymentAlreadyProcessed
	}

	// Publish events.
	s.publish(events.EventDepositCompleted, map[string]interface{}{
		"user_id":       inv.UserID,
		"investment_id": inv.ID,
		"amount":        inv.AmountPaid,
		"allocated_cdt": inv.AllocatedCDT,
		"reference":     reference,
	})

	s.audit(ctx, inv.UserID, audit.ActionDeposit, "investment", inv.ID, map[string]interface{}{
		"provider": provider, "reference": reference, "amount": amountPaid, "cdt": inv.AllocatedCDT,
	})

	return nil
}

// ─── Investment Queries ───────────────────────────────

func (s *Service) GetUserInvestments(ctx context.Context, userID string, status string, page, perPage int) ([]*models.Investment, int, error) {
	return s.store.ListUserInvestments(ctx, userID, status, page, perPage)
}

func (s *Service) GetInvestment(ctx context.Context, userID, investmentID string) (*models.Investment, error) {
	inv, err := s.store.GetInvestmentByID(ctx, investmentID)
	if err != nil {
		return nil, err
	}
	if inv == nil || inv.UserID != userID {
		return nil, errors.ErrInvestmentNotFound
	}
	return inv, nil
}

func (s *Service) GetDashboard(ctx context.Context, userID string) (*models.InvestmentDashboard, error) {
	if !s.hasStore() {
		return &models.InvestmentDashboard{
			AvailableCDT: 0,
			LockedCDT:    0,
			Investments:  []*models.InvestmentSummary{},
		}, nil
	}

	investments, _, err := s.store.ListUserInvestments(ctx, userID, "", 1, 500)
	if err != nil {
		return nil, err
	}
	s.logger.Info("investments loaded", zap.String("user_id", userID), zap.Int("count", len(investments)))

	wallet, err := s.store.GetOrCreateWallet(ctx, userID, "CDT")
	if err != nil {
		return nil, err
	}

	dash := &models.InvestmentDashboard{
		AvailableCDT: wallet.AvailableBalance,
		LockedCDT:    wallet.LockedBalance,
	}

	var summaries []*models.InvestmentSummary
	now := time.Now().UTC()

	for _, inv := range investments {
		dash.TotalInvested += inv.AmountPaid
		dash.TotalROIEarned += inv.ROICDT

		switch inv.Status {
		case models.InvestmentStatusActive:
			dash.ActiveInvestments++
		case models.InvestmentStatusCompleted:
			dash.CompletedInvestments++
		}

		planName := ""
		if inv.Plan != nil {
			planName = inv.Plan.Name
		}
		summary := &models.InvestmentSummary{
			ID:             inv.ID,
			PlanName:       planName,
			AmountPaid:     inv.AmountPaid,
			AllocatedCDT:   inv.AllocatedCDT,
			ROICDT:         inv.ROICDT,
			ROIPercent:     inv.ROIPercent,
			Status:         inv.Status,
			LockPeriodDays: inv.LockPeriodDays,
			StartedAt:      inv.StartedAt,
			MaturesAt:      inv.MaturesAt,
			CreatedAt:      inv.CreatedAt,
		}

		if inv.MaturesAt != nil && inv.Status == models.InvestmentStatusActive {
			remaining := time.Until(*inv.MaturesAt)
			if remaining > 0 {
				summary.DaysRemaining = int(remaining.Hours() / 24)
			}
			totalDuration := inv.MaturesAt.Sub(*inv.StartedAt)
			elapsed := now.Sub(*inv.StartedAt)
			if totalDuration > 0 {
				summary.ProgressPct = math.Min((elapsed.Seconds()/totalDuration.Seconds())*100, 100)
			}

			// Track upcoming maturity
			if dash.UpcomingMaturity == nil || inv.MaturesAt.Before(*dash.UpcomingMaturity) {
				dash.UpcomingMaturity = inv.MaturesAt
			}
		}

		if inv.Status == models.InvestmentStatusCompleted {
			summary.ProgressPct = 100
		}

		summaries = append(summaries, summary)
	}

	dash.Investments = summaries
	return dash, nil
}

// ─── Wallet ───────────────────────────────────────────

func (s *Service) GetWallet(ctx context.Context, userID string) (*models.Wallet, error) {
	if !s.hasStore() {
		return &models.Wallet{UserID: userID, AvailableBalance: 0, LockedBalance: 0, TotalBalance: 0}, nil
	}
	wallet, err := s.store.GetOrCreateWallet(ctx, userID, "CDT")
	if err != nil {
		return nil, err
	}
	s.logger.Info("wallet loaded",
		zap.String("user_id", userID),
		zap.Float64("available", wallet.AvailableBalance),
		zap.Float64("locked", wallet.LockedBalance),
		zap.Float64("total", wallet.TotalBalance),
	)
	return wallet, nil
}

func (s *Service) GetWalletTransactions(ctx context.Context, userID string, page, perPage int) ([]*models.WalletTransaction, int, error) {
	if !s.hasStore() {
		return []*models.WalletTransaction{}, 0, nil
	}
	wallet, err := s.store.GetOrCreateWallet(ctx, userID, "CDT")
	if err != nil {
		return nil, 0, err
	}
	return s.store.ListWalletTransactions(ctx, wallet.ID, page, perPage)
}

// ─── Maturity Processing ──────────────────────────────

// ProcessMaturedInvestments checks for matured investments and processes them.
func (s *Service) ProcessMaturedInvestments(ctx context.Context) error {
	if !s.hasStore() {
		return nil
	}
	investments, err := s.store.ListMaturedInvestments(ctx, 500)
	if err != nil {
		return err
	}

	s.logger.Info("investment maturity check started", zap.Int("matured_count", len(investments)))

	for _, inv := range investments {
		if err := s.processMaturedInvestment(ctx, inv); err != nil {
			s.logger.Error("failed to process matured investment",
				zap.String("investment_id", inv.ID),
				zap.Error(err),
			)
			continue
		}
		s.logger.Info("investment maturity processed",
			zap.String("investment_id", inv.ID),
			zap.String("user_id", inv.UserID),
			zap.Float64("payout", inv.AllocatedCDT+inv.ROICDT),
			zap.Float64("roi", inv.ROICDT),
		)
	}
	return nil
}

func (s *Service) processMaturedInvestment(ctx context.Context, inv *models.Investment) error {
	now := time.Now().UTC()

	// Calculate ROI
	roiCDT := inv.AllocatedCDT * (inv.ROIPercent / 100.0)
	totalPayout := inv.AllocatedCDT + roiCDT

	// Update investment
	inv.Status = models.InvestmentStatusCompleted
	inv.CompletedAt = &now
	inv.ROICDT = roiCDT
	inv.UpdatedAt = now

	if err := s.store.UpdateInvestment(ctx, inv); err != nil {
		return err
	}

	// Get CDT wallet
	wallet, err := s.store.GetOrCreateWallet(ctx, inv.UserID, "CDT")
	if err != nil {
		return err
	}

	// Unlock locked balance and credit ROI
	if err := s.store.UnlockWalletBalance(ctx, wallet.ID, inv.AllocatedCDT); err != nil {
		return err
	}
	if err := s.store.CreditWalletAvailable(ctx, wallet.ID, roiCDT); err != nil {
		return err
	}

	// Record wallet transaction for unlock
	_ = s.store.CreateWalletTransaction(ctx, &models.WalletTransaction{
		ID:            uuidlib.NewString(),
		WalletID:      wallet.ID,
		Type:          models.WalletTxUnlock,
		Amount:        inv.AllocatedCDT,
		BalanceBefore: wallet.TotalBalance,
		BalanceAfter:  wallet.TotalBalance + inv.AllocatedCDT + roiCDT,
		Reference:     inv.ID,
		Description:   fmt.Sprintf("Investment matured - %s plan unlocked", planNameOrDefault(inv)),
		CreatedAt:     now,
	})

	// Record wallet transaction for ROI
	_ = s.store.CreateWalletTransaction(ctx, &models.WalletTransaction{
		ID:            uuidlib.NewString(),
		WalletID:      wallet.ID,
		Type:          models.WalletTxROI,
		Amount:        roiCDT,
		BalanceBefore: wallet.TotalBalance,
		BalanceAfter:  wallet.TotalBalance + roiCDT,
		Reference:     inv.ID,
		Description:   fmt.Sprintf("ROI credited - %s plan (%.2f%%)", planNameOrDefault(inv), inv.ROIPercent),
		CreatedAt:     now,
	})

	s.logger.Info("CDT credited",
		zap.String("user_id", inv.UserID),
		zap.String("investment_id", inv.ID),
		zap.Float64("unlocked_cdt", inv.AllocatedCDT),
		zap.Float64("roi_cdt", roiCDT),
		zap.Float64("total_payout", totalPayout),
	)

	s.publish(events.EventEarnParticipationCompleted, map[string]interface{}{
		"user_id":       inv.UserID,
		"investment_id": inv.ID,
		"payout":        totalPayout,
		"roi":           roiCDT,
	})

	s.audit(ctx, inv.UserID, audit.ActionEarnExit, "investment", inv.ID, map[string]interface{}{
		"status": "completed", "payout": totalPayout, "roi": roiCDT,
	})

	return nil
}

func planNameOrDefault(inv *models.Investment) string {
	if inv != nil && inv.Plan != nil && inv.Plan.Name != "" {
		return inv.Plan.Name
	}
	return "Genesis"
}

// ─── Admin ────────────────────────────────────────────

func (s *Service) AdminListInvestments(ctx context.Context, status string, page, perPage int) ([]*models.Investment, int, error) {
	return s.store.ListAllInvestments(ctx, status, page, perPage)
}

func (s *Service) AdminListPayments(ctx context.Context, status string, page, perPage int) ([]*models.PaymentTransaction, int, error) {
	return s.store.ListPaymentTransactions(ctx, status, page, perPage)
}

func (s *Service) AdminListWallets(ctx context.Context, page, perPage int) ([]*models.Wallet, int, error) {
	return s.store.ListWallets(ctx, page, perPage)
}

func (s *Service) AdminListWebhookEvents(ctx context.Context, provider, status string, page, perPage int) ([]map[string]interface{}, int, error) {
	return s.store.ListWebhookEvents(ctx, provider, status, page, perPage)
}

func (s *Service) AdminGetStats(ctx context.Context) (*models.AdminInvestmentStats, error) {
	stats := &models.AdminInvestmentStats{}

	totalInvested, _ := s.store.SumInvestmentsByStatus(ctx, "")
	stats.TotalInvested = totalInvested

	lockedCDT, _ := s.store.SumAllocatedCDTByStatus(ctx, "active")
	stats.TotalLockedCDT = lockedCDT

	completedCDT, _ := s.store.SumAllocatedCDTByStatus(ctx, "completed")
	stats.TotalAvailableCDT = completedCDT

	roiPaid, _ := s.store.SumAllocatedCDTByStatus(ctx, "completed")
	stats.TotalROIPaid = roiPaid

	active, _ := s.store.CountInvestmentsByStatus(ctx, "active")
	stats.ActiveInvestments = active

	completed, _ := s.store.CountInvestmentsByStatus(ctx, "completed")
	stats.CompletedInvestments = completed

	users, _ := s.store.CountDistinctUsersWithInvestments(ctx)
	stats.TotalUsers = users

	pending, _ := s.store.CountInvestmentsByStatus(ctx, "pending")
	stats.PendingPayments = pending

	failed, _ := s.store.CountInvestmentsByStatus(ctx, "failed")
	stats.FailedPayments = failed

	return stats, nil
}

// ─── Paystack API ─────────────────────────────────────

func (s *Service) paystackCallbackURL() string {
	if strings.TrimSpace(s.cfg.PaystackCallbackURL) != "" {
		return strings.TrimSpace(s.cfg.PaystackCallbackURL)
	}
	base := strings.TrimRight(s.cfg.BaseURL, "/")
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
			"payment_for": "genesis_investment",
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
	mac := hmac.New(sha512.New, []byte(secret))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

// ─── Flutterwave API ──────────────────────────────────

func (s *Service) callFlutterwaveInitialize(ctx context.Context, amount float64, currency, reference, userID string) (string, error) {
	if s.cfg.FlutterwaveSecretKey == "" {
		return "", errors.ErrGatewayNotConfigured
	}

	payload := map[string]interface{}{
		"tx_ref":       reference,
		"amount":       amount,
		"currency":     currency,
		"redirect_url": fmt.Sprintf("%s/earn/verify", s.cfg.BaseURL),
		"customer": map[string]string{
			"email": "", // Will be filled by caller
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

// ─── Workers ──────────────────────────────────────────

func (s *Service) registerWorkers() {
	if s.jobRegistry == nil {
		return
	}
	s.jobRegistry.Register("investment.maturity_check", func(ctx context.Context, job workers.Job) error {
		return s.ProcessMaturedInvestments(ctx)
	})
}

// ─── Helpers ──────────────────────────────────────────

func (s *Service) publish(eventType string, data map[string]interface{}) {
	if s.eventBus == nil {
		return
	}
	_ = s.eventBus.Publish(context.Background(), events.NewEvent(eventType, "investment-service", data))
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
