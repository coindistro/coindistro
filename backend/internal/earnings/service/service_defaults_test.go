package service

import "testing"

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
