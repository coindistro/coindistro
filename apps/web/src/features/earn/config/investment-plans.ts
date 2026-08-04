export interface InvestmentPlanConfig {
  id: string;
  name: string;
  slug: string;
  usdAmount: number;
  dailyRewardNgn: number;
  workingDays: number;
  monthlyRewardNgn: number;
  referralBonusPercent: number;
  minReferrals: number;
  displayOrder: number;
  enabled: boolean;
  futureActive: boolean;
  description: string;
  features: string[];
}

export const INVESTMENT_PLANS: readonly InvestmentPlanConfig[] = [
  {
    id: "genesis-seed",
    name: "Genesis Seed",
    slug: "genesis-seed",
    usdAmount: 30,
    dailyRewardNgn: 650,
    workingDays: 20,
    monthlyRewardNgn: 13000,
    referralBonusPercent: 10,
    minReferrals: 5,
    displayOrder: 1,
    enabled: true,
    futureActive: false,
    description: "Start your investment journey with our entry-level plan.",
    features: [
      "Daily rewards for 20 business days",
      "10% referral bonus",
      "Minimum 5 referrals for payout",
      "24-hour withdrawal processing",
    ],
  },
  {
    id: "genesis-growth",
    name: "Genesis Growth",
    slug: "genesis-growth",
    usdAmount: 50,
    dailyRewardNgn: 1100,
    workingDays: 20,
    monthlyRewardNgn: 22000,
    referralBonusPercent: 10,
    minReferrals: 5,
    displayOrder: 2,
    enabled: true,
    futureActive: false,
    description: "Grow your portfolio with enhanced daily rewards.",
    features: [
      "Higher daily rewards",
      "10% referral bonus",
      "Minimum 5 referrals for payout",
      "24-hour withdrawal processing",
    ],
  },
  {
    id: "genesis-prosper",
    name: "Genesis Prosper",
    slug: "genesis-prosper",
    usdAmount: 100,
    dailyRewardNgn: 2200,
    workingDays: 20,
    monthlyRewardNgn: 44000,
    referralBonusPercent: 10,
    minReferrals: 5,
    displayOrder: 3,
    enabled: true,
    futureActive: false,
    description: "Prosper with substantial daily returns on your investment.",
    features: [
      "Substantial daily rewards",
      "10% referral bonus",
      "Minimum 5 referrals for payout",
      "24-hour withdrawal processing",
    ],
  },
  {
    id: "genesis-elite",
    name: "Genesis Elite",
    slug: "genesis-elite",
    usdAmount: 200,
    dailyRewardNgn: 4500,
    workingDays: 20,
    monthlyRewardNgn: 90000,
    referralBonusPercent: 10,
    minReferrals: 5,
    displayOrder: 4,
    enabled: true,
    futureActive: false,
    description: "Elite tier for serious investors seeking premium returns.",
    features: [
      "Premium daily rewards",
      "10% referral bonus",
      "Minimum 5 referrals for payout",
      "24-hour withdrawal processing",
    ],
  },
  {
    id: "genesis-legacy",
    name: "Genesis Legacy",
    slug: "genesis-legacy",
    usdAmount: 500,
    dailyRewardNgn: 12000,
    workingDays: 20,
    monthlyRewardNgn: 240000,
    referralBonusPercent: 10,
    minReferrals: 5,
    displayOrder: 5,
    enabled: true,
    futureActive: false,
    description: "Build a legacy with significant investment returns.",
    features: [
      "Significant daily rewards",
      "10% referral bonus",
      "Minimum 5 referrals for payout",
      "24-hour withdrawal processing",
    ],
  },
  {
    id: "genesis-founders",
    name: "Genesis Founders",
    slug: "genesis-founders",
    usdAmount: 1000,
    dailyRewardNgn: 25000,
    workingDays: 20,
    monthlyRewardNgn: 500000,
    referralBonusPercent: 10,
    minReferrals: 5,
    displayOrder: 6,
    enabled: true,
    futureActive: false,
    description: "Founders tier for maximum returns and exclusive benefits.",
    features: [
      "Maximum daily rewards",
      "10% referral bonus",
      "Minimum 5 referrals for payout",
      "24-hour withdrawal processing",
    ],
  },
] as const;

export const ENABLED_PLANS = INVESTMENT_PLANS.filter(
  (plan) => plan.enabled && !plan.futureActive,
);

export const PAYSTACK_PAYMENT_LINK = "https://paystack.shop/pay/coindistro";