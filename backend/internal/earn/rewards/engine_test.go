package rewards_test

import (
	"context"
	"testing"
	"time"

	"github.com/coindistro/backend/internal/earn/models"
	"github.com/coindistro/backend/internal/earn/rewards"
)

func TestFlexibleEstimate(t *testing.T) {
	engine := rewards.NewEngine()
	product := &models.Product{RewardModel: "flexible", RewardAPR: 36.5} // 0.1% daily
	joined := time.Now().UTC().Add(-10 * 24 * time.Hour)
	part := &models.Participation{
		CurrentBalance: 1000,
		JoinedAt:       joined,
	}
	est, err := engine.Estimate(context.Background(), product, part, time.Now().UTC())
	if err != nil {
		t.Fatalf("estimate: %v", err)
	}
	if est <= 0 {
		t.Fatalf("expected positive estimate, got %v", est)
	}
	// roughly 1000 * 0.001 * 10 = 10
	if est < 5 || est > 15 {
		t.Fatalf("estimate out of expected range: %v", est)
	}
}

func TestFixedDurationValidation(t *testing.T) {
	if err := rewards.ValidateFixedDuration(30); err != nil {
		t.Fatal(err)
	}
	if err := rewards.ValidateFixedDuration(45); err == nil {
		t.Fatal("expected invalid duration")
	}
}

func TestStrategyProfileValidation(t *testing.T) {
	if err := rewards.ValidateStrategyProfile("balanced"); err != nil {
		t.Fatal(err)
	}
	if err := rewards.ValidateStrategyProfile("yolo"); err == nil {
		t.Fatal("expected invalid strategy")
	}
}

func TestFixedAccrueAtMaturity(t *testing.T) {
	engine := rewards.NewEngine()
	days := 30
	product := &models.Product{RewardModel: "fixed", RewardAPR: 12, DurationDays: &days}
	start := time.Now().UTC().Add(-31 * 24 * time.Hour)
	end := time.Now().UTC()
	part := &models.Participation{
		AllocatedAmount: 1000,
		CurrentBalance:  1000,
		Status:          models.ParticipationLocked,
		LockStartAt:     &start,
		LockEndAt:       &end,
	}
	amt, rtype, err := engine.Accrue(context.Background(), product, part, start, end.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if rtype != "fixed_maturity" {
		t.Fatalf("type: %s", rtype)
	}
	if amt <= 0 {
		t.Fatalf("expected maturity reward, got %v", amt)
	}
}

func TestEngineFallback(t *testing.T) {
	engine := rewards.NewEngine()
	c := engine.Get("unknown-model")
	if c.Name() != "flexible" {
		t.Fatalf("expected flexible fallback, got %s", c.Name())
	}
}
