package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/coindistro/backend/internal/audit"
	"github.com/coindistro/backend/internal/earn/errors"
	"github.com/coindistro/backend/internal/earn/models"
	"github.com/coindistro/backend/internal/earn/rewards"
	"github.com/coindistro/backend/internal/earn/store"
	"github.com/coindistro/backend/internal/events"
	"github.com/coindistro/backend/internal/featureflags"
	"github.com/coindistro/backend/internal/metrics"
	uuidlib "github.com/coindistro/backend/internal/uuid"
	"github.com/coindistro/backend/internal/workers"
)

// Service implements Earn business logic.
type Service struct {
	store        *store.Store
	engine       *rewards.Engine
	eventBus     *events.InMemoryBus
	jobRegistry  *workers.Registry
	workerPool   *workers.Pool
	featureFlags *featureflags.Manager
	auditLogger  *audit.Logger
	promMetrics  *metrics.Metrics
	logger       *zap.Logger
}

// New creates the Earn service.
func New(
	st *store.Store,
	eventBus *events.InMemoryBus,
	jobRegistry *workers.Registry,
	workerPool *workers.Pool,
	featureFlags *featureflags.Manager,
	auditLogger *audit.Logger,
	promMetrics *metrics.Metrics,
	logger *zap.Logger,
) *Service {
	svc := &Service{
		store:        st,
		engine:       rewards.NewEngine(),
		eventBus:     eventBus,
		jobRegistry:  jobRegistry,
		workerPool:   workerPool,
		featureFlags: featureFlags,
		auditLogger:  auditLogger,
		promMetrics:  promMetrics,
		logger:       logger,
	}
	svc.registerWorkers()
	return svc
}

func (s *Service) ensureEnabled() error {
	if s.featureFlags != nil && !s.featureFlags.IsEnabled(featureflags.FlagEarn) {
		return errors.ErrEarnDisabled
	}
	return nil
}

func (s *Service) ensureCategory(category string) error {
	if s.featureFlags == nil {
		return nil
	}
	flag := categoryFlag(category)
	if flag != "" && !s.featureFlags.IsEnabled(flag) {
		return errors.ErrCategoryDisabled
	}
	return nil
}

func categoryFlag(category string) string {
	switch category {
	case models.CategoryFlexible:
		return featureflags.FlagEarnFlexible
	case models.CategoryFixed:
		return featureflags.FlagEarnFixed
	case models.CategoryStablecoin:
		return featureflags.FlagEarnStablecoin
	case models.CategoryAISmart:
		return featureflags.FlagEarnAI
	case models.CategorySignalVault:
		return featureflags.FlagEarnSignalVault
	case models.CategoryLaunchpool:
		return featureflags.FlagEarnLaunchpool
	case models.CategoryLearnEarn:
		return featureflags.FlagEarnLearn
	case models.CategoryReferral:
		return featureflags.FlagEarnReferral
	default:
		return ""
	}
}

// ─── Products ─────────────────────────────────────────

func (s *Service) ListProducts(ctx context.Context, f models.ProductListFilter) ([]*models.Product, int, error) {
	if err := s.ensureEnabled(); err != nil {
		return nil, 0, err
	}
	return s.store.ListProducts(ctx, f)
}

func (s *Service) GetProduct(ctx context.Context, idOrSlug string) (*models.Product, error) {
	if err := s.ensureEnabled(); err != nil {
		return nil, err
	}
	p, err := s.store.GetProductByID(ctx, idOrSlug)
	if err != nil {
		return nil, err
	}
	if p == nil {
		p, err = s.store.GetProductBySlug(ctx, idOrSlug)
		if err != nil {
			return nil, err
		}
	}
	if p == nil {
		return nil, errors.ErrProductNotFound
	}
	return p, nil
}

func (s *Service) CreateProduct(ctx context.Context, req *models.CreateProductRequest, actorID string) (*models.Product, error) {
	if err := s.ensureEnabled(); err != nil {
		return nil, err
	}
	if err := validateCategory(req.Category); err != nil {
		return nil, err
	}
	if req.Category == models.CategoryFixed && req.DurationDays != nil {
		if err := rewards.ValidateFixedDuration(*req.DurationDays); err != nil {
			return nil, errors.ErrInvalidDuration
		}
	}
	existing, _ := s.store.GetProductBySlug(ctx, req.Slug)
	if existing != nil {
		return nil, errors.ErrSlugExists
	}

	now := time.Now().UTC()
	status := req.Status
	if status == "" {
		status = models.StatusDraft
	}
	risk := req.RiskLevel
	if risk == "" {
		risk = "medium"
	}
	rewardModel := req.RewardModel
	if rewardModel == "" {
		rewardModel = defaultRewardModel(req.Category)
	}

	p := &models.Product{
		ID:               uuidlib.NewString(),
		Name:             req.Name,
		Slug:             strings.ToLower(req.Slug),
		Description:      req.Description,
		Category:         req.Category,
		SupportedAssets:  req.SupportedAssets,
		DurationDays:     req.DurationDays,
		CapacityTotal:    req.CapacityTotal,
		Status:           status,
		RiskLevel:        risk,
		MinAllocation:    req.MinAllocation,
		MaxAllocation:    req.MaxAllocation,
		RewardModel:      rewardModel,
		RewardAPR:        req.RewardAPR,
		Eligibility:      req.Eligibility,
		Rules:            req.Rules,
		StrategyProfiles: req.StrategyProfiles,
		Featured:         req.Featured,
		Metadata:         req.Metadata,
		StartsAt:         req.StartsAt,
		EndsAt:           req.EndsAt,
		CreatedBy:        &actorID,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if p.Eligibility == nil {
		p.Eligibility = map[string]interface{}{}
	}
	if p.Rules == nil {
		p.Rules = map[string]interface{}{}
	}
	if p.Metadata == nil {
		p.Metadata = map[string]interface{}{}
	}

	if err := s.store.CreateProduct(ctx, p); err != nil {
		return nil, err
	}
	s.publish(events.EventEarnProductCreated, map[string]interface{}{"product_id": p.ID, "slug": p.Slug, "category": p.Category})
	s.audit(ctx, actorID, audit.ActionEarnProductCreated, audit.EntityEarnProduct, p.ID, nil)
	s.incMetric("product_created")
	return p, nil
}

func (s *Service) UpdateProduct(ctx context.Context, id string, req *models.UpdateProductRequest, actorID string) (*models.Product, error) {
	p, err := s.GetProduct(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		p.Name = *req.Name
	}
	if req.Description != nil {
		p.Description = *req.Description
	}
	if req.SupportedAssets != nil {
		p.SupportedAssets = req.SupportedAssets
	}
	if req.DurationDays != nil {
		p.DurationDays = req.DurationDays
	}
	if req.CapacityTotal != nil {
		p.CapacityTotal = req.CapacityTotal
	}
	if req.RiskLevel != nil {
		p.RiskLevel = *req.RiskLevel
	}
	if req.MinAllocation != nil {
		p.MinAllocation = *req.MinAllocation
	}
	if req.MaxAllocation != nil {
		p.MaxAllocation = req.MaxAllocation
	}
	if req.RewardModel != nil {
		p.RewardModel = *req.RewardModel
	}
	if req.RewardAPR != nil {
		p.RewardAPR = *req.RewardAPR
	}
	if req.Eligibility != nil {
		p.Eligibility = req.Eligibility
	}
	if req.Rules != nil {
		p.Rules = req.Rules
	}
	if req.StrategyProfiles != nil {
		p.StrategyProfiles = req.StrategyProfiles
	}
	if req.Featured != nil {
		p.Featured = *req.Featured
	}
	if req.Metadata != nil {
		p.Metadata = req.Metadata
	}
	if req.StartsAt != nil {
		p.StartsAt = req.StartsAt
	}
	if req.EndsAt != nil {
		p.EndsAt = req.EndsAt
	}
	if req.Status != nil {
		p.Status = *req.Status
	}
	p.UpdatedAt = time.Now().UTC()
	if err := s.store.UpdateProduct(ctx, p); err != nil {
		return nil, err
	}
	s.publish(events.EventEarnProductUpdated, map[string]interface{}{"product_id": p.ID, "status": p.Status})
	s.audit(ctx, actorID, audit.ActionEarnProductUpdated, audit.EntityEarnProduct, p.ID, nil)
	return p, nil
}

func (s *Service) SetProductStatus(ctx context.Context, id, status, actorID string) (*models.Product, error) {
	return s.UpdateProduct(ctx, id, &models.UpdateProductRequest{Status: &status}, actorID)
}

// ─── Participation ────────────────────────────────────

func (s *Service) JoinProduct(ctx context.Context, userID string, productID string, req *models.JoinProductRequest) (*models.Participation, error) {
	if err := s.ensureEnabled(); err != nil {
		return nil, err
	}
	product, err := s.GetProduct(ctx, productID)
	if err != nil {
		return nil, err
	}
	if err := s.ensureCategory(product.Category); err != nil {
		return nil, err
	}
	if product.Status != models.StatusActive {
		return nil, errors.ErrProductNotActive
	}
	if !assetSupported(product, req.Asset) {
		return nil, errors.ErrInvalidAsset
	}
	if req.Amount < product.MinAllocation {
		return nil, errors.ErrInvalidAllocation
	}
	if product.MaxAllocation != nil && req.Amount > *product.MaxAllocation {
		return nil, errors.ErrInvalidAllocation
	}
	if product.CapacityTotal != nil && product.CapacityUsed+req.Amount > *product.CapacityTotal {
		return nil, errors.ErrCapacityExceeded
	}
	if product.Category == models.CategoryAISmart {
		if err := rewards.ValidateStrategyProfile(req.StrategyProfile); err != nil {
			return nil, errors.ErrInvalidStrategy
		}
	}

	// Allow multiple flexible positions; single active for fixed-like products
	if product.Category == models.CategoryFixed || product.Category == models.CategoryStablecoin {
		existing, _ := s.store.GetUserProductParticipation(ctx, userID, product.ID)
		if existing != nil {
			return nil, errors.ErrAlreadyParticipating
		}
	}

	now := time.Now().UTC()
	status := models.ParticipationActive
	var lockStart, lockEnd *time.Time
	if product.Category == models.CategoryFixed && product.DurationDays != nil {
		status = models.ParticipationLocked
		ls := now
		le := now.AddDate(0, 0, *product.DurationDays)
		lockStart, lockEnd = &ls, &le
	}

	var profile *string
	if req.StrategyProfile != "" {
		profile = &req.StrategyProfile
	}

	part := &models.Participation{
		ID:              uuidlib.NewString(),
		UserID:          userID,
		ProductID:       product.ID,
		Asset:           strings.ToUpper(req.Asset),
		AllocatedAmount: req.Amount,
		CurrentBalance:  req.Amount,
		Status:          status,
		StrategyProfile: profile,
		JoinedAt:        now,
		LockStartAt:     lockStart,
		LockEndAt:       lockEnd,
		Metadata:        map[string]interface{}{},
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	est, _ := s.engine.Estimate(ctx, product, part, now)
	part.EstimatedRewards = est

	if err := s.store.CreateParticipation(ctx, part); err != nil {
		return nil, err
	}
	_ = s.store.IncrementCapacity(ctx, product.ID, req.Amount)
	pid := product.ID
	_ = s.store.CreateTransaction(ctx, &models.Transaction{
		ID: uuidlib.NewString(), UserID: userID, ProductID: &pid, ParticipationID: &part.ID,
		Type: "join", Asset: part.Asset, Amount: req.Amount, BalanceAfter: &part.CurrentBalance,
		Status: "completed", Description: "Joined earn product", CreatedAt: now,
	})

	s.publish(events.EventEarnParticipationCreated, map[string]interface{}{
		"participation_id": part.ID, "user_id": userID, "product_id": product.ID, "amount": req.Amount,
	})
	s.audit(ctx, userID, audit.ActionEarnJoin, audit.EntityEarnParticipation, part.ID, map[string]interface{}{"amount": req.Amount})
	s.incMetric("participation_created")
	s.enqueue(workers.JobEarnParticipationReminder, map[string]interface{}{"participation_id": part.ID, "user_id": userID})
	return part, nil
}

func (s *Service) AddFunds(ctx context.Context, userID, participationID string, amount float64) (*models.Participation, error) {
	part, product, err := s.getOwnedParticipation(ctx, userID, participationID)
	if err != nil {
		return nil, err
	}
	if product.Category != models.CategoryFlexible && product.Category != models.CategoryStablecoin && product.Category != models.CategoryAISmart {
		return nil, errors.ErrExitNotAllowed
	}
	if part.Status != models.ParticipationActive {
		return nil, errors.ErrExitNotAllowed
	}
	if product.CapacityTotal != nil && product.CapacityUsed+amount > *product.CapacityTotal {
		return nil, errors.ErrCapacityExceeded
	}
	if product.MaxAllocation != nil && part.AllocatedAmount+amount > *product.MaxAllocation {
		return nil, errors.ErrInvalidAllocation
	}

	part.AllocatedAmount += amount
	part.CurrentBalance += amount
	part.UpdatedAt = time.Now().UTC()
	est, _ := s.engine.Estimate(ctx, product, part, time.Now().UTC())
	part.EstimatedRewards = est
	if err := s.store.UpdateParticipation(ctx, part); err != nil {
		return nil, err
	}
	_ = s.store.IncrementCapacity(ctx, product.ID, amount)
	pid, partID := product.ID, part.ID
	_ = s.store.CreateTransaction(ctx, &models.Transaction{
		ID: uuidlib.NewString(), UserID: userID, ProductID: &pid, ParticipationID: &partID,
		Type: "add_funds", Asset: part.Asset, Amount: amount, BalanceAfter: &part.CurrentBalance,
		Status: "completed", Description: "Added funds", CreatedAt: time.Now().UTC(),
	})
	s.audit(ctx, userID, audit.ActionEarnAddFunds, audit.EntityEarnParticipation, part.ID, map[string]interface{}{"amount": amount})
	return part, nil
}

func (s *Service) Withdraw(ctx context.Context, userID, participationID string, amount float64) (*models.Participation, error) {
	part, product, err := s.getOwnedParticipation(ctx, userID, participationID)
	if err != nil {
		return nil, err
	}
	if product.Category == models.CategoryFixed && part.Status == models.ParticipationLocked {
		return nil, errors.ErrExitNotAllowed
	}
	if amount > part.CurrentBalance {
		return nil, errors.ErrInvalidAllocation
	}
	// Check rules.allow_partial_withdraw default true for flexible
	if rulesAllow, ok := product.Rules["allow_withdraw"].(bool); ok && !rulesAllow {
		return nil, errors.ErrExitNotAllowed
	}

	part.CurrentBalance -= amount
	part.AllocatedAmount -= amount
	if part.AllocatedAmount < 0 {
		part.AllocatedAmount = 0
	}
	part.UpdatedAt = time.Now().UTC()
	if err := s.store.UpdateParticipation(ctx, part); err != nil {
		return nil, err
	}
	_ = s.store.IncrementCapacity(ctx, product.ID, -amount)
	pid, partID := product.ID, part.ID
	_ = s.store.CreateTransaction(ctx, &models.Transaction{
		ID: uuidlib.NewString(), UserID: userID, ProductID: &pid, ParticipationID: &partID,
		Type: "withdraw", Asset: part.Asset, Amount: amount, BalanceAfter: &part.CurrentBalance,
		Status: "completed", Description: "Withdrew funds", CreatedAt: time.Now().UTC(),
	})
	s.audit(ctx, userID, audit.ActionEarnWithdraw, audit.EntityEarnParticipation, part.ID, map[string]interface{}{"amount": amount})
	return part, nil
}

func (s *Service) ExitParticipation(ctx context.Context, userID, participationID, reason string) (*models.Participation, error) {
	part, product, err := s.getOwnedParticipation(ctx, userID, participationID)
	if err != nil {
		return nil, err
	}
	if part.Status == models.ParticipationExited || part.Status == models.ParticipationCompleted {
		return nil, errors.ErrExitNotAllowed
	}
	if product.Category == models.CategoryFixed && part.Status == models.ParticipationLocked {
		if part.LockEndAt != nil && time.Now().UTC().Before(*part.LockEndAt) {
			if early, ok := product.Rules["allow_early_exit"].(bool); !ok || !early {
				return nil, errors.ErrExitNotAllowed
			}
		}
	}

	now := time.Now().UTC()
	part.Status = models.ParticipationExited
	part.ExitedAt = &now
	part.UpdatedAt = now
	if err := s.store.UpdateParticipation(ctx, part); err != nil {
		return nil, err
	}
	_ = s.store.IncrementCapacity(ctx, product.ID, -part.CurrentBalance)
	pid, partID := product.ID, part.ID
	_ = s.store.CreateTransaction(ctx, &models.Transaction{
		ID: uuidlib.NewString(), UserID: userID, ProductID: &pid, ParticipationID: &partID,
		Type: "exit", Asset: part.Asset, Amount: part.CurrentBalance, BalanceAfter: floatPtr(0),
		Status: "completed", Description: reason, CreatedAt: now,
	})
	s.publish(events.EventEarnParticipationExited, map[string]interface{}{
		"participation_id": part.ID, "user_id": userID, "product_id": product.ID,
	})
	s.audit(ctx, userID, audit.ActionEarnExit, audit.EntityEarnParticipation, part.ID, map[string]interface{}{"reason": reason})
	return part, nil
}

func (s *Service) GetParticipation(ctx context.Context, userID, id string) (*models.Participation, error) {
	part, product, err := s.getOwnedParticipation(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	part.Product = product
	est, _ := s.engine.Estimate(ctx, product, part, time.Now().UTC())
	part.EstimatedRewards = est
	return part, nil
}

func (s *Service) ListParticipations(ctx context.Context, userID, status string, page, perPage int) ([]*models.Participation, int, error) {
	if err := s.ensureEnabled(); err != nil {
		return nil, 0, err
	}
	return s.store.ListUserParticipations(ctx, userID, status, page, perPage)
}

func (s *Service) getOwnedParticipation(ctx context.Context, userID, id string) (*models.Participation, *models.Product, error) {
	if err := s.ensureEnabled(); err != nil {
		return nil, nil, err
	}
	part, err := s.store.GetParticipationByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if part == nil || part.UserID != userID {
		return nil, nil, errors.ErrParticipationNotFound
	}
	product, err := s.store.GetProductByID(ctx, part.ProductID)
	if err != nil || product == nil {
		return nil, nil, errors.ErrProductNotFound
	}
	return part, product, nil
}

// ─── Portfolio / history ──────────────────────────────

func (s *Service) PortfolioOverview(ctx context.Context, userID string) (*models.PortfolioOverview, error) {
	if err := s.ensureEnabled(); err != nil {
		return nil, err
	}
	return s.store.GetPortfolioOverview(ctx, userID)
}

func (s *Service) ListRewards(ctx context.Context, userID string, page, perPage int) ([]*models.Reward, int, error) {
	if err := s.ensureEnabled(); err != nil {
		return nil, 0, err
	}
	return s.store.ListUserRewards(ctx, userID, page, perPage)
}

func (s *Service) ListTransactions(ctx context.Context, userID string, page, perPage int) ([]*models.Transaction, int, error) {
	if err := s.ensureEnabled(); err != nil {
		return nil, 0, err
	}
	return s.store.ListUserTransactions(ctx, userID, page, perPage)
}

// ─── Admin analytics / campaigns ──────────────────────

func (s *Service) ProductAnalytics(ctx context.Context, productID string) (*models.ProductAnalytics, error) {
	return s.store.GetProductAnalytics(ctx, productID)
}

func (s *Service) ListParticipants(ctx context.Context, productID string, page, perPage int) ([]*models.Participation, int, error) {
	return s.store.ListProductParticipants(ctx, productID, page, perPage)
}

func (s *Service) CreateLaunchpool(ctx context.Context, req *models.CreateLaunchpoolRequest, actorID string) (*models.LaunchpoolCampaign, error) {
	if err := s.ensureEnabled(); err != nil {
		return nil, err
	}
	if err := s.ensureCategory(models.CategoryLaunchpool); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	status := req.Status
	if status == "" {
		status = models.StatusDraft
	}
	c := &models.LaunchpoolCampaign{
		ID: uuidlib.NewString(), ProductID: req.ProductID, Name: req.Name, Description: req.Description,
		SupportedAssets: req.SupportedAssets, WindowStart: req.WindowStart, WindowEnd: req.WindowEnd,
		AllocationRules: req.AllocationRules, RewardDistribution: req.RewardDistribution,
		Status: status, CreatedAt: now, UpdatedAt: now,
	}
	if c.AllocationRules == nil {
		c.AllocationRules = map[string]interface{}{}
	}
	if c.RewardDistribution == nil {
		c.RewardDistribution = map[string]interface{}{}
	}
	if err := s.store.CreateLaunchpool(ctx, c); err != nil {
		return nil, err
	}
	s.publish(events.EventLaunchpoolCreated, map[string]interface{}{"campaign_id": c.ID, "name": c.Name})
	s.audit(ctx, actorID, audit.ActionEarnLaunchpoolCreated, audit.EntityEarnLaunchpool, c.ID, nil)
	return c, nil
}

func (s *Service) ListLaunchpools(ctx context.Context, status string) ([]*models.LaunchpoolCampaign, error) {
	if err := s.ensureEnabled(); err != nil {
		return nil, err
	}
	return s.store.ListLaunchpools(ctx, status)
}

func (s *Service) CreateLearnCampaign(ctx context.Context, req *models.CreateLearnCampaignRequest, actorID string) (*models.LearnCampaign, error) {
	if err := s.ensureEnabled(); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	status := req.Status
	if status == "" {
		status = models.StatusDraft
	}
	asset := req.RewardAsset
	if asset == "" {
		asset = "USDT"
	}
	c := &models.LearnCampaign{
		ID: uuidlib.NewString(), ProductID: req.ProductID, Name: req.Name, Description: req.Description,
		AcademyCourseID: req.AcademyCourseID, RewardAsset: asset, RewardAmount: req.RewardAmount,
		Status: status, StartsAt: req.StartsAt, EndsAt: req.EndsAt, CreatedAt: now, UpdatedAt: now,
	}
	if err := s.store.CreateLearnCampaign(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Service) ListLearnCampaigns(ctx context.Context, status string) ([]*models.LearnCampaign, error) {
	if err := s.ensureEnabled(); err != nil {
		return nil, err
	}
	if status == "" {
		status = models.StatusActive
	}
	return s.store.ListLearnCampaigns(ctx, status)
}

func (s *Service) CompleteLearnCampaign(ctx context.Context, userID, campaignID string, meta map[string]interface{}) (*models.LearnCompletion, error) {
	if err := s.ensureEnabled(); err != nil {
		return nil, err
	}
	if err := s.ensureCategory(models.CategoryLearnEarn); err != nil {
		return nil, err
	}
	campaign, err := s.store.GetLearnCampaign(ctx, campaignID)
	if err != nil || campaign == nil {
		return nil, errors.ErrCampaignNotFound
	}
	existing, _ := s.store.GetLearnCompletion(ctx, campaignID, userID)
	if existing != nil {
		return nil, errors.ErrAlreadyCompleted
	}
	now := time.Now().UTC()
	comp := &models.LearnCompletion{
		ID: uuidlib.NewString(), CampaignID: campaignID, UserID: userID,
		CompletedAt: now, RewardEligible: true, RewardGranted: false,
		Metadata: meta, CreatedAt: now,
	}
	if comp.Metadata == nil {
		comp.Metadata = map[string]interface{}{}
	}
	if err := s.store.CreateLearnCompletion(ctx, comp); err != nil {
		return nil, err
	}

	// Record reward eligibility (not custody transfer)
	rewardID := uuidlib.NewString()
	var productID string
	if campaign.ProductID != nil {
		productID = *campaign.ProductID
	} else {
		productID = campaignID // fallback reference
	}
	partID := ""
	r := &models.Reward{
		ID: rewardID, UserID: userID, ProductID: productID, Asset: campaign.RewardAsset,
		Amount: campaign.RewardAmount, RewardType: "educational", Status: "calculated",
		Description: "Learn & Earn completion eligibility", CreatedAt: now, UpdatedAt: now,
		Metadata: map[string]interface{}{"campaign_id": campaignID},
	}
	if productID != "" {
		_ = s.store.CreateReward(ctx, r)
		comp.RewardID = &rewardID
		_ = partID
	}
	s.publish(events.EventLearnRewardGranted, map[string]interface{}{
		"user_id": userID, "campaign_id": campaignID, "reward_id": rewardID, "amount": campaign.RewardAmount,
	})
	s.audit(ctx, userID, audit.ActionEarnLearnComplete, audit.EntityEarnLearn, campaignID, nil)
	return comp, nil
}

func (s *Service) ReferralRewardSummary(ctx context.Context, userID string) (*models.ReferralRewardSummary, error) {
	if err := s.ensureEnabled(); err != nil {
		return nil, err
	}
	milestones, _ := s.store.ListReferralMilestones(ctx)
	claims, _ := s.store.ListReferralClaims(ctx, userID)
	summary := &models.ReferralRewardSummary{Milestones: milestones, Claims: claims}
	for _, c := range claims {
		if c.Status == "eligible" {
			summary.TotalEligible += c.Amount
		}
		if c.Status == "granted" {
			summary.TotalGranted += c.Amount
		}
	}
	return summary, nil
}

// ─── Reward engine jobs ───────────────────────────────

func (s *Service) RunDailyRewardCalculations(ctx context.Context) error {
	parts, err := s.store.ListActiveParticipations(ctx, 1000)
	if err != nil {
		return err
	}
	end := time.Now().UTC()
	start := end.Add(-24 * time.Hour)
	for _, part := range parts {
		product, err := s.store.GetProductByID(ctx, part.ProductID)
		if err != nil || product == nil {
			continue
		}
		amt, rtype, err := s.engine.Accrue(ctx, product, part, start, end)
		if err != nil || amt <= 0 {
			continue
		}
		now := time.Now().UTC()
		partID := part.ID
		r := &models.Reward{
			ID: uuidlib.NewString(), UserID: part.UserID, ProductID: part.ProductID, ParticipationID: &partID,
			Asset: part.Asset, Amount: amt, RewardType: rtype, Status: "calculated",
			PeriodStart: &start, PeriodEnd: &end, Description: "Scheduled reward calculation",
			CreatedAt: now, UpdatedAt: now, Metadata: map[string]interface{}{},
		}
		if err := s.store.CreateReward(ctx, r); err != nil {
			s.incMetric("reward_failed")
			continue
		}
		part.AccruedRewards += amt
		part.LifetimeRewards += amt
		part.EstimatedRewards, _ = s.engine.Estimate(ctx, product, part, now)
		_ = s.store.UpdateParticipation(ctx, part)
		s.publish(events.EventRewardCalculated, map[string]interface{}{
			"reward_id": r.ID, "user_id": part.UserID, "amount": amt, "type": rtype,
		})
		s.incMetric("reward_calculated")
	}
	return nil
}

func (s *Service) RunLifecycleUpdates(ctx context.Context) error {
	parts, err := s.store.ListActiveParticipations(ctx, 1000)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	for _, part := range parts {
		if part.Status == models.ParticipationLocked && part.LockEndAt != nil && !now.Before(*part.LockEndAt) {
			part.Status = models.ParticipationCompleted
			part.CompletedAt = &now
			part.UpdatedAt = now
			_ = s.store.UpdateParticipation(ctx, part)
			s.publish(events.EventEarnParticipationCompleted, map[string]interface{}{
				"participation_id": part.ID, "user_id": part.UserID,
			})
			s.enqueue(workers.JobEarnCompletionNotify, map[string]interface{}{
				"participation_id": part.ID, "user_id": part.UserID,
			})
		}
	}
	// Close expired products
	products, _ := s.store.ListProductsByStatus(ctx, models.StatusActive)
	for _, p := range products {
		if p.EndsAt != nil && now.After(*p.EndsAt) {
			st := models.StatusClosed
			_, _ = s.UpdateProduct(ctx, p.ID, &models.UpdateProductRequest{Status: &st}, "system")
		}
	}
	return nil
}

func (s *Service) RunPerformanceSnapshots(ctx context.Context) error {
	products, err := s.store.ListProductsByStatus(ctx, models.StatusActive)
	if err != nil {
		return err
	}
	day := time.Now().UTC().Truncate(24 * time.Hour)
	for _, p := range products {
		a, err := s.store.GetProductAnalytics(ctx, p.ID)
		if err != nil {
			continue
		}
		_ = s.store.CreatePerformanceSnapshot(ctx, p.ID, day, a.Participants, a.TotalAllocated, a.TotalRewards, a.CapacityUsedPct)
	}
	return nil
}

func (s *Service) RefreshMetrics(ctx context.Context) {
	if s.promMetrics == nil {
		return
	}
	if n, err := s.store.CountActiveProducts(ctx); err == nil {
		s.promMetrics.EarnActiveProducts.Set(float64(n))
	}
	if n, err := s.store.CountActiveParticipants(ctx); err == nil {
		s.promMetrics.EarnActiveParticipants.Set(float64(n))
	}
}

// ─── helpers ──────────────────────────────────────────

func (s *Service) registerWorkers() {
	if s.jobRegistry == nil {
		return
	}
	s.jobRegistry.Register(workers.JobEarnRewardCalculate, func(ctx context.Context, job workers.Job) error {
		return s.RunDailyRewardCalculations(ctx)
	})
	s.jobRegistry.Register(workers.JobEarnParticipationReminder, func(ctx context.Context, job workers.Job) error {
		s.logger.Info("earn participation reminder queued", zap.Any("payload", job.Payload))
		return nil
	})
	s.jobRegistry.Register(workers.JobEarnCompletionNotify, func(ctx context.Context, job workers.Job) error {
		s.logger.Info("earn completion notification queued", zap.Any("payload", job.Payload))
		return nil
	})
	s.jobRegistry.Register(workers.JobEarnPromoCampaign, func(ctx context.Context, job workers.Job) error {
		s.logger.Info("earn promo campaign job", zap.Any("payload", job.Payload))
		return nil
	})
}

func (s *Service) enqueue(jobType string, payload map[string]interface{}) {
	if s.workerPool == nil {
		return
	}
	s.workerPool.Submit(workers.Job{
		ID: uuidlib.NewString(), Type: jobType, Payload: payload, MaxRetries: 3,
	})
}

func (s *Service) publish(eventType string, data map[string]interface{}) {
	if s.eventBus == nil {
		return
	}
	_ = s.eventBus.Publish(context.Background(), events.NewEvent(eventType, "earn-service", data))
}

func (s *Service) audit(ctx context.Context, actorID string, action audit.Action, entityType audit.EntityType, entityID string, meta map[string]interface{}) {
	if s.auditLogger == nil {
		return
	}
	ev := audit.NewEvent(actorID, action).
		WithUserID(actorID).
		WithEntity(entityType, entityID).
		WithOutcome("success")
	if meta != nil {
		ev = ev.WithMetadata(meta)
	}
	_ = s.auditLogger.Record(ctx, ev.Build())
}

func (s *Service) incMetric(name string) {
	if s.promMetrics == nil {
		return
	}
	switch name {
	case "product_created":
		// tracked via gauges refresh
	case "participation_created":
		s.promMetrics.EarnParticipationsTotal.Inc()
	case "reward_calculated":
		s.promMetrics.EarnRewardCalculations.Inc()
	case "reward_failed":
		s.promMetrics.EarnFailedOperations.WithLabelValues("reward_calculate").Inc()
	}
}

func validateCategory(c string) error {
	switch c {
	case models.CategoryFlexible, models.CategoryFixed, models.CategoryStablecoin,
		models.CategoryAISmart, models.CategorySignalVault, models.CategoryLaunchpool,
		models.CategoryLearnEarn, models.CategoryReferral:
		return nil
	default:
		return fmt.Errorf("%w: %s", errors.ErrInvalidAllocation, "invalid category")
	}
}

func defaultRewardModel(category string) string {
	switch category {
	case models.CategoryFixed:
		return "fixed"
	case models.CategoryLearnEarn:
		return "educational"
	case models.CategoryReferral:
		return "referral"
	default:
		return "flexible"
	}
}

func assetSupported(p *models.Product, asset string) bool {
	a := strings.ToUpper(asset)
	for _, s := range p.SupportedAssets {
		if strings.ToUpper(s) == a {
			return true
		}
	}
	return false
}

func floatPtr(v float64) *float64 { return &v }
