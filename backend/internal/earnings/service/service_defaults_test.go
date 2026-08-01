package service

import (
	"context"
	"testing"
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
