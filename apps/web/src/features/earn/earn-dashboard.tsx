"use client";

import * as React from "react";
import { motion } from "framer-motion";
import { RefreshCw, TrendingUp, Gift, Sparkles, Wallet } from "lucide-react";
import { useAuth } from "@/features/authentication/auth-provider";
import { useWallet, useInvestments } from "@/features/earn/hooks";
import { useCountUp } from "@/features/earn/components/use-count-up";
import { QuickActions } from "@/features/earn/components/quick-actions";
import { ActiveInvestmentCard } from "@/features/earn/components/active-investment-card";
import { RewardTimeline } from "@/features/earn/components/reward-timeline";
import { PremiumInvestmentCard } from "@/features/earn/components/premium-investment-card";
import { InvestmentModal } from "@/features/earn/components/investment-modal";
import { WithdrawalRequestModal } from "@/features/earn/withdrawal-request-modal";
import {
  formatCurrency,
  formatWithdrawalNextAvailable,
  getCompletedBusinessDays,
  getProgressPercentage,
  getWithdrawalCooldown,
} from "@/features/earn/utils";
import {
  useDashboard,
  useExchangeRate,
  useInvestmentSettings,
} from "@/features/investments";
import * as investmentApi from "@/features/investments/api";
import type { InvestmentSummary } from "@/lib/api/types";
import type { EarningsSummary } from "@/features/investments/types";
import { displayName } from "@/lib/utils/format";
import {
  ENABLED_PLANS,
  WITHDRAWAL_PROCESSING_HOURS,
  type InvestmentPlanConfig,
} from "@/features/earn/config/investment-plans";

function matchPlanName(amountUsd: number): string {
  const plan = ENABLED_PLANS.find((p) => p.usdAmount === amountUsd);
  return plan?.name ?? "Genesis";
}

function matchPlanRoi(amountUsd: number, fallbackRoi: number): number {
  const plan = ENABLED_PLANS.find((p) => p.usdAmount === amountUsd);
  return plan?.roiPercent ?? fallbackRoi;
}

function normalizeInvestment(
  item: InvestmentSummary | EarningsSummary,
  settingsRoi = 18,
): InvestmentSummary {
  if ("plan_name" in item) return item;
  const amountNgn = item.amount_ngn;
  const earnedNgn = item.total_earned_ngn;
  const amountUsd = item.amount_usd;
  return {
    id: item.id,
    plan_name: `${matchPlanName(amountUsd)} Plan`,
    amount_paid: amountNgn,
    allocated_cdt: amountNgn,
    roi_cdt: earnedNgn,
    roi_percent: matchPlanRoi(amountUsd, settingsRoi),
    daily_reward_ngn: item.daily_reward_ngn,
    status: item.status,
    lock_period_days: item.max_business_days,
    days_remaining: item.remaining_days,
    progress_pct: item.progress_pct,
    started_at: item.started_at,
    matures_at: item.maturity_date,
    created_at: item.created_at,
  };
}

function formatCompact(value: number) {
  if (value >= 1_000_000) return `${(value / 1_000_000).toFixed(1)}M`;
  if (value >= 1_000) return `${(value / 1_000).toFixed(1)}K`;
  return value.toFixed(0);
}

export function EarnDashboard() {
  const { user } = useAuth();
  const walletQ = useWallet();
  const legacyInvestmentsQ = useInvestments();
  const dashboardQ = useDashboard();
  const rateQ = useExchangeRate();
  const settingsQ = useInvestmentSettings();

  const [investOpen, setInvestOpen] = React.useState(false);
  const [selectedPlan, setSelectedPlan] = React.useState<InvestmentPlanConfig | null>(null);
  const [withdrawOpen, setWithdrawOpen] = React.useState(false);
  const [withdrawing, setWithdrawing] = React.useState(false);
  const [currency, setCurrency] = React.useState<"USD" | "NGN">("USD");
  const [refreshing, setRefreshing] = React.useState(false);

  const dashboard = dashboardQ.data;
  const settings = settingsQ.data;
  const rate = rateQ.data?.usd_to_ngn ?? dashboard?.exchange_rate ?? 0;
  const settingsRoi = settings?.roi_percent ?? 18;
  const investments: InvestmentSummary[] =
    dashboard?.investments?.map((item) => normalizeInvestment(item, settingsRoi)) ??
    (legacyInvestmentsQ.data?.investments ?? []).map((item) => normalizeInvestment(item, settingsRoi));

  const durationDays = settings?.max_business_days ?? 20;
  const dailyReward =
    investments.find((item) => item.status === "active")?.daily_reward_ngn ??
    settings?.daily_reward_ngn ??
    ENABLED_PLANS[0]?.dailyRewardNgn ??
    0;
  const processingHours = settings?.withdrawal_processing_hours ?? WITHDRAWAL_PROCESSING_HOURS;
  const feePercent = settings?.early_withdrawal_fee_percent ?? 0;
  const penaltyPercent = settings?.early_withdrawal_penalty_percent ?? 0;

  const available =
    dashboard?.available_balance_ngn ??
    dashboard?.referral_info?.withdrawable_balance_ngn ??
    walletQ.data?.available_balance ??
    0;
  const early = investments.some((item) => item.status === "active");
  const activeInvestment = investments.find((item) => item.status === "active");
  const totalDays = activeInvestment?.lock_period_days || durationDays;
  const daysRemaining = activeInvestment?.days_remaining ?? totalDays;
  const completedDays = activeInvestment
    ? getCompletedBusinessDays(daysRemaining, totalDays)
    : 0;
  const progress = activeInvestment
    ? activeInvestment.progress_pct ?? getProgressPercentage(daysRemaining, totalDays)
    : 0;
  const totalEarnedFromActive = activeInvestment?.roi_cdt ?? dashboard?.today_earnings_ngn ?? 0;

  const todayReward = dashboard?.today_earnings_ngn ?? 0;
  const monthReward = dashboard?.monthly_earnings_ngn ?? 0;
  const referralEarnings = dashboard?.referral_earnings_ngn ?? 0;
  const lastWithdrawalAt = dashboard?.last_withdrawal_at ?? null;
  const withdrawalCooldown = React.useMemo(
    () => getWithdrawalCooldown(lastWithdrawalAt),
    [lastWithdrawalAt],
  );

  const loading = dashboardQ.isLoading || rateQ.isLoading || settingsQ.isLoading;
  const firstName = displayName(user).split(" ")[0] || "Investor";

  // Animated counters
  const totalInvestedUsd = dashboard?.total_invested_usd ?? 0;
  const totalInvestedNgn = totalInvestedUsd * rate;
  const animatedInvested = useCountUp(currency === "USD" ? totalInvestedUsd : totalInvestedNgn, 700, !loading);
  const animatedToday = useCountUp(todayReward, 700, !loading);
  const animatedReferral = useCountUp(referralEarnings, 700, !loading);
  const animatedTotalEarned = useCountUp(totalEarnedFromActive || monthReward, 700, !loading);
  const animatedAvailable = useCountUp(available, 700, !loading);

  const openInvest = (plan: InvestmentPlanConfig) => {
    setSelectedPlan(plan);
    setInvestOpen(true);
  };

  const handleRefresh = React.useCallback(() => {
    setRefreshing(true);
    void Promise.all([
      dashboardQ.refetch(),
      rateQ.refetch(),
      settingsQ.refetch(),
      walletQ.refetch(),
      legacyInvestmentsQ.refetch(),
    ]).finally(() => {
      window.setTimeout(() => setRefreshing(false), 600);
    });
  }, [dashboardQ, rateQ, settingsQ, walletQ, legacyInvestmentsQ]);

  // After Paystack redirects back to /app/earn?reference=...&trxref=...,
  // verify the payment and refresh dashboard without a manual reload.
  React.useEffect(() => {
    if (typeof window === "undefined") return;
    const params = new URLSearchParams(window.location.search);
    const reference = params.get("reference") || params.get("trxref");
    if (!reference) return;

    let cancelled = false;
    (async () => {
      try {
        await investmentApi.verifyPaystackPayment(reference);
      } catch {
        // Webhook may already have activated the investment; still refresh UI.
      } finally {
        if (cancelled) return;
        // Strip gateway query params so refresh doesn't re-run verify forever.
        const url = new URL(window.location.href);
        url.searchParams.delete("reference");
        url.searchParams.delete("trxref");
        window.history.replaceState({}, "", url.pathname + url.search + url.hash);
        handleRefresh();
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [handleRefresh]);

  const submitWithdrawal = async (withdrawAmount: number) => {
    setWithdrawing(true);
    try {
      await investmentApi.requestWithdrawal(undefined, withdrawAmount);
      setWithdrawOpen(false);
    } finally {
      setWithdrawing(false);
    }
  };

  const displayValue = (value: number) => {
    if (currency === "USD") {
      return `$${formatCompact(value)}`;
    }
    return formatCurrency(value);
  };

  return (
    <div className="relative mx-auto max-w-6xl space-y-6 pb-10">
      {/* ─── Hero Section ─────────────────────────────── */}
      <motion.section
        initial={{ opacity: 0, y: 16 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.4 }}
        className="relative overflow-hidden rounded-3xl border border-primary/20 bg-gradient-to-br from-primary/15 via-card to-transparent p-6 sm:p-8"
      >
        <div className="pointer-events-none absolute -right-24 -top-24 h-64 w-64 rounded-full bg-primary/25 blur-3xl" />
        <div className="pointer-events-none absolute -bottom-24 -left-16 h-64 w-64 rounded-full bg-secondary/10 blur-3xl" />

        <div className="relative">
          <div className="flex items-center justify-between gap-3">
            <div>
              <p className="text-sm text-muted-foreground">Welcome back,</p>
              <h1 className="text-2xl font-bold text-foreground sm:text-3xl">{firstName}</h1>
            </div>
            <div className="flex items-center gap-2">
              <div className="flex rounded-xl border border-border bg-background/60 p-1">
                {(["USD", "NGN"] as const).map((c) => (
                  <button
                    key={c}
                    type="button"
                    onClick={() => setCurrency(c)}
                    className={`rounded-lg px-3 py-1.5 text-xs font-semibold transition-colors ${
                      currency === c ? "bg-primary text-primary-foreground" : "text-muted-foreground hover:text-foreground"
                    }`}
                  >
                    {c}
                  </button>
                ))}
              </div>
              <button
                type="button"
                onClick={handleRefresh}
                aria-label="Refresh rates"
                className="flex h-9 w-9 items-center justify-center rounded-xl border border-border bg-background/60 text-muted-foreground transition-colors hover:text-primary"
              >
                <RefreshCw className={`h-4 w-4 ${refreshing ? "animate-spin" : ""}`} />
              </button>
            </div>
          </div>

          <div className="mt-6">
            <p className="text-sm text-muted-foreground">Total Portfolio Value</p>
            <p className="mt-1 text-4xl font-bold tabular-nums tracking-tight text-foreground sm:text-5xl">
              {loading ? "…" : displayValue(animatedInvested)}
            </p>
            <p className="mt-2 text-sm text-muted-foreground">
              1 USD = {formatCurrency(rate)}
            </p>
          </div>

          <div className="mt-6 grid grid-cols-2 gap-3 sm:grid-cols-4">
            <div className="rounded-2xl border border-border/60 bg-background/50 p-4">
              <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
                <Sparkles className="h-3.5 w-3.5 text-amber-500" />
                Today&apos;s Earnings
              </div>
              <p className="mt-1.5 text-lg font-bold tabular-nums text-amber-500">
                {loading ? "…" : formatCurrency(animatedToday)}
              </p>
            </div>
            <div className="rounded-2xl border border-border/60 bg-background/50 p-4">
              <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
                <Gift className="h-3.5 w-3.5 text-cyan-500" />
                Referral Earnings
              </div>
              <p className="mt-1.5 text-lg font-bold tabular-nums text-cyan-500">
                {loading ? "…" : formatCurrency(animatedReferral)}
              </p>
            </div>
            <div className="rounded-2xl border border-border/60 bg-background/50 p-4">
              <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
                <TrendingUp className="h-3.5 w-3.5 text-emerald-500" />
                Total Earnings
              </div>
              <p className="mt-1.5 text-lg font-bold tabular-nums text-emerald-500">
                {loading ? "…" : formatCurrency(animatedTotalEarned)}
              </p>
            </div>
            <div className="rounded-2xl border border-border/60 bg-background/50 p-4">
              <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
                <Wallet className="h-3.5 w-3.5 text-primary" />
                Withdrawable
              </div>
              <p className="mt-1.5 text-lg font-bold tabular-nums text-primary">
                {loading ? "…" : formatCurrency(animatedAvailable)}
              </p>
            </div>
          </div>
        </div>
      </motion.section>

      {/* ─── Quick Actions ────────────────────────────── */}
      <QuickActions
        onInvest={() => openInvest(ENABLED_PLANS[0])}
        onWithdraw={() => setWithdrawOpen(true)}
        onRewards={() => {}}
        onReferrals={() => {}}
        withdrawDisabled={!withdrawalCooldown.available}
        withdrawLabel={
          withdrawalCooldown.available
            ? "Withdraw"
            : `${withdrawalCooldown.daysRemaining}d left`
        }
      />

      {/* ─── Withdrawal Status Card (weekly countdown) ── */}
      {!loading && (
        <section>
          <div
            className={`rounded-2xl border p-5 ${
              withdrawalCooldown.available
                ? "border-emerald-500/30 bg-emerald-500/10"
                : "border-amber-500/30 bg-amber-500/10"
            }`}
          >
            <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
              <div className="space-y-2">
                <p className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                  Weekly Withdrawals
                </p>
                <p className="text-xs font-medium text-muted-foreground">Next Withdrawal</p>
                {withdrawalCooldown.available ? (
                  <p className="text-xl font-bold text-emerald-600 dark:text-emerald-400">
                    Available Now
                  </p>
                ) : (
                  <>
                    <p className="text-xl font-bold text-amber-600 dark:text-amber-400">
                      Next withdrawal available in {withdrawalCooldown.daysRemaining} day
                      {withdrawalCooldown.daysRemaining === 1 ? "" : "s"}
                    </p>
                    {withdrawalCooldown.nextAvailableAt && (
                      <p className="text-xs text-muted-foreground">
                        Next withdrawal: {formatWithdrawalNextAvailable(withdrawalCooldown.nextAvailableAt)}
                      </p>
                    )}
                  </>
                )}
                <p className="text-xs text-muted-foreground">
                  Withdrawal Processing: {processingHours} Hours
                </p>
              </div>
              <button
                type="button"
                onClick={() => setWithdrawOpen(true)}
                disabled={!withdrawalCooldown.available}
                className={`rounded-xl px-4 py-2 text-sm font-semibold shadow-lg transition-transform ${
                  withdrawalCooldown.available
                    ? "bg-gradient-to-r from-primary to-secondary text-primary-foreground shadow-primary/20 active:scale-[0.98]"
                    : "cursor-not-allowed bg-muted text-muted-foreground shadow-none"
                }`}
              >
                {withdrawalCooldown.available ? "Withdraw" : "Locked"}
              </button>
            </div>
          </div>
        </section>
      )}

      {/* ─── Active Investment ────────────────────────── */}
      <section>
        <h2 className="mb-3 text-lg font-bold text-foreground">Active Investment</h2>
        <ActiveInvestmentCard
          investment={activeInvestment ?? null}
          totalDays={totalDays}
          completedDays={completedDays}
          daysRemaining={daysRemaining}
          progress={progress}
          dailyReward={dailyReward}
          totalEarned={totalEarnedFromActive}
          expectedRoi={activeInvestment?.roi_percent ?? settingsRoi}
          processingHours={processingHours}
        />
      </section>

      {/* ─── Investment Plans ─────────────────────────── */}
      <section>
        <div className="mb-3 flex items-center justify-between">
          <h2 className="text-lg font-bold text-foreground">Investment Plans</h2>
          <span className="text-xs text-muted-foreground">Tap to expand</span>
        </div>
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          {ENABLED_PLANS.map((plan) => (
            <PremiumInvestmentCard
              key={plan.id}
              plan={plan}
              exchangeRate={rate}
              onSelect={openInvest}
            />
          ))}
        </div>
      </section>

      {/* ─── Reward Timeline ──────────────────────────── */}
      {activeInvestment ? (
        <section>
          <h2 className="mb-3 text-lg font-bold text-foreground">Reward Timeline</h2>
          <RewardTimeline
            days={totalDays}
            rewardAmount={dailyReward}
            completedDays={completedDays}
          />
        </section>
      ) : null}

      {/* ─── Modals ───────────────────────────────────── */}
      <InvestmentModal
        open={investOpen}
        plan={selectedPlan}
        exchangeRate={rate}
        onClose={() => setInvestOpen(false)}
      />

      <WithdrawalRequestModal
        open={withdrawOpen}
        availableBalance={available}
        processingHours={processingHours}
        feePercent={feePercent}
        penaltyPercent={penaltyPercent}
        earlyWithdrawal={early}
        isSubmitting={withdrawing}
        lastWithdrawalAt={lastWithdrawalAt}
        onClose={() => setWithdrawOpen(false)}
        onConfirm={submitWithdrawal}
      />
    </div>
  );
}