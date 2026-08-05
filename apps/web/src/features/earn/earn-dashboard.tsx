"use client";

import * as React from "react";
import { motion } from "framer-motion";
import { RefreshCw } from "lucide-react";
import { useAuth } from "@/features/authentication/auth-provider";
import { useWallet, useInvestments } from "@/features/earn/hooks";
import { useCountUp } from "@/features/earn/components/use-count-up";
import { QuickActions } from "@/features/earn/components/quick-actions";
import { ActiveInvestmentCard } from "@/features/earn/components/active-investment-card";
import { PortfolioAssetsCard } from "@/features/earn/components/portfolio-assets-card";
import {
  PortfolioHistory,
  buildPortfolioHistory,
} from "@/features/earn/components/portfolio-history";
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
  const earnedUsd =
    item.total_earned_usd ??
    (item.exchange_rate > 0 ? earnedNgn / item.exchange_rate : 0);
  const roiPct =
    item.roi_percentage ??
    (amountUsd > 0 ? (earnedUsd / amountUsd) * 100 : matchPlanRoi(amountUsd, settingsRoi));
  return {
    id: item.id,
    plan_name: `${matchPlanName(amountUsd)} Plan`,
    amount_paid: amountNgn,
    allocated_cdt: amountNgn,
    roi_cdt: earnedNgn,
    roi_percent: roiPct,
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

function formatUsd(value: number) {
  return `$${value.toLocaleString(undefined, { maximumFractionDigits: 2 })}`;
}

function formatDual(usd: number, ngn: number) {
  return {
    primaryUsd: formatUsd(usd),
    secondaryNgn: formatCurrency(ngn),
  };
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
  const totalEarnedFromActive =
    activeInvestment?.roi_cdt ?? dashboard?.total_profit_ngn ?? dashboard?.today_earnings_ngn ?? 0;
  const lastWithdrawalAt = dashboard?.last_withdrawal_at ?? null;
  const withdrawalCooldown = React.useMemo(
    () => getWithdrawalCooldown(lastWithdrawalAt),
    [lastWithdrawalAt],
  );

  // Referral unlock gate (withdrawals_unlocked = successful referrals >= 5)
  const minReferrals = dashboard?.min_referrals_required ?? dashboard?.referral_info?.minimum_target ?? 5;
  const activeReferrals = dashboard?.active_referrals ?? dashboard?.referral_info?.active_referrals ?? 0;
  const remainingReferrals =
    dashboard?.remaining_referrals ?? Math.max(0, minReferrals - activeReferrals);
  const referralsUnlocked =
    dashboard?.withdrawals_unlocked ?? activeReferrals >= minReferrals;
  const canWithdraw = referralsUnlocked && withdrawalCooldown.available;

  const loading = dashboardQ.isLoading || rateQ.isLoading || settingsQ.isLoading;
  const firstName = displayName(user).split(" ")[0] || "Investor";

  // Portfolio metrics (capital + profit) — position view, not free cash
  const totalInvestedUsd =
    dashboard?.capital_invested_usd ?? dashboard?.total_invested_usd ?? 0;
  const totalInvestedNgn =
    dashboard?.capital_invested_ngn ??
    dashboard?.total_invested_ngn ??
    totalInvestedUsd * rate;
  const totalProfitUsd =
    dashboard?.profit_earned_usd ??
    dashboard?.total_profit_usd ??
    (rate > 0 ? (dashboard?.total_profit_ngn ?? 0) / rate : 0);
  const totalProfitNgn =
    dashboard?.profit_earned_ngn ?? dashboard?.total_profit_ngn ?? totalProfitUsd * rate;
  const portfolioUsd =
    dashboard?.portfolio_value_usd ?? totalInvestedUsd + totalProfitUsd;
  const portfolioNgn =
    dashboard?.portfolio_value_ngn ?? totalInvestedNgn + totalProfitNgn;
  const availableUsd = dashboard?.available_balance_usd ?? 0;
  const lockedUsd =
    dashboard?.locked_balance_usd ??
    (referralsUnlocked ? totalInvestedUsd : totalInvestedUsd + totalProfitUsd);
  const portfolioWalletUsd =
    dashboard?.portfolio_value_usd ?? availableUsd + lockedUsd;
  const referralUsd =
    dashboard?.referral_earnings_usd ??
    (rate > 0 ? (dashboard?.referral_earnings_ngn ?? 0) / rate : 0);
  const withdrawableUsd = referralsUnlocked
    ? dashboard?.withdrawable_balance_usd ?? availableUsd
    : 0;
  const roiPercentage =
    dashboard?.roi_percentage ??
    (totalInvestedUsd > 0 ? (totalProfitUsd / totalInvestedUsd) * 100 : 0);

  // Active position metrics for the investment card
  const activeEarnings = dashboard?.investments?.find((i) => i.status === "active");
  const positionCapitalUsd = activeEarnings?.amount_usd ?? totalInvestedUsd;
  const positionCapitalNgn = activeEarnings?.amount_ngn ?? totalInvestedNgn;
  const positionProfitUsd =
    activeEarnings?.total_earned_usd ??
    (activeEarnings && activeEarnings.exchange_rate > 0
      ? activeEarnings.total_earned_ngn / activeEarnings.exchange_rate
      : totalProfitUsd);
  const positionProfitNgn = activeEarnings?.total_earned_ngn ?? totalProfitNgn;
  const positionValueUsd =
    activeEarnings?.portfolio_value_usd ?? positionCapitalUsd + positionProfitUsd;
  const positionValueNgn =
    activeEarnings?.portfolio_value_ngn ?? positionCapitalNgn + positionProfitNgn;
  const positionRoi =
    activeEarnings?.roi_percentage ??
    (positionCapitalUsd > 0 ? (positionProfitUsd / positionCapitalUsd) * 100 : roiPercentage);

  // Animated counters
  const animatedPortfolio = useCountUp(portfolioUsd, 700, !loading);
  const animatedInvested = useCountUp(totalInvestedUsd, 700, !loading);
  const animatedProfit = useCountUp(totalProfitUsd, 700, !loading);

  const historyEvents = React.useMemo(
    () =>
      buildPortfolioHistory({
        capitalUsd: positionCapitalUsd,
        profitUsd: positionProfitUsd,
        portfolioUsd: positionValueUsd,
        capitalNgn: positionCapitalNgn,
        profitNgn: positionProfitNgn,
        startedAt: activeInvestment?.started_at,
        hasProfit: positionProfitUsd > 0 || positionProfitNgn > 0,
      }),
    [
      positionCapitalUsd,
      positionProfitUsd,
      positionValueUsd,
      positionCapitalNgn,
      positionProfitNgn,
      activeInvestment?.started_at,
    ],
  );

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

  const capitalDisplay = formatDual(totalInvestedUsd, totalInvestedNgn);
  const profitDisplay = formatDual(totalProfitUsd, totalProfitNgn);
  const portfolioDisplay = formatDual(portfolioUsd, portfolioNgn);

  return (
    <div className="relative mx-auto max-w-6xl space-y-6 pb-10">
      {/* ─── Hero Portfolio Card ───────────────────────── */}
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
            <button
              type="button"
              onClick={handleRefresh}
              aria-label="Refresh portfolio"
              className="flex h-9 w-9 items-center justify-center rounded-xl border border-border bg-background/60 text-muted-foreground transition-colors hover:text-primary"
            >
              <RefreshCw className={`h-4 w-4 ${refreshing ? "animate-spin" : ""}`} />
            </button>
          </div>

          <div className="mt-6">
            <p className="text-sm font-medium text-muted-foreground">Portfolio Value</p>
            <p className="mt-1 text-4xl font-bold tabular-nums tracking-tight text-foreground sm:text-5xl">
              {loading ? "…" : formatUsd(animatedPortfolio)}
            </p>
            <p className="mt-2 text-base text-muted-foreground">
              ≈ {loading ? "…" : formatCurrency(portfolioNgn)}
            </p>
            <p className="mt-1 text-xs text-muted-foreground">
              1 USD = {formatCurrency(rate || 1400)} · Position value (capital + profit)
            </p>
          </div>

          <div className="mt-6 grid grid-cols-2 gap-3 sm:grid-cols-4">
            <div className="rounded-2xl border border-border/60 bg-background/50 p-4">
              <p className="text-xs text-muted-foreground">Invested Capital</p>
              <p className="mt-1.5 text-lg font-bold tabular-nums text-foreground">
                {loading ? "…" : formatUsd(animatedInvested)}
              </p>
              <p className="text-xs text-muted-foreground">{capitalDisplay.secondaryNgn}</p>
            </div>
            <div className="rounded-2xl border border-border/60 bg-background/50 p-4">
              <p className="text-xs text-muted-foreground">Profit Earned</p>
              <p className="mt-1.5 text-lg font-bold tabular-nums text-emerald-500">
                {loading ? "…" : `+${formatUsd(animatedProfit)}`}
              </p>
              <p className="text-xs text-muted-foreground">+{profitDisplay.secondaryNgn}</p>
            </div>
            <div className="rounded-2xl border border-border/60 bg-background/50 p-4">
              <p className="text-xs text-muted-foreground">ROI</p>
              <p className="mt-1.5 text-lg font-bold tabular-nums text-amber-500">
                {loading ? "…" : `${roiPercentage.toFixed(2)}%`}
              </p>
              <p className="text-xs text-muted-foreground">On invested capital</p>
            </div>
            <div className="rounded-2xl border border-border/60 bg-background/50 p-4">
              <p className="text-xs text-muted-foreground">Status</p>
              <p className="mt-1.5 text-lg font-bold text-emerald-500">
                {activeInvestment ? "Active" : "—"}
              </p>
              <p className="text-xs text-muted-foreground">
                {activeInvestment ? "Growing" : "No active plan"}
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
        withdrawDisabled={!canWithdraw}
        withdrawLabel={
          !referralsUnlocked
            ? "Locked"
            : withdrawalCooldown.available
              ? "Withdraw"
              : `${withdrawalCooldown.daysRemaining}d left`
        }
      />

      {/* ─── Withdrawal Lock (referrals) ──────────────── */}
      {!loading && (
        <section>
          <div
            className={`rounded-2xl border p-5 ${
              referralsUnlocked
                ? "border-emerald-500/30 bg-emerald-500/10"
                : "border-amber-500/30 bg-amber-500/10"
            }`}
          >
            <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
              <div className="space-y-2">
                <p className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                  Withdrawal Lock Status
                </p>
                {referralsUnlocked ? (
                  <>
                    <p className="text-xl font-bold text-emerald-600 dark:text-emerald-400">
                      Withdrawals Unlocked
                    </p>
                    <p className="text-sm text-muted-foreground">
                      You have {activeReferrals} successful referrals (required {minReferrals}).
                    </p>
                  </>
                ) : (
                  <>
                    <p className="text-xl font-bold text-amber-600 dark:text-amber-400">
                      Withdrawals Locked
                    </p>
                    <p className="text-sm text-muted-foreground">
                      {dashboard?.withdrawal_lock_message ||
                        `Complete ${minReferrals} successful referrals to unlock your earnings.`}
                    </p>
                    <p className="text-sm font-semibold text-foreground">
                      Progress: {activeReferrals} / {minReferrals}
                    </p>
                    <div className="mt-2 h-2 max-w-xs overflow-hidden rounded-full bg-muted">
                      <div
                        className="h-full rounded-full bg-gradient-to-r from-primary to-secondary transition-all"
                        style={{
                          width: `${Math.min(100, (activeReferrals / Math.max(1, minReferrals)) * 100)}%`,
                        }}
                      />
                    </div>
                  </>
                )}
                {referralsUnlocked && !withdrawalCooldown.available && (
                  <p className="text-xs text-muted-foreground">
                    Next weekly withdrawal in {withdrawalCooldown.daysRemaining} day
                    {withdrawalCooldown.daysRemaining === 1 ? "" : "s"}
                    {withdrawalCooldown.nextAvailableAt
                      ? ` (${formatWithdrawalNextAvailable(withdrawalCooldown.nextAvailableAt)})`
                      : ""}
                  </p>
                )}
                <p className="text-xs text-muted-foreground">
                  Withdrawal Processing: {processingHours} Hours
                </p>
              </div>
              <button
                type="button"
                onClick={() => setWithdrawOpen(true)}
                disabled={!canWithdraw}
                className={`rounded-xl px-4 py-2 text-sm font-semibold shadow-lg transition-transform ${
                  canWithdraw
                    ? "bg-gradient-to-r from-primary to-secondary text-primary-foreground shadow-primary/20 active:scale-[0.98]"
                    : "cursor-not-allowed bg-muted text-muted-foreground shadow-none"
                }`}
              >
                {canWithdraw ? "Withdraw" : "Locked"}
              </button>
            </div>
          </div>
        </section>
      )}

      {/* ─── Wallet: Available vs Locked ──────────────── */}
      <PortfolioAssetsCard
        availableUsd={availableUsd}
        lockedUsd={lockedUsd}
        portfolioUsd={portfolioWalletUsd}
        capitalUsd={totalInvestedUsd}
        profitUsd={totalProfitUsd}
        referralUsd={referralUsd}
        withdrawableUsd={withdrawableUsd}
        withdrawalsUnlocked={referralsUnlocked}
        lockMessage={dashboard?.withdrawal_lock_message}
        minReferrals={minReferrals}
        activeReferrals={activeReferrals}
        exchangeRate={rate || 1400}
      />

      {/* ─── Earnings Summary ─────────────────────────── */}
      <section className="space-y-3">
        <div>
          <h2 className="text-lg font-bold text-foreground">Earnings Summary</h2>
          <p className="text-sm text-muted-foreground">Capital, rewards, and lock status</p>
        </div>
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6">
          <SummaryTile label="Capital Invested" value={portfolioDisplay.primaryUsd} sub={capitalDisplay.secondaryNgn} />
          <SummaryTile label="Total Profit" value={`+${profitDisplay.primaryUsd}`} sub={`+${profitDisplay.secondaryNgn}`} accent="text-emerald-500" />
          <SummaryTile label="Total Portfolio" value={portfolioDisplay.primaryUsd} sub={portfolioDisplay.secondaryNgn} accent="text-primary" />
          <SummaryTile label="Total Rewards" value={formatCurrency(totalProfitNgn)} sub={`${formatUsd(totalProfitUsd)} USD`} />
          <SummaryTile
            label="Pending Withdrawals"
            value={formatCurrency(dashboard?.pending_withdrawal_ngn ?? 0)}
            sub="In review / processing"
          />
          <SummaryTile
            label="Withdrawal Lock"
            value={referralsUnlocked ? "Unlocked" : "Locked"}
            sub={
              referralsUnlocked
                ? `${activeReferrals}/${minReferrals} referrals`
                : `${remainingReferrals} remaining · ${activeReferrals}/${minReferrals}`
            }
            accent={referralsUnlocked ? "text-emerald-500" : "text-amber-500"}
          />
        </div>
      </section>

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
          expectedRoi={settingsRoi}
          processingHours={processingHours}
          position={{
            capitalUsd: positionCapitalUsd,
            capitalNgn: positionCapitalNgn,
            profitUsd: positionProfitUsd,
            profitNgn: positionProfitNgn,
            currentValueUsd: positionValueUsd,
            currentValueNgn: positionValueNgn,
            roiPercent: positionRoi,
            exchangeRate: rate || 1400,
          }}
        />
      </section>

      {/* ─── Portfolio History ────────────────────────── */}
      <PortfolioHistory events={historyEvents} />

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
        availableBalance={referralsUnlocked ? available : 0}
        processingHours={processingHours}
        feePercent={feePercent}
        penaltyPercent={penaltyPercent}
        earlyWithdrawal={early}
        isSubmitting={withdrawing}
        lastWithdrawalAt={lastWithdrawalAt}
        referralsUnlocked={referralsUnlocked}
        activeReferrals={activeReferrals}
        minReferrals={minReferrals}
        lockMessage={dashboard?.withdrawal_lock_message}
        onClose={() => setWithdrawOpen(false)}
        onConfirm={submitWithdrawal}
      />
    </div>
  );
}

function SummaryTile({
  label,
  value,
  sub,
  accent,
}: {
  label: string;
  value: string;
  sub?: string;
  accent?: string;
}) {
  return (
    <div className="rounded-[1.25rem] border border-border/60 bg-card/80 p-4 shadow-sm">
      <p className="text-xs text-muted-foreground">{label}</p>
      <p className={`mt-1.5 text-base font-bold tabular-nums ${accent ?? "text-foreground"}`}>
        {value}
      </p>
      {sub ? <p className="mt-0.5 text-[11px] text-muted-foreground">{sub}</p> : null}
    </div>
  );
}