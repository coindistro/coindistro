// ─── CoinDistro Investment Plan Configuration (Centralized) ───────────────
// Single source of truth for all investment tiers. The ROI ladder increases
// by exactly 3% per tier, starting at 18% for Genesis. Plans are consumed by
// the Earn dashboard, investment cards, calculator, payment modal, active
// investment card, and admin/API responses.

export interface InvestmentPlanConfig {
  id: string;
  name: string;
  slug: string;
  usdAmount: number;
  roiPercent: number;
  dailyRewardNgn: number;
  workingDays: number;
  monthlyRewardNgn: number;
  totalReturnNgn: number;
  referralBonusPercent: number;
  minReferrals: number;
  displayOrder: number;
  enabled: boolean;
  futureActive: boolean;
  description: string;
  features: string[];
}

// Default exchange rate (₦1,400 / $1). Overridden at runtime by the exchange
// rate service, but kept here as the fallback for the calculator & display.
export const DEFAULT_EXCHANGE_RATE = 1400;

// Withdrawal rules (centralized).
export const WITHDRAWAL_INTERVAL_DAYS = 7;
export const WITHDRAWAL_PROCESSING_HOURS = 24;

// ROI ladder: each tier is exactly +3% above the previous, starting at 18%.
export const ROI_LADDER: readonly { tier: string; roiPercent: number }[] = [
  { tier: "Genesis", roiPercent: 18 },
  { tier: "Starter", roiPercent: 21 },
  { tier: "Growth", roiPercent: 24 },
  { tier: "Premium", roiPercent: 27 },
  { tier: "Professional", roiPercent: 30 },
  { tier: "Enterprise", roiPercent: 33 },
];

/** Derive the daily reward (NGN) for a plan from its ROI and capital. */
export function deriveDailyRewardNgn(usdAmount: number, roiPercent: number, exchangeRate: number, workingDays = 20): number {
  const capitalNgn = usdAmount * exchangeRate;
  const totalReturnNgn = capitalNgn * (roiPercent / 100);
  return totalReturnNgn / workingDays;
}

/** Derive the total return (NGN) for a plan from its ROI and capital. */
export function deriveTotalReturnNgn(usdAmount: number, roiPercent: number, exchangeRate: number): number {
  const capitalNgn = usdAmount * exchangeRate;
  return capitalNgn * (roiPercent / 100);
}

export const INVESTMENT_PLANS: readonly InvestmentPlanConfig[] = [
  {
    id: "genesis",
    name: "Genesis",
    slug: "genesis",
    usdAmount: 10,
    roiPercent: 18,
    dailyRewardNgn: 126,
    workingDays: 20,
    monthlyRewardNgn: 2520,
    totalReturnNgn: 2520,
    referralBonusPercent: 10,
    minReferrals: 5,
    displayOrder: 1,
    enabled: true,
    futureActive: false,
    description: "Start your investment journey with the smallest tier.",
    features: [
      "18% ROI over 20 business days",
      "Daily rewards for 20 business days",
      "10% referral bonus",
      "24-hour withdrawal processing",
    ],
  },
  {
    id: "starter",
    name: "Starter",
    slug: "starter",
    usdAmount: 30,
    roiPercent: 21,
    dailyRewardNgn: 441,
    workingDays: 20,
    monthlyRewardNgn: 8820,
    totalReturnNgn: 8820,
    referralBonusPercent: 10,
    minReferrals: 5,
    displayOrder: 2,
    enabled: true,
    futureActive: false,
    description: "Perfect for beginners stepping up from Genesis.",
    features: [
      "21% ROI over 20 business days",
      "Daily rewards for 20 business days",
      "10% referral bonus",
      "24-hour withdrawal processing",
    ],
  },
  {
    id: "growth",
    name: "Growth",
    slug: "growth",
    usdAmount: 100,
    roiPercent: 24,
    dailyRewardNgn: 1680,
    workingDays: 20,
    monthlyRewardNgn: 33600,
    totalReturnNgn: 33600,
    referralBonusPercent: 10,
    minReferrals: 5,
    displayOrder: 3,
    enabled: true,
    futureActive: false,
    description: "Grow your portfolio with enhanced daily returns.",
    features: [
      "24% ROI over 20 business days",
      "Daily rewards for 20 business days",
      "10% referral bonus",
      "24-hour withdrawal processing",
    ],
  },
  {
    id: "premium",
    name: "Premium",
    slug: "premium",
    usdAmount: 250,
    roiPercent: 27,
    dailyRewardNgn: 4725,
    workingDays: 20,
    monthlyRewardNgn: 94500,
    totalReturnNgn: 94500,
    referralBonusPercent: 10,
    minReferrals: 5,
    displayOrder: 4,
    enabled: true,
    futureActive: false,
    description: "Premium tier for serious investors seeking strong returns.",
    features: [
      "27% ROI over 20 business days",
      "Daily rewards for 20 business days",
      "10% referral bonus",
      "24-hour withdrawal processing",
    ],
  },
  {
    id: "professional",
    name: "Professional",
    slug: "professional",
    usdAmount: 500,
    roiPercent: 30,
    dailyRewardNgn: 10500,
    workingDays: 20,
    monthlyRewardNgn: 210000,
    totalReturnNgn: 210000,
    referralBonusPercent: 10,
    minReferrals: 5,
    displayOrder: 5,
    enabled: true,
    futureActive: false,
    description: "Professional tier with significant daily returns.",
    features: [
      "30% ROI over 20 business days",
      "Daily rewards for 20 business days",
      "10% referral bonus",
      "24-hour withdrawal processing",
    ],
  },
  {
    id: "enterprise",
    name: "Enterprise",
    slug: "enterprise",
    usdAmount: 1000,
    roiPercent: 33,
    dailyRewardNgn: 23100,
    workingDays: 20,
    monthlyRewardNgn: 462000,
    totalReturnNgn: 462000,
    referralBonusPercent: 10,
    minReferrals: 5,
    displayOrder: 6,
    enabled: true,
    futureActive: false,
    description: "Enterprise tier for maximum returns and exclusive benefits.",
    features: [
      "33% ROI over 20 business days",
      "Daily rewards for 20 business days",
      "10% referral bonus",
      "24-hour withdrawal processing",
    ],
  },
] as const;

export const ENABLED_PLANS = INVESTMENT_PLANS.filter(
  (plan) => plan.enabled && !plan.futureActive,
);