export interface InvestmentCalculationInput {
  amountUsd: number;
  exchangeRate: number;
  dailyRewardNgn: number;
  durationBusinessDays: number;
  roiPercent?: number;
}

export interface InvestmentCalculation {
  amountNgn: number;
  dailyEarningsNgn: number;
  monthlyEarningsNgn: number;
  totalEarningsNgn: number;
  totalPayoutNgn: number;
  roiPercent: number;
  businessDaysRemaining: number;
}

/** Derive ROI % from total rewards over capital when settings omit an explicit ROI. */
export function deriveRoiPercent(totalEarningsNgn: number, amountNgn: number): number {
  if (!amountNgn || amountNgn <= 0) return 0;
  return Math.round((totalEarningsNgn / amountNgn) * 10000) / 100;
}

/**
 * CoinDistro investment projection:
 * capital = USD × rate
 * total rewards = daily × business days
 * monthly ≈ daily × 20 business days
 * total withdrawal = capital + total rewards
 * ROI = explicit setting, else rewards / capital
 */
export function calculateInvestment(input: InvestmentCalculationInput): InvestmentCalculation {
  const amountUsd = Math.max(0, Number(input.amountUsd) || 0);
  const rate = Math.max(0, Number(input.exchangeRate) || 0);
  const daily = Math.max(0, Number(input.dailyRewardNgn) || 0);
  const days = Math.max(0, Math.floor(Number(input.durationBusinessDays) || 0));
  const amountNgn = amountUsd * rate;
  const totalEarningsNgn = daily * days;
  const explicitRoi = Math.max(0, Number(input.roiPercent) || 0);
  const roiPercent = explicitRoi > 0 ? explicitRoi : deriveRoiPercent(totalEarningsNgn, amountNgn);
  return {
    amountNgn,
    dailyEarningsNgn: daily,
    monthlyEarningsNgn: daily * 20,
    totalEarningsNgn,
    totalPayoutNgn: amountNgn + totalEarningsNgn,
    roiPercent,
    businessDaysRemaining: days,
  };
}

export function calculateWithdrawal(amount: number, feePercent: number, penaltyPercent: number, early: boolean) {
  const requested = Math.max(0, Number(amount) || 0);
  const fee = requested * Math.max(0, feePercent || 0) / 100;
  const penalty = early ? requested * Math.max(0, penaltyPercent || 0) / 100 : 0;
  return { requested, fee, penalty, deductions: fee + penalty, net: Math.max(0, requested - fee - penalty) };
}

export function formatCurrency(value: number | null | undefined, currency = "₦") {
  if (value == null || Number.isNaN(value)) return "—";
  return `${currency}${new Intl.NumberFormat("en-NG", { maximumFractionDigits: 2 }).format(value)}`;
}

export function formatRoi(value: number | null | undefined) {
  if (value == null || Number.isNaN(value)) return "—";
  return `${new Intl.NumberFormat("en-NG", { maximumFractionDigits: 2 }).format(value)}%`;
}

export function getProgressPercentage(daysRemaining: number, totalDays: number) {
  if (!totalDays) return 0;
  return Math.min(100, Math.max(0, Math.round(((totalDays - Math.max(0, daysRemaining)) / totalDays) * 100)));
}

export function getCompletedBusinessDays(daysRemaining: number, totalDays: number) {
  if (!totalDays) return 0;
  return Math.min(totalDays, Math.max(0, totalDays - Math.max(0, daysRemaining)));
}

export function buildRewardTimeline(days: number, rewardAmount: number) {
  return Array.from({ length: Math.max(0, days) }, (_, index) => ({ day: index + 1, amount: rewardAmount }));
}

export function greetingForHour(hour = new Date().getHours()) {
  if (hour < 12) return "Good Morning";
  if (hour < 17) return "Good Afternoon";
  return "Good Evening";
}
