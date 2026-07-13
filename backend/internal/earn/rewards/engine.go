package rewards

import (
	"context"
	"fmt"
	"time"

	"github.com/coindistro/backend/internal/earn/models"
)

// Calculator computes estimated or period rewards without executing custody.
// Strategies are pluggable so calculation models can evolve independently.
type Calculator interface {
	// Name returns the strategy identifier (matches product reward_model).
	Name() string
	// Estimate returns an estimated reward amount for display.
	Estimate(ctx context.Context, product *models.Product, participation *models.Participation, asOf time.Time) (float64, error)
	// Accrue returns a reward amount for a period (e.g. daily job).
	// Implementations must not mutate domain state; the service applies results.
	Accrue(ctx context.Context, product *models.Product, participation *models.Participation, periodStart, periodEnd time.Time) (float64, string, error)
}

// Engine dispatches to registered calculators.
type Engine struct {
	calculators map[string]Calculator
	fallback    Calculator
}

// NewEngine creates a reward engine with built-in strategies.
func NewEngine() *Engine {
	e := &Engine{
		calculators: make(map[string]Calculator),
		fallback:    &FlexibleCalculator{},
	}
	e.Register(&FlexibleCalculator{})
	e.Register(&FixedCalculator{})
	e.Register(&PromotionalCalculator{})
	e.Register(&EducationalCalculator{})
	e.Register(&ReferralCalculator{})
	return e
}

// Register adds a calculator strategy.
func (e *Engine) Register(c Calculator) {
	e.calculators[c.Name()] = c
}

// Get returns a calculator by reward model name.
func (e *Engine) Get(model string) Calculator {
	if c, ok := e.calculators[model]; ok {
		return c
	}
	return e.fallback
}

// Estimate delegates to the product's reward model strategy.
func (e *Engine) Estimate(ctx context.Context, product *models.Product, participation *models.Participation, asOf time.Time) (float64, error) {
	return e.Get(product.RewardModel).Estimate(ctx, product, participation, asOf)
}

// Accrue delegates accrual for a period.
func (e *Engine) Accrue(ctx context.Context, product *models.Product, participation *models.Participation, periodStart, periodEnd time.Time) (float64, string, error) {
	return e.Get(product.RewardModel).Accrue(ctx, product, participation, periodStart, periodEnd)
}

// dailyRateFromAPR converts annual percentage rate to a simple daily rate.
// This is display/estimate only — not financial product execution.
func dailyRateFromAPR(apr float64) float64 {
	if apr <= 0 {
		return 0
	}
	return apr / 100.0 / 365.0
}

// FlexibleCalculator: continuous display accrual estimate.
type FlexibleCalculator struct{}

func (c *FlexibleCalculator) Name() string { return "flexible" }

func (c *FlexibleCalculator) Estimate(_ context.Context, product *models.Product, p *models.Participation, asOf time.Time) (float64, error) {
	days := asOf.Sub(p.JoinedAt).Hours() / 24.0
	if days < 0 {
		days = 0
	}
	return p.CurrentBalance * dailyRateFromAPR(product.RewardAPR) * days, nil
}

func (c *FlexibleCalculator) Accrue(_ context.Context, product *models.Product, p *models.Participation, start, end time.Time) (float64, string, error) {
	hours := end.Sub(start).Hours()
	if hours <= 0 {
		return 0, "daily", nil
	}
	days := hours / 24.0
	return p.CurrentBalance * dailyRateFromAPR(product.RewardAPR) * days, "daily", nil
}

// FixedCalculator: pro-rata estimate over fixed lock period.
type FixedCalculator struct{}

func (c *FixedCalculator) Name() string { return "fixed" }

func (c *FixedCalculator) Estimate(_ context.Context, product *models.Product, p *models.Participation, asOf time.Time) (float64, error) {
	duration := 30
	if product.DurationDays != nil && *product.DurationDays > 0 {
		duration = *product.DurationDays
	}
	totalReward := p.AllocatedAmount * (product.RewardAPR / 100.0) * (float64(duration) / 365.0)
	if p.LockStartAt == nil || p.LockEndAt == nil {
		return totalReward, nil
	}
	total := p.LockEndAt.Sub(*p.LockStartAt).Seconds()
	if total <= 0 {
		return totalReward, nil
	}
	elapsed := asOf.Sub(*p.LockStartAt).Seconds()
	if elapsed < 0 {
		elapsed = 0
	}
	if elapsed > total {
		elapsed = total
	}
	return totalReward * (elapsed / total), nil
}

func (c *FixedCalculator) Accrue(_ context.Context, product *models.Product, p *models.Participation, start, end time.Time) (float64, string, error) {
	// Fixed products typically grant at maturity; intermediate accrual is estimate-only zero.
	if p.LockEndAt != nil && !end.Before(*p.LockEndAt) && (p.Status == models.ParticipationLocked || p.Status == models.ParticipationActive) {
		duration := 30
		if product.DurationDays != nil && *product.DurationDays > 0 {
			duration = *product.DurationDays
		}
		amt := p.AllocatedAmount * (product.RewardAPR / 100.0) * (float64(duration) / 365.0)
		return amt, "fixed_maturity", nil
	}
	return 0, "fixed_maturity", nil
}

// PromotionalCalculator: flat or APR-based promo.
type PromotionalCalculator struct{}

func (c *PromotionalCalculator) Name() string { return "promotional" }

func (c *PromotionalCalculator) Estimate(ctx context.Context, product *models.Product, p *models.Participation, asOf time.Time) (float64, error) {
	return (&FlexibleCalculator{}).Estimate(ctx, product, p, asOf)
}

func (c *PromotionalCalculator) Accrue(ctx context.Context, product *models.Product, p *models.Participation, start, end time.Time) (float64, string, error) {
	amt, _, err := (&FlexibleCalculator{}).Accrue(ctx, product, p, start, end)
	return amt, "promotional", err
}

// EducationalCalculator: fixed grant amounts (Learn & Earn).
type EducationalCalculator struct{}

func (c *EducationalCalculator) Name() string { return "educational" }

func (c *EducationalCalculator) Estimate(_ context.Context, product *models.Product, _ *models.Participation, _ time.Time) (float64, error) {
	if product.Metadata != nil {
		if v, ok := product.Metadata["reward_amount"].(float64); ok {
			return v, nil
		}
	}
	return 0, nil
}

func (c *EducationalCalculator) Accrue(_ context.Context, product *models.Product, _ *models.Participation, _, _ time.Time) (float64, string, error) {
	amt, err := c.Estimate(context.Background(), product, nil, time.Now())
	return amt, "educational", err
}

// ReferralCalculator: milestone-based (amount from external claim records).
type ReferralCalculator struct{}

func (c *ReferralCalculator) Name() string { return "referral" }

func (c *ReferralCalculator) Estimate(_ context.Context, _ *models.Product, p *models.Participation, _ time.Time) (float64, error) {
	return p.EstimatedRewards, nil
}

func (c *ReferralCalculator) Accrue(_ context.Context, _ *models.Product, _ *models.Participation, _, _ time.Time) (float64, string, error) {
	return 0, "referral", nil
}

// ValidateFixedDuration checks allowed fixed earn durations.
func ValidateFixedDuration(days int) error {
	for _, d := range models.FixedDurations {
		if d == days {
			return nil
		}
	}
	return fmt.Errorf("duration must be one of %v", models.FixedDurations)
}

// ValidateStrategyProfile checks AI strategy profiles.
func ValidateStrategyProfile(profile string) error {
	switch profile {
	case models.StrategyConservative, models.StrategyBalanced, models.StrategyGrowth, models.StrategyAggressive, "":
		return nil
	default:
		return fmt.Errorf("invalid strategy profile: %s", profile)
	}
}
