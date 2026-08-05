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
