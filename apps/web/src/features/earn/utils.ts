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

// ─── Weekly Withdrawal Lock ─────────────────────────────────

export const WITHDRAWAL_INTERVAL_DAYS = 7;
export const WITHDRAWAL_PROCESSING_HOURS = 24;

export interface WithdrawalCooldown {
  /** True when the user is allowed to submit a new withdrawal. */
  available: boolean;
  /** Whole days remaining until the next withdrawal is available. */
  daysRemaining: number;
  /** ISO timestamp of when the next withdrawal becomes available. */
  nextAvailableAt: string | null;
}

/** Compute the weekly withdrawal cooldown based on the last withdrawal time. */
export function getWithdrawalCooldown(lastWithdrawalAt?: string | null): WithdrawalCooldown {
  if (!lastWithdrawalAt) {
    return { available: true, daysRemaining: 0, nextAvailableAt: null };
  }
  const last = new Date(lastWithdrawalAt).getTime();
  if (Number.isNaN(last)) {
    return { available: true, daysRemaining: 0, nextAvailableAt: null };
  }
  const nextAvailable = last + WITHDRAWAL_INTERVAL_DAYS * 24 * 60 * 60 * 1000;
  const now = Date.now();
  const remainingMs = nextAvailable - now;
  if (remainingMs <= 0) {
    return { available: true, daysRemaining: 0, nextAvailableAt: new Date(nextAvailable).toISOString() };
  }
  return {
    available: false,
    daysRemaining: Math.ceil(remainingMs / (24 * 60 * 60 * 1000)),
    nextAvailableAt: new Date(nextAvailable).toISOString(),
  };
}

/** Format the next available date e.g. "Tuesday, 14 August". */
export function formatWithdrawalNextAvailable(isoString?: string | null): string {
  if (!isoString) return "";
  const date = new Date(isoString);
  if (Number.isNaN(date.getTime())) return "";
  return new Intl.DateTimeFormat("en-GB", {
    weekday: "long",
    day: "numeric",
    month: "long",
  }).format(date);
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

export function buildInvestmentGrowthSeries(
  payments: Array<{ amount_usd?: number; created_at?: string | null }>,
  currentValueUsd: number,
) {
  const points = payments
    .filter((payment) => Number(payment.amount_usd) > 0)
    .slice(0, 5)
    .map((payment) => ({
      name: payment.created_at
        ? new Intl.DateTimeFormat("en-US", { month: "short", day: "2-digit" }).format(new Date(payment.created_at))
        : "Deposit",
      value: Number(payment.amount_usd) || 0,
    }));

  if (!points.length) {
    return [{ name: "Start", value: 0 }, { name: "Now", value: Math.max(0, currentValueUsd) }];
  }

  const cumulative = points.reduce<Array<{ name: string; value: number }>>((acc, point) => {
    const prev = acc[acc.length - 1]?.value ?? 0;
    acc.push({ name: point.name, value: prev + point.value });
    return acc;
  }, []);

  const targetValue = Math.max(0, currentValueUsd);
  if (points.length > 0) {
    cumulative.push({ name: "Now", value: targetValue });
  } else if (cumulative[cumulative.length - 1]?.value !== targetValue) {
    cumulative[cumulative.length - 1].value = targetValue;
  }

  return cumulative;
}

export function greetingForHour(hour = new Date().getHours()) {
  if (hour < 12) return "Good Morning";
  if (hour < 17) return "Good Afternoon";
  return "Good Evening";
}
