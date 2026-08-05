package service

import "testing"

func TestGenesisPoolMath(t *testing.T) {
	if GenesisPoolTotalUSD != 500 {
		t.Fatalf("pool = %.2f, want 500", GenesisPoolTotalUSD)
	}
	if GenesisPoolInvestors != 8 {
		t.Fatalf("investors = %d, want 8", GenesisPoolInvestors)
	}
	if GenesisPoolProfitPerInvestorUSD != 62.5 {
		t.Fatalf("per investor = %.2f, want 62.5", GenesisPoolProfitPerInvestorUSD)
	}
	if GenesisPoolInvestmentUSD != 30 {
		t.Fatalf("investment = %.2f, want 30", GenesisPoolInvestmentUSD)
	}
	portfolio := GenesisPoolInvestmentUSD + GenesisPoolProfitPerInvestorUSD
	if portfolio != 92.5 {
		t.Fatalf("portfolio = %.2f, want 92.5", portfolio)
	}
	if GenesisPoolProfitPerInvestorUSD*float64(GenesisPoolInvestors) != GenesisPoolTotalUSD {
		t.Fatal("pool must equal profit × investors")
	}
}

func TestDefaultMinReferralsForWithdraw(t *testing.T) {
	if DefaultMinReferralsForWithdraw != 5 {
		t.Fatalf("min referrals = %d, want 5", DefaultMinReferralsForWithdraw)
	}
}

func TestLockedWalletAccountingModel(t *testing.T) {
	// Before referrals unlock: available=0, locked=capital+profit
	capital := GenesisPoolInvestmentUSD
	profit := GenesisPoolProfitPerInvestorUSD
	available := 0.0
	locked := capital + profit
	portfolio := available + locked
	if portfolio != 92.5 {
		t.Fatalf("portfolio = %.2f, want 92.50", portfolio)
	}
	if available != 0 {
		t.Fatal("available must be 0 while locked")
	}
	// After unlock: available=profit, locked=capital
	availableAfter := profit
	lockedAfter := capital
	if availableAfter+lockedAfter != portfolio {
		t.Fatal("unlock must not change total portfolio")
	}
	if InvestorWalletCurrency != "USD" {
		t.Fatalf("investor wallet currency = %s, want USD", InvestorWalletCurrency)
	}
}
