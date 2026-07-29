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

// Config holds the investment service configuration.
type Config struct {
	BaseURL               string
	PaystackSecretKey     string
	PaystackPublicKey     string
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
	return s.store.ListPlans(ctx, onlyEnabled)
}

func (s *Service) GetPlan(ctx context.Context, id string) (*models.InvestmentPlan, error) {
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
	plan, pricing, err := s.validateInvestmentRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	allocatedCDT := req.Amount / pricing.PriceNGN
	reference := fmt.Sprintf("CDT-PS-%s-%d", uuidlib.NewString()[:8], time.Now().Unix())

	// Create pending payment transaction
	pt := &models.PaymentTransaction{
		ID:        uuidlib.NewString(),
		UserID:    userID,
		Provider:  "paystack",
		Reference: reference,
		Status:    "pending",
		Amount:    req.Amount,
		Currency:  req.Currency,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.store.CreatePaymentTransaction(ctx, pt); err != nil {
		return nil, err
	}

	// Call Paystack initialize API
	authURL, accessCode, err := s.callPaystackInitialize(ctx, req.Amount, req.Currency, reference, userID)
	if err != nil {
		return nil, err
	}

	// Create pending investment record
	now := time.Now().UTC()
	maturesAt := now.AddDate(0, 0, req.LockPeriodDays)
	roiCDT := allocatedCDT * (plan.ROIPercent / 100.0)

	inv := &models.Investment{
		ID:               uuidlib.NewString(),
		UserID:           userID,
		PlanID:           plan.ID,
		PaymentProvider:  "paystack",
		PaymentReference: reference,
		PaymentStatus:    "pending",
		AmountPaid:       req.Amount,
		Currency:         req.Currency,
		ExchangeRate:     1.0,
		CDTPrice:         pricing.PriceNGN,
		AllocatedCDT:     allocatedCDT,
		ROIPercent:       plan.ROIPercent,
		ROICDT:           roiCDT,
		LockPeriodDays:   req.LockPeriodDays,
		Status:           models.InvestmentStatusPending,
		StartedAt:        nil,
		MaturesAt:        &maturesAt,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := s.store.CreateInvestment(ctx, inv); err != nil {
		return nil, err
	}

	s.audit(ctx, userID, audit.ActionDeposit, "investment", inv.ID, map[string]interface{}{
		"provider": "paystack", "reference": reference, "amount": req.Amount,
	})

	return &models.InitPaymentResponse{
		AuthorizationURL: authURL,
		Reference:        reference,
		AccessCode:       accessCode,
	}, nil
}

func (s *Service) InitFlutterwavePayment(ctx context.Context, userID string, req *models.InitPaymentRequest) (*models.InitPaymentResponse, error) {
	plan, pricing, err := s.validateInvestmentRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	allocatedCDT := req.Amount / pricing.PriceNGN
	reference := fmt.Sprintf("CDT-FW-%s-%d", uuidlib.NewString()[:8], time.Now().Unix())

	// Create pending payment transaction
	pt := &models.PaymentTransaction{
		ID:        uuidlib.NewString(),
		UserID:    userID,
		Provider:  "flutterwave",
		Reference: reference,
		Status:    "pending",
		Amount:    req.Amount,
		Currency:  req.Currency,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.store.CreatePaymentTransaction(ctx, pt); err != nil {
		return nil, err
	}

	// Call Flutterwave initialize API
	authURL, err := s.callFlutterwaveInitialize(ctx, req.Amount, req.Currency, reference, userID)
	if err != nil {
		return nil, err
	}

	// Create pending investment record
	now := time.Now().UTC()
	maturesAt := now.AddDate(0, 0, req.LockPeriodDays)
	roiCDT := allocatedCDT * (plan.ROIPercent / 100.0)

	inv := &models.Investment{
		ID:               uuidlib.NewString(),
		UserID:           userID,
		PlanID:           plan.ID,
		PaymentProvider:  "flutterwave",
		PaymentReference: reference,
		PaymentStatus:    "pending",
		AmountPaid:       req.Amount,
		Currency:         req.Currency,
		ExchangeRate:     1.0,
		CDTPrice:         pricing.PriceNGN,
		AllocatedCDT:     allocatedCDT,
		ROIPercent:       plan.ROIPercent,
		ROICDT:           roiCDT,
		LockPeriodDays:   req.LockPeriodDays,
		Status:           models.InvestmentStatusPending,
		StartedAt:        nil,
		MaturesAt:        &maturesAt,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := s.store.CreateInvestment(ctx, inv); err != nil {
		return nil, err
	}

	s.audit(ctx, userID, audit.ActionDeposit, "investment", inv.ID, map[string]interface{}{
		"provider": "flutterwave", "reference": reference, "amount": req.Amount,
	})

	return &models.InitPaymentResponse{
		AuthorizationURL: authURL,
		Reference:        reference,
	}, nil
}

func (s *Service) validateInvestmentRequest(ctx context.Context, req *models.InitPaymentRequest) (*models.InvestmentPlan, *models.Pricing, error) {
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

	// Deduplicate webhook
	processed, err := s.store.IsWebhookProcessed(ctx, "paystack", eventID)
	if err != nil {
		return err
	}
	if processed {
		return errors.ErrDuplicateWebhook
	}

	// Record webhook event
	if err := s.store.CreateWebhookEvent(ctx, "paystack", eventID, reference, payload); err != nil {
		return err
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
	if err := s.processSuccessfulPayment(ctx, "paystack", reference, event.Data.Amount/100, event.Data.Currency); err != nil {
		return err
	}

	// Mark webhook as processed
	return s.store.MarkWebhookProcessed(ctx, "paystack", eventID)
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

	// Deduplicate webhook
	processed, err := s.store.IsWebhookProcessed(ctx, "flutterwave", eventID)
	if err != nil {
		return err
	}
	if processed {
		return errors.ErrDuplicateWebhook
	}

	// Record webhook event
	if err := s.store.CreateWebhookEvent(ctx, "flutterwave", eventID, reference, payload); err != nil {
		return err
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
	return s.store.MarkWebhookProcessed(ctx, "flutterwave", eventID)
}

func (s *Service) processSuccessfulPayment(ctx context.Context, provider, reference string, amountPaid float64, currency string) error {
	s.logger.Info("payment verified",
		zap.String("provider", provider),
		zap.String("reference", reference),
		zap.Float64("amount", amountPaid),
		zap.String("currency", currency),
	)

	// Get the pending investment
	inv, err := s.store.GetInvestmentByReference(ctx, provider, reference)
	if err != nil {
		return err
	}
	if inv == nil {
		return errors.ErrInvestmentNotFound
	}

	// Prevent duplicate processing
	if inv.Status != models.InvestmentStatusPending {
		return errors.ErrPaymentAlreadyProcessed
	}

	// Verify payment amount matches (allow small difference for gateway fees)
	diff := math.Abs(inv.AmountPaid - amountPaid)
	if diff > 100 { // Allow up to 100 currency units difference
		s.logger.Warn("payment amount mismatch",
			zap.String("reference", reference),
			zap.Float64("expected", inv.AmountPaid),
			zap.Float64("received", amountPaid),
		)
		return errors.ErrPaymentVerificationFailed
	}

	now := time.Now().UTC()

	// Update payment transaction
	pt, err := s.store.GetPaymentTransactionByReference(ctx, provider, reference)
	if err == nil && pt != nil {
		_ = s.store.UpdatePaymentTransaction(ctx, pt.ID, "completed", &now)
	}

	// Update investment to active
	inv.PaymentStatus = "completed"
	inv.Status = models.InvestmentStatusActive
	inv.StartedAt = &now

	// Recalculate matures_at from now
	maturesAt := now.AddDate(0, 0, inv.LockPeriodDays)
	inv.MaturesAt = &maturesAt
	inv.UpdatedAt = now

	if err := s.store.UpdateInvestment(ctx, inv); err != nil {
		return err
	}

	// Get or create wallet
	wallet, err := s.store.GetOrCreateWallet(ctx, inv.UserID)
	if err != nil {
		return err
	}

	s.logger.Info("CDT credited",
		zap.String("user_id", inv.UserID),
		zap.Float64("amount", inv.AllocatedCDT),
		zap.String("reference", reference),
		zap.String("type", "locked"),
	)

	// Credit locked CDT to wallet
	if err := s.store.CreditWalletLocked(ctx, wallet.ID, inv.AllocatedCDT); err != nil {
		return err
	}

	// Record wallet transaction
	wt := &models.WalletTransaction{
		ID:            uuidlib.NewString(),
		WalletID:      wallet.ID,
		Type:          models.WalletTxInvestment,
		Amount:        inv.AllocatedCDT,
		BalanceBefore: wallet.TotalBalance,
		BalanceAfter:  wallet.TotalBalance + inv.AllocatedCDT,
		Reference:     reference,
		Description:   fmt.Sprintf("Investment in %s plan - %d CDT locked until %s", inv.Plan.Name, int(inv.AllocatedCDT), inv.MaturesAt.Format("2006-01-02")),
		CreatedAt:     now,
	}
	_ = s.store.CreateWalletTransaction(ctx, wt)

	// Publish events
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
	investments, _, err := s.store.ListUserInvestments(ctx, userID, "", 1, 500)
	if err != nil {
		return nil, err
	}
	s.logger.Info("investments loaded", zap.String("user_id", userID), zap.Int("count", len(investments)))

	wallet, err := s.store.GetOrCreateWallet(ctx, userID)
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
	wallet, err := s.store.GetOrCreateWallet(ctx, userID)
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
	wallet, err := s.store.GetOrCreateWallet(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	return s.store.ListWalletTransactions(ctx, wallet.ID, page, perPage)
}

// ─── Maturity Processing ──────────────────────────────

// ProcessMaturedInvestments checks for matured investments and processes them.
func (s *Service) ProcessMaturedInvestments(ctx context.Context) error {
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

	// Get wallet
	wallet, err := s.store.GetOrCreateWallet(ctx, inv.UserID)
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

func (s *Service) callPaystackInitialize(ctx context.Context, amount float64, currency, reference, userID string) (string, string, error) {
	if s.cfg.PaystackSecretKey == "" {
		return "", "", errors.ErrGatewayNotConfigured
	}

	amountInKobo := int64(amount * 100)
	payload := map[string]interface{}{
		"amount":       amountInKobo,
		"currency":     currency,
		"reference":    reference,
		"callback_url": fmt.Sprintf("%s/earn/verify", s.cfg.BaseURL),
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
