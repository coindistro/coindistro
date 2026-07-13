package models

import (
	"time"
)

// Product categories.
const (
	CategoryFlexible    = "flexible"
	CategoryFixed       = "fixed"
	CategoryStablecoin  = "stablecoin"
	CategoryAISmart     = "ai_smart"
	CategorySignalVault = "signal_vault"
	CategoryLaunchpool  = "launchpool"
	CategoryLearnEarn   = "learn_earn"
	CategoryReferral    = "referral"
)

// Product statuses.
const (
	StatusDraft    = "draft"
	StatusActive   = "active"
	StatusPaused   = "paused"
	StatusClosed   = "closed"
	StatusArchived = "archived"
)

// Participation statuses.
const (
	ParticipationActive    = "active"
	ParticipationLocked    = "locked"
	ParticipationCompleted = "completed"
	ParticipationExited    = "exited"
	ParticipationCancelled = "cancelled"
)

// Fixed earn durations (days).
var FixedDurations = []int{30, 60, 90, 180, 365}

// AI strategy profiles.
const (
	StrategyConservative = "conservative"
	StrategyBalanced     = "balanced"
	StrategyGrowth       = "growth"
	StrategyAggressive   = "aggressive"
)

// Product is an earn product definition.
type Product struct {
	ID               string                 `json:"id" db:"id"`
	Name             string                 `json:"name" db:"name"`
	Slug             string                 `json:"slug" db:"slug"`
	Description      string                 `json:"description" db:"description"`
	Category         string                 `json:"category" db:"category"`
	SupportedAssets  []string               `json:"supported_assets" db:"supported_assets"`
	DurationDays     *int                   `json:"duration_days,omitempty" db:"duration_days"`
	CapacityTotal    *float64               `json:"capacity_total,omitempty" db:"capacity_total"`
	CapacityUsed     float64                `json:"capacity_used" db:"capacity_used"`
	Status           string                 `json:"status" db:"status"`
	RiskLevel        string                 `json:"risk_level" db:"risk_level"`
	MinAllocation    float64                `json:"min_allocation" db:"min_allocation"`
	MaxAllocation    *float64               `json:"max_allocation,omitempty" db:"max_allocation"`
	RewardModel      string                 `json:"reward_model" db:"reward_model"`
	RewardAPR        float64                `json:"reward_apr" db:"reward_apr"`
	Eligibility      map[string]interface{} `json:"eligibility" db:"eligibility"`
	Rules            map[string]interface{} `json:"rules" db:"rules"`
	StrategyProfiles []string               `json:"strategy_profiles,omitempty" db:"strategy_profiles"`
	Featured         bool                   `json:"featured" db:"featured"`
	Metadata         map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	StartsAt         *time.Time             `json:"starts_at,omitempty" db:"starts_at"`
	EndsAt           *time.Time             `json:"ends_at,omitempty" db:"ends_at"`
	CreatedBy        *string                `json:"created_by,omitempty" db:"created_by"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at" db:"updated_at"`
}

// Participation is a user allocation in a product.
type Participation struct {
	ID               string                 `json:"id" db:"id"`
	UserID           string                 `json:"user_id" db:"user_id"`
	ProductID        string                 `json:"product_id" db:"product_id"`
	Asset            string                 `json:"asset" db:"asset"`
	AllocatedAmount  float64                `json:"allocated_amount" db:"allocated_amount"`
	CurrentBalance   float64                `json:"current_balance" db:"current_balance"`
	EstimatedRewards float64                `json:"estimated_rewards" db:"estimated_rewards"`
	AccruedRewards   float64                `json:"accrued_rewards" db:"accrued_rewards"`
	LifetimeRewards  float64                `json:"lifetime_rewards" db:"lifetime_rewards"`
	Status           string                 `json:"status" db:"status"`
	StrategyProfile  *string                `json:"strategy_profile,omitempty" db:"strategy_profile"`
	JoinedAt         time.Time              `json:"joined_at" db:"joined_at"`
	LockStartAt      *time.Time             `json:"lock_start_at,omitempty" db:"lock_start_at"`
	LockEndAt        *time.Time             `json:"lock_end_at,omitempty" db:"lock_end_at"`
	CompletedAt      *time.Time             `json:"completed_at,omitempty" db:"completed_at"`
	ExitedAt         *time.Time             `json:"exited_at,omitempty" db:"exited_at"`
	Metadata         map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at" db:"updated_at"`

	// Joined fields (optional)
	Product *Product `json:"product,omitempty"`
}

// Reward is a reward ledger entry.
type Reward struct {
	ID              string                 `json:"id" db:"id"`
	UserID          string                 `json:"user_id" db:"user_id"`
	ProductID       string                 `json:"product_id" db:"product_id"`
	ParticipationID *string                `json:"participation_id,omitempty" db:"participation_id"`
	Asset           string                 `json:"asset" db:"asset"`
	Amount          float64                `json:"amount" db:"amount"`
	RewardType      string                 `json:"reward_type" db:"reward_type"`
	Status          string                 `json:"status" db:"status"`
	Description     string                 `json:"description,omitempty" db:"description"`
	PeriodStart     *time.Time             `json:"period_start,omitempty" db:"period_start"`
	PeriodEnd       *time.Time             `json:"period_end,omitempty" db:"period_end"`
	GrantedAt       *time.Time             `json:"granted_at,omitempty" db:"granted_at"`
	Metadata        map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`
}

// Transaction is an earn activity record.
type Transaction struct {
	ID              string                 `json:"id" db:"id"`
	UserID          string                 `json:"user_id" db:"user_id"`
	ProductID       *string                `json:"product_id,omitempty" db:"product_id"`
	ParticipationID *string                `json:"participation_id,omitempty" db:"participation_id"`
	Type            string                 `json:"type" db:"type"`
	Asset           string                 `json:"asset" db:"asset"`
	Amount          float64                `json:"amount" db:"amount"`
	BalanceAfter    *float64               `json:"balance_after,omitempty" db:"balance_after"`
	Status          string                 `json:"status" db:"status"`
	Reference       *string                `json:"reference,omitempty" db:"reference"`
	Description     string                 `json:"description,omitempty" db:"description"`
	Metadata        map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
}

// LaunchpoolCampaign models a launch campaign.
type LaunchpoolCampaign struct {
	ID                 string                 `json:"id" db:"id"`
	ProductID          *string                `json:"product_id,omitempty" db:"product_id"`
	Name               string                 `json:"name" db:"name"`
	Description        string                 `json:"description" db:"description"`
	SupportedAssets    []string               `json:"supported_assets" db:"supported_assets"`
	WindowStart        time.Time              `json:"window_start" db:"window_start"`
	WindowEnd          time.Time              `json:"window_end" db:"window_end"`
	AllocationRules    map[string]interface{} `json:"allocation_rules" db:"allocation_rules"`
	RewardDistribution map[string]interface{} `json:"reward_distribution" db:"reward_distribution"`
	Status             string                 `json:"status" db:"status"`
	CreatedAt          time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at" db:"updated_at"`
}

// LearnCampaign models a learn & earn campaign.
type LearnCampaign struct {
	ID              string     `json:"id" db:"id"`
	ProductID       *string    `json:"product_id,omitempty" db:"product_id"`
	Name            string     `json:"name" db:"name"`
	Description     string     `json:"description" db:"description"`
	AcademyCourseID *string    `json:"academy_course_id,omitempty" db:"academy_course_id"`
	RewardAsset     string     `json:"reward_asset" db:"reward_asset"`
	RewardAmount    float64    `json:"reward_amount" db:"reward_amount"`
	Status          string     `json:"status" db:"status"`
	StartsAt        *time.Time `json:"starts_at,omitempty" db:"starts_at"`
	EndsAt          *time.Time `json:"ends_at,omitempty" db:"ends_at"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

// LearnCompletion tracks learning completion for reward eligibility.
type LearnCompletion struct {
	ID             string                 `json:"id" db:"id"`
	CampaignID     string                 `json:"campaign_id" db:"campaign_id"`
	UserID         string                 `json:"user_id" db:"user_id"`
	CompletedAt    time.Time              `json:"completed_at" db:"completed_at"`
	RewardEligible bool                   `json:"reward_eligible" db:"reward_eligible"`
	RewardGranted  bool                   `json:"reward_granted" db:"reward_granted"`
	RewardID       *string                `json:"reward_id,omitempty" db:"reward_id"`
	Metadata       map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt      time.Time              `json:"created_at" db:"created_at"`
}

// ReferralMilestone defines a referral reward milestone.
type ReferralMilestone struct {
	ID                string                 `json:"id" db:"id"`
	Name              string                 `json:"name" db:"name"`
	Description       string                 `json:"description" db:"description"`
	RequiredReferrals int                    `json:"required_referrals" db:"required_referrals"`
	RewardAsset       string                 `json:"reward_asset" db:"reward_asset"`
	RewardAmount      float64                `json:"reward_amount" db:"reward_amount"`
	Status            string                 `json:"status" db:"status"`
	Metadata          map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt         time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" db:"updated_at"`
}

// ReferralRewardClaim records referral reward eligibility/grant.
type ReferralRewardClaim struct {
	ID          string                 `json:"id" db:"id"`
	UserID      string                 `json:"user_id" db:"user_id"`
	MilestoneID *string                `json:"milestone_id,omitempty" db:"milestone_id"`
	Amount      float64                `json:"amount" db:"amount"`
	Asset       string                 `json:"asset" db:"asset"`
	Status      string                 `json:"status" db:"status"`
	GrantedAt   *time.Time             `json:"granted_at,omitempty" db:"granted_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
}

// ─── Request DTOs ─────────────────────────────────────

type CreateProductRequest struct {
	Name             string                 `json:"name" binding:"required,min=2,max=255"`
	Slug             string                 `json:"slug" binding:"required,min=2,max=255"`
	Description      string                 `json:"description"`
	Category         string                 `json:"category" binding:"required"`
	SupportedAssets  []string               `json:"supported_assets" binding:"required,min=1"`
	DurationDays     *int                   `json:"duration_days"`
	CapacityTotal    *float64               `json:"capacity_total"`
	RiskLevel        string                 `json:"risk_level"`
	MinAllocation    float64                `json:"min_allocation"`
	MaxAllocation    *float64               `json:"max_allocation"`
	RewardModel      string                 `json:"reward_model"`
	RewardAPR        float64                `json:"reward_apr"`
	Eligibility      map[string]interface{} `json:"eligibility"`
	Rules            map[string]interface{} `json:"rules"`
	StrategyProfiles []string               `json:"strategy_profiles"`
	Featured         bool                   `json:"featured"`
	Metadata         map[string]interface{} `json:"metadata"`
	StartsAt         *time.Time             `json:"starts_at"`
	EndsAt           *time.Time             `json:"ends_at"`
	Status           string                 `json:"status"`
}

type UpdateProductRequest struct {
	Name             *string                `json:"name"`
	Description      *string                `json:"description"`
	SupportedAssets  []string               `json:"supported_assets"`
	DurationDays     *int                   `json:"duration_days"`
	CapacityTotal    *float64               `json:"capacity_total"`
	RiskLevel        *string                `json:"risk_level"`
	MinAllocation    *float64               `json:"min_allocation"`
	MaxAllocation    *float64               `json:"max_allocation"`
	RewardModel      *string                `json:"reward_model"`
	RewardAPR        *float64               `json:"reward_apr"`
	Eligibility      map[string]interface{} `json:"eligibility"`
	Rules            map[string]interface{} `json:"rules"`
	StrategyProfiles []string               `json:"strategy_profiles"`
	Featured         *bool                  `json:"featured"`
	Metadata         map[string]interface{} `json:"metadata"`
	StartsAt         *time.Time             `json:"starts_at"`
	EndsAt           *time.Time             `json:"ends_at"`
	Status           *string                `json:"status"`
}

type JoinProductRequest struct {
	Asset           string  `json:"asset" binding:"required"`
	Amount          float64 `json:"amount" binding:"required,gt=0"`
	StrategyProfile string  `json:"strategy_profile"`
}

type AddFundsRequest struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

type WithdrawRequest struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

type ExitParticipationRequest struct {
	Reason string `json:"reason"`
}

type CompleteLearnRequest struct {
	// Optional metadata from Academy client
	Metadata map[string]interface{} `json:"metadata"`
}

type CreateLaunchpoolRequest struct {
	ProductID          *string                `json:"product_id"`
	Name               string                 `json:"name" binding:"required"`
	Description        string                 `json:"description"`
	SupportedAssets    []string               `json:"supported_assets" binding:"required,min=1"`
	WindowStart        time.Time              `json:"window_start" binding:"required"`
	WindowEnd          time.Time              `json:"window_end" binding:"required"`
	AllocationRules    map[string]interface{} `json:"allocation_rules"`
	RewardDistribution map[string]interface{} `json:"reward_distribution"`
	Status             string                 `json:"status"`
}

type CreateLearnCampaignRequest struct {
	ProductID       *string    `json:"product_id"`
	Name            string     `json:"name" binding:"required"`
	Description     string     `json:"description"`
	AcademyCourseID *string    `json:"academy_course_id"`
	RewardAsset     string     `json:"reward_asset"`
	RewardAmount    float64    `json:"reward_amount"`
	Status          string     `json:"status"`
	StartsAt        *time.Time `json:"starts_at"`
	EndsAt          *time.Time `json:"ends_at"`
}

// ─── Response DTOs ────────────────────────────────────

type PortfolioOverview struct {
	TotalAssetsInEarn   float64            `json:"total_assets_in_earn"`
	EstimatedRewards    float64            `json:"estimated_rewards"`
	TodaysRewards       float64            `json:"todays_rewards"`
	LifetimeRewards     float64            `json:"lifetime_rewards"`
	ActiveProducts      int                `json:"active_products"`
	AvailableBalance    float64            `json:"available_balance"` // future wallet
	LockedBalance       float64            `json:"locked_balance"`
	AllocationByProduct map[string]float64 `json:"allocation_by_product"`
	AllocationByAsset   map[string]float64 `json:"allocation_by_asset"`
}

type ProductAnalytics struct {
	ProductID       string   `json:"product_id"`
	Participants    int      `json:"participants"`
	TotalAllocated  float64  `json:"total_allocated"`
	CapacityUsed    float64  `json:"capacity_used"`
	CapacityTotal   *float64 `json:"capacity_total,omitempty"`
	CapacityUsedPct float64  `json:"capacity_used_pct"`
	TotalRewards    float64  `json:"total_rewards"`
	AvgAllocation   float64  `json:"avg_allocation"`
}

type ReferralRewardSummary struct {
	TotalEligible float64                `json:"total_eligible"`
	TotalGranted  float64                `json:"total_granted"`
	Claims        []*ReferralRewardClaim `json:"claims"`
	Milestones    []*ReferralMilestone   `json:"milestones"`
}

type ProductListFilter struct {
	Category string
	Status   string
	Featured *bool
	Asset    string
	Page     int
	PerPage  int
}
