package service

import (
	"context"
	"testing"
	"time"
)

func TestDefaultInvestmentSettingsProvideSaneFallbacks(t *testing.T) {
	settings := defaultInvestmentSettings()
	if settings == nil {
		t.Fatal("expected default investment settings")
	}
	if settings.MinimumInvestmentUSD <= 0 {
		t.Fatalf("expected positive minimum investment, got %.2f", settings.MinimumInvestmentUSD)
	}
	if settings.DailyRewardNGN <= 0 {
		t.Fatalf("expected positive daily reward, got %.2f", settings.DailyRewardNGN)
	}
	if !settings.Enabled {
		t.Fatal("expected settings to be enabled by default")
	}
}

func TestDefaultExchangeRateProvideSaneFallbacks(t *testing.T) {
	rate := defaultExchangeRate()
	if rate == nil {
		t.Fatal("expected default exchange rate")
	}
	if rate.USDTNGN <= 0 {
		t.Fatalf("expected positive exchange rate, got %.2f", rate.USDTNGN)
	}
}

func TestGetSettingsReturnsDefaultsWhenStoreMissing(t *testing.T) {
	svc := &Service{}
	settings, err := svc.GetSettings(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if settings == nil {
		t.Fatal("expected settings fallback")
	}
	if settings.MinimumInvestmentUSD <= 0 {
		t.Fatalf("expected minimum investment fallback, got %.2f", settings.MinimumInvestmentUSD)
	}
}

func TestGetExchangeRateReturnsDefaultsWhenStoreMissing(t *testing.T) {
	svc := &Service{}
	rate, err := svc.GetExchangeRate(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rate == nil {
		t.Fatal("expected exchange rate fallback")
	}
	if rate.USDTNGN <= 0 {
		t.Fatalf("expected exchange rate fallback, got %.2f", rate.USDTNGN)
	}
}

// ─── Genesis Tier & ROI Ladder ───────────────────────────

func TestDefaultSettingsIncludeGenesisTier(t *testing.T) {
	settings := defaultInvestmentSettings()
	if settings.MinimumInvestmentUSD != 10 {
		t.Fatalf("expected Genesis minimum of $10, got %.2f", settings.MinimumInvestmentUSD)
	}
	if settings.ROIPercent != 18 {
		t.Fatalf("expected Genesis ROI of 18%%, got %.2f", settings.ROIPercent)
	}
}

func TestDefaultSettingsIncludeWeeklyWithdrawalInterval(t *testing.T) {
	settings := defaultInvestmentSettings()
	if settings.WithdrawalIntervalDays != 7 {
		t.Fatalf("expected 7-day withdrawal interval, got %d", settings.WithdrawalIntervalDays)
	}
	if settings.WithdrawalProcessingHours != 24 {
		t.Fatalf("expected 24-hour processing, got %d", settings.WithdrawalProcessingHours)
	}
}

func TestDefaultWithdrawalIntervalIsSevenDays(t *testing.T) {
	if defaultWithdrawalInterval != 7*24*time.Hour {
		t.Fatalf("expected 168h, got %v", defaultWithdrawalInterval)
	}
}

func TestDefaultSettingsDailyRewardMatchesGenesis(t *testing.T) {
	settings := defaultInvestmentSettings()
	// $10 × ₦1,400 × 18% / 20 business days = ₦126/day
	if settings.DailyRewardNGN != 126 {
		t.Fatalf("expected Genesis daily reward of 126, got %.2f", settings.DailyRewardNGN)
	}
	if settings.MaxBusinessDays != 20 {
		t.Fatalf("expected 20 business days, got %d", settings.MaxBusinessDays)
	}
}

func TestDefaultExchangeRateIsFourteenHundred(t *testing.T) {
	rate := defaultExchangeRate()
	if rate.USDTNGN != 1400 {
		t.Fatalf("expected default exchange rate of 1400, got %.2f", rate.USDTNGN)
	}
}
