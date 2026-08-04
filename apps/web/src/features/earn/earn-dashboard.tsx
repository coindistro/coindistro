"use client";

import Link from "next/link";
import * as React from "react";
import { useQueryClient } from "@tanstack/react-query";
import {
  Badge,
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  Progress,
  Skeleton,
} from "@coindistro/cds";
import {
  ArrowRight,
  CheckCircle2,
  Clock3,
  CreditCard,
  Gift,
  Landmark,
  PiggyBank,
  Sparkles,
  TrendingUp,
  Wallet,
} from "lucide-react";
import { Area, AreaChart, CartesianGrid, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts";
import { useAuth } from "@/features/authentication/auth-provider";
import { useWallet, useInvestments } from "@/features/earn/hooks";
import { InvestmentGrid } from "@/features/earn/components/investment-grid";
import { InvestmentModal } from "@/features/earn/components/investment-modal";
import { WithdrawalRequestModal } from "@/features/earn/withdrawal-request-modal";
import {
  buildInvestmentGrowthSeries,
  buildRewardTimeline,
  formatCurrency,
  getCompletedBusinessDays,
  getProgressPercentage,
  greetingForHour,
} from "@/features/earn/utils";
import {
  useDashboard,
  useExchangeRate,
  useInvestmentSettings,
  usePaymentHistory,
  useRewardHistory,
  useWithdrawalHistory,
} from "@/features/investments";
import * as investmentApi from "@/features/investments/api";
import type { InvestmentSummary } from "@/lib/api/types";
import type { EarningsSummary } from "@/features/investments/types";
import { displayName } from "@/lib/utils/format";
import { ENABLED_PLANS, type InvestmentPlanConfig } from "@/features/earn/config/investment-plans";

function formatDate(value?: string | null) {
  if (!value) return "—";
  return new Date(value).toLocaleDateString(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

function normalizeInvestment(item: InvestmentSummary | EarningsSummary): InvestmentSummary {
  if ("plan_name" in item) return item;
  const amountNgn = item.amount_ngn;
  const earnedNgn = item.total_earned_ngn;
  return {
    id: item.id,
    plan_name: "CoinDistro Plan",
    amount_paid: amountNgn,
    allocated_cdt: amountNgn,
    roi_cdt: earnedNgn,
    roi_percent: 0,
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

function MetricCard({
  title,
  value,
  hint,
  icon: Icon,
  accent,
  loading,
  numeric,
  prefix = "",
  suffix = "",
}: {
  title: string;
  value: string;
  hint?: string;
  icon: React.ComponentType<{ className?: string }>;
  accent: string;
  loading?: boolean;
  numeric?: number;
  prefix?: string;
  suffix?: string;
}) {
  const shown = loading ? "…" : numeric != null ? `${prefix}${numeric.toLocaleString()}${suffix}` : value;

  return (
    <div
      className={`group relative overflow-hidden rounded-2xl border border-white/10 bg-white/5 p-4 shadow-lg backdrop-blur-md transition duration-300 hover:-translate-y-0.5 hover:border-primary/40 hover:shadow-primary/20 ${accent}`}
    >
      <div className="absolute inset-0 bg-gradient-to-br from-primary/10 via-transparent to-cyan-400/5 opacity-80" />
      <div className="relative flex items-start justify-between gap-3">
        <div>
          <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">{title}</p>
          <p className="mt-2 text-2xl font-bold tabular-nums tracking-tight">{shown}</p>
          {hint ? <p className="mt-1 text-xs text-muted-foreground">{hint}</p> : null}
        </div>
        <div className="rounded-xl bg-primary/15 p-2 text-primary transition group-hover:scale-110">
          <Icon className="h-5 w-5" />
        </div>
      </div>
    </div>
  );
}

export function EarnDashboard() {
  const { user } = useAuth();
  const walletQ = useWallet();
  const legacyInvestmentsQ = useInvestments();
  const dashboardQ = useDashboard();
  const rateQ = useExchangeRate();
  const settingsQ = useInvestmentSettings();
  const paymentsQ = usePaymentHistory();
  const withdrawalsQ = useWithdrawalHistory();
  const rewardsQ = useRewardHistory();
  const qc = useQueryClient();

  const [investOpen, setInvestOpen] = React.useState(false);
  const [selectedPlan, setSelectedPlan] = React.useState<InvestmentPlanConfig | null>(null);
  const [withdrawOpen, setWithdrawOpen] = React.useState(false);
  const [withdrawing, setWithdrawing] = React.useState(false);

  const dashboard = dashboardQ.data;
  const settings = settingsQ.data;
  const rate = rateQ.data?.usd_to_ngn ?? dashboard?.exchange_rate ?? 0;
  const investments: InvestmentSummary[] =
    dashboard?.investments?.map(normalizeInvestment) ??
    (legacyInvestmentsQ.data?.investments ?? []).map(normalizeInvestment);

  const durationDays = settings?.max_business_days ?? 20;
  const dailyReward = settings?.daily_reward_ngn ?? 0;
  const processingHours = settings?.withdrawal_processing_hours ?? 24;
  const feePercent = settings?.early_withdrawal_fee_percent ?? 0;
  const penaltyPercent = settings?.early_withdrawal_penalty_percent ?? 0;

  const available =
    dashboard?.available_balance_ngn ??
    dashboard?.referral_info?.withdrawable_balance_ngn ??
    walletQ.data?.available_balance ??
    0;
  const pendingWithdrawal = dashboard?.pending_withdrawal_ngn ?? 0;
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
  const remainingRewards = Math.max(0, (totalDays - completedDays) * dailyReward);
  const totalEarnedFromActive = activeInvestment?.roi_cdt ?? dashboard?.today_earnings_ngn ?? 0;
  const timeline = buildRewardTimeline(totalDays, dailyReward);
  const totalWithdrawn = (withdrawalsQ.data ?? []).reduce((sum, row) => sum + (row.net_amount_ngn || 0), 0);
  const growthSeries = buildInvestmentGrowthSeries(paymentsQ.data ?? [], dashboard?.total_invested_usd ?? 0);
  const referralSeries = [
    { name: "Jan", value: 0 },
    { name: "Feb", value: Math.round((dashboard?.referral_earnings_ngn ?? 0) * 0.25) },
    { name: "Mar", value: Math.round((dashboard?.referral_earnings_ngn ?? 0) * 0.5) },
    { name: "Apr", value: Math.round((dashboard?.referral_earnings_ngn ?? 0) * 0.75) },
    { name: "May", value: dashboard?.referral_earnings_ngn ?? 0 },
  ];
  const withdrawalSeries = (withdrawalsQ.data ?? []).slice(0, 6).map((row) => ({
    name: formatDate(row.created_at),
    value: row.net_amount_ngn,
  }));

  const rewards = rewardsQ.data ?? [];
  const todayReward = dashboard?.today_earnings_ngn ?? 0;
  const yesterdayReward = rewards.find((row) => {
    const day = new Date(row.reward_date || row.created_at);
    const yesterday = new Date();
    yesterday.setDate(yesterday.getDate() - 1);
    return day.toDateString() === yesterday.toDateString();
  })?.amount_ngn;
  const monthReward = dashboard?.monthly_earnings_ngn ?? 0;
  const rewardClaimedToday = todayReward > 0;

  const loading = dashboardQ.isLoading || rateQ.isLoading || settingsQ.isLoading;
  const firstName = displayName(user).split(" ")[0] || "Investor";

  const openInvest = (plan: InvestmentPlanConfig) => {
    setSelectedPlan(plan);
    setInvestOpen(true);
  };

  const submitWithdrawal = async (withdrawAmount: number) => {
    setWithdrawing(true);
    try {
      await investmentApi.requestWithdrawal(undefined, withdrawAmount);
      setWithdrawOpen(false);
      await Promise.all([
        qc.invalidateQueries({ queryKey: ["investments"] }),
        qc.invalidateQueries({ queryKey: ["wallet"] }),
      ]);
    } finally {
      setWithdrawing(false);
    }
  };

  return (
    <div className="relative space-y-6 animate-cds-fade-in pb-10">
      <div className="pointer-events-none absolute inset-x-0 -top-6 h-72 rounded-3xl bg-gradient-to-br from-[#7C3AED]/25 via-[#06B6D4]/10 to-transparent blur-2xl" />

      {/* ─── Hero Section ─────────────────────────────────── */}
      <section className="relative overflow-hidden rounded-3xl border border-primary/20 bg-gradient-to-br from-[#1a0b2e]/90 via-background to-[#0b1f33]/80 p-6 shadow-2xl backdrop-blur-xl sm:p-8">
        <div className="pointer-events-none absolute -right-20 -top-20 h-64 w-64 rounded-full bg-primary/20 blur-3xl" />
        <div className="pointer-events-none absolute -bottom-24 -left-16 h-64 w-64 rounded-full bg-cyan-500/10 blur-3xl" />

        <div className="relative flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <div className="flex items-center gap-2">
              <Badge variant="secondary" className="gap-1">
                <Sparkles className="h-3 w-3 text-primary" />
                {greetingForHour()}
              </Badge>
            </div>
            <h1 className="mt-2 text-3xl font-bold tracking-tight sm:text-4xl">
              {firstName}&apos;s Investment Portfolio
            </h1>
            <p className="mt-2 max-w-2xl text-sm text-muted-foreground">
              Choose a Genesis Investment Plan and start earning daily rewards.
            </p>
          </div>
          <Button asChild variant="outline" size="sm" className="backdrop-blur">
            <Link href="/app/dashboard">
              <ArrowRight className="mr-2 h-4 w-4" /> Dashboard
            </Link>
          </Button>
        </div>

        <div className="relative mt-6 grid gap-3 sm:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-6">
          {loading ? (
            Array.from({ length: 6 }).map((_, i) => <Skeleton key={i} className="h-28 rounded-2xl" />)
          ) : (
            <>
              <MetricCard
                title="Total Invested"
                value={`$${dashboard?.total_invested_usd ?? 0}`}
                numeric={dashboard?.total_invested_usd ?? 0}
                prefix="$"
                icon={PiggyBank}
                accent="border-violet-500/30"
                hint="Lifetime capital"
              />
              <MetricCard
                title="Total Withdrawn"
                value={formatCurrency(totalWithdrawn)}
                numeric={totalWithdrawn}
                prefix="₦"
                icon={Wallet}
                accent="border-fuchsia-500/30"
                hint="Completed cash-outs"
              />
              <MetricCard title="Today&apos;s Reward"
                value={formatCurrency(todayReward || dailyReward)}
                numeric={todayReward || dailyReward}
                prefix="₦"
                icon={Sparkles}
                accent="border-amber-400/30"
                hint="Business day credit"
              />
              <MetricCard
                title="Total Earned"
                value={formatCurrency(totalEarnedFromActive || monthReward)}
                numeric={totalEarnedFromActive || monthReward}
                prefix="₦"
                icon={TrendingUp}
                accent="border-emerald-400/30"
                hint="Rewards credited"
              />
              <MetricCard
                title="Referral Earnings"
                value={formatCurrency(dashboard?.referral_earnings_ngn)}
                numeric={dashboard?.referral_earnings_ngn ?? 0}
                prefix="₦"
                icon={Gift}
                accent="border-cyan-400/30"
                hint="Lifetime referrals"
              />
              <MetricCard
                title="Available Balance"
                value={formatCurrency(available)}
                numeric={available}
                prefix="₦"
                icon={Landmark}
                accent="border-primary/40"
                hint="Withdrawable"
              />
            </>
          )}
        </div>
      </section>

      {/* ─── Genesis Investment Plans ─────────────────────── */}
      <section>
        <div className="mb-6">
          <h2 className="text-2xl font-bold tracking-tight">Genesis Investment Plans</h2>
          <p className="mt-1 text-sm text-muted-foreground">
            Select a fixed plan to start earning. No custom amounts.
          </p>
        </div>
        <InvestmentGrid
          plans={ENABLED_PLANS}
          exchangeRate={rate}
          onSelectPlan={openInvest}
        />
      </section>

      {/* ─── Summary Cards ───────────────────────────────── */}
      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <MetricCard title="Active Investments" value={String(dashboard?.active_investments ?? investments.filter((item) => item.status === "active").length)} numeric={dashboard?.active_investments ?? investments.filter((item) => item.status === "active").length} prefix="" icon={Wallet} accent="border-primary/40" hint="Live positions" />
        <MetricCard title="Today&apos;s Earnings" value={formatCurrency(todayReward)} numeric={todayReward} prefix="₦" icon={Sparkles} accent="border-amber-400/30" hint="Business day payout" />
        <MetricCard title="Pending Withdrawal" value={formatCurrency(pendingWithdrawal)} numeric={pendingWithdrawal} prefix="₦" icon={Clock3} accent="border-fuchsia-400/30" hint="Awaiting review" />
        <MetricCard title="Processing Time" value={`${processingHours}h`} numeric={processingHours} prefix="" icon={CreditCard} accent="border-cyan-400/30" hint="Withdrawal processing" />
      </div>

      {/* ─── Progress + Daily Rewards ─────────────────────── */}
      <div className="grid gap-6 lg:grid-cols-2">
        <Card className="border-primary/15">
          <CardHeader>
            <CardTitle className="text-base">Investment Progress</CardTitle>
            <CardDescription>
              {activeInvestment
                ? `Day ${completedDays} / ${totalDays}`
                : "Start an investment to unlock live progress tracking."}
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <div className="mb-2 flex justify-between text-sm">
                <span className="text-muted-foreground">Progress</span>
                <span className="font-semibold">{Math.round(progress)}%</span>
              </div>
              <Progress value={progress} className="h-3" />
            </div>
            <div className="grid grid-cols-2 gap-3 text-sm">
              <div className="rounded-xl border p-3">
                <p className="text-xs text-muted-foreground">Remaining Days</p>
                <p className="mt-1 text-lg font-semibold">{activeInvestment ? daysRemaining : "—"}</p>
              </div>
              <div className="rounded-xl border p-3">
                <p className="text-xs text-muted-foreground">Expected Maturity</p>
                <p className="mt-1 text-lg font-semibold">{formatDate(activeInvestment?.matures_at)}</p>
              </div>
              <div className="rounded-xl border p-3">
                <p className="text-xs text-muted-foreground">Total Earned</p>
                <p className="mt-1 text-lg font-semibold">{formatCurrency(totalEarnedFromActive)}</p>
              </div>
              <div className="rounded-xl border p-3">
                <p className="text-xs text-muted-foreground">Remaining Rewards</p>
                <p className="mt-1 text-lg font-semibold">
                  {formatCurrency(activeInvestment ? remainingRewards : 0)}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card className="border-amber-500/20">
          <CardHeader>
            <CardTitle className="text-base">Daily Rewards</CardTitle>
            <CardDescription>Track today&apos;s credit and upcoming payouts.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3 text-sm">
            <div className="grid grid-cols-2 gap-3">
              <div className="rounded-xl border bg-amber-500/5 p-3">
                <p className="text-xs text-muted-foreground">Today&apos;s Reward</p>
                <p className="mt-1 text-xl font-bold">{formatCurrency(todayReward || dailyReward)}</p>
              </div>
              <div className="rounded-xl border p-3">
                <p className="text-xs text-muted-foreground">Yesterday</p>
                <p className="mt-1 text-xl font-bold">
                  {yesterdayReward != null ? `+${formatCurrency(yesterdayReward)}` : `+${formatCurrency(dailyReward)}`}
                </p>
              </div>
              <div className="rounded-xl border p-3">
                <p className="text-xs text-muted-foreground">This Month</p>
                <p className="mt-1 text-xl font-bold">{formatCurrency(monthReward)}</p>
              </div>
              <div className="rounded-xl border p-3">
                <p className="text-xs text-muted-foreground">Upcoming Reward</p>
                <p className="mt-1 font-semibold">Tomorrow</p>
                <p className="text-lg font-bold">{formatCurrency(dailyReward)}</p>
              </div>
            </div>
            {rewardClaimedToday ? (
              <p className="flex items-center gap-2 rounded-lg bg-emerald-500/10 px-3 py-2 text-emerald-700 dark:text-emerald-300">
                <CheckCircle2 className="h-4 w-4" /> Today&apos;s reward credited
              </p>
            ) : (
              <p className="text-xs text-muted-foreground">
                Rewards credit on business days once your investment is active.
              </p>
            )}
          </CardContent>
        </Card>
      </div>

      {/* ─── Charts ─────────────────────────────────────── */}
      <div className="grid gap-6 xl:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Daily Earnings</CardTitle>
            <CardDescription>Reward cadence across your current plan.</CardDescription>
          </CardHeader>
          <CardContent className="h-72">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={timeline.map((item) => ({ name: `Day ${item.day}`, value: item.amount }))}>
                <CartesianGrid strokeDasharray="3 3" stroke="hsl(var(--border))" />
                <XAxis dataKey="name" tickLine={false} axisLine={false} />
                <YAxis tickLine={false} axisLine={false} />
                <Tooltip formatter={(value) => formatCurrency(Number(value) || 0)} />
                <Area type="monotone" dataKey="value" stroke="hsl(var(--primary))" fill="hsl(var(--primary) / 0.15)" />
              </AreaChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-base">Investment Growth</CardTitle>
            <CardDescription>Accumulated capital growth from your deposits.</CardDescription>
          </CardHeader>
          <CardContent className="h-72">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={growthSeries}>
                <CartesianGrid strokeDasharray="3 3" stroke="hsl(var(--border))" />
                <XAxis dataKey="name" tickLine={false} axisLine={false} />
                <YAxis tickLine={false} axisLine={false} />
                <Tooltip formatter={(value) => `$${Number(value || 0).toLocaleString()}`} />
                <Area type="monotone" dataKey="value" stroke="#7C3AED" fill="#7C3AED22" />
              </AreaChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-6 xl:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Referral Growth</CardTitle>
            <CardDescription>Your referral rewards over time.</CardDescription>
          </CardHeader>
          <CardContent className="h-72">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={referralSeries}>
                <CartesianGrid strokeDasharray="3 3" stroke="hsl(var(--border))" />
                <XAxis dataKey="name" tickLine={false} axisLine={false} />
                <YAxis tickLine={false} axisLine={false} />
                <Tooltip formatter={(value) => formatCurrency(Number(value) || 0)} />
                <Area type="monotone" dataKey="value" stroke="#06B6D4" fill="#06B6D422" />
              </AreaChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-base">Withdrawal History</CardTitle>
            <CardDescription>Recent net withdrawals and processing activity.</CardDescription>
          </CardHeader>
          <CardContent className="h-72">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={withdrawalSeries.length ? withdrawalSeries : [{ name: "No data", value: 0 }]}>
                <CartesianGrid strokeDasharray="3 3" stroke="hsl(var(--border))" />
                <XAxis dataKey="name" tickLine={false} axisLine={false} />
                <YAxis tickLine={false} axisLine={false} />
                <Tooltip formatter={(value) => formatCurrency(Number(value) || 0)} />
                <Area type="monotone" dataKey="value" stroke="#F59E0B" fill="#F59E0B22" />
              </AreaChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>
      </div>

      {/* ─── Timeline ─────────────────────────────────────── */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Investment Timeline</CardTitle>
          <CardDescription>Completed days, current day, remaining days, then withdrawal.</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex gap-2 overflow-x-auto pb-2">
            {timeline.map((item) => {
              const done = activeInvestment ? item.day <= completedDays : false;
              const current = activeInvestment ? item.day === completedDays + 1 : item.day === 1;
              return (
                <div
                  key={item.day}
                  className={`min-w-[4.5rem] rounded-xl border px-3 py-3 text-center text-xs transition ${
                    done
                      ? "border-emerald-500/40 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300"
                      : current
                        ? "border-primary bg-primary/15 text-primary shadow-md shadow-primary/20"
                        : "border-border/60 text-muted-foreground"
                  }`}
                >
                  <p className="font-semibold">Day {item.day}</p>
                  <p className="mt-1">{done ? "✔" : current ? "●" : "○"}</p>
                </div>
              );
            })}
            <div className="min-w-[5.5rem] rounded-xl border border-cyan-500/40 bg-cyan-500/10 px-3 py-3 text-center text-xs text-cyan-700 dark:text-cyan-300">
              <p className="font-semibold">Withdrawal</p>
              <p className="mt-1">Day {totalDays}+</p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* ─── Investment Modal ─────────────────────────── */}
      <InvestmentModal
        open={investOpen}
        plan={selectedPlan}
        exchangeRate={rate}
        onClose={() => setInvestOpen(false)}
      />

      {/* ─── Withdrawal Modal ─────────────────────────── */}
      <WithdrawalRequestModal
        open={withdrawOpen}
        availableBalance={available}
        processingHours={processingHours}
        feePercent={feePercent}
        penaltyPercent={penaltyPercent}
        earlyWithdrawal={early}
        isSubmitting={withdrawing}
        onClose={() => setWithdrawOpen(false)}
        onConfirm={submitWithdrawal}
      />
    </div>
  );
}