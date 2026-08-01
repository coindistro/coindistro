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
  Input,
  Label,
  Progress,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
  Skeleton,
  Switch,
} from "@coindistro/cds";
import {
  ArrowRight,
  ArrowUpRight,
  CalendarDays,
  CheckCircle2,
  Clock3,
  Copy,
  CreditCard,
  Gift,
  Landmark,
  Lock,
  PiggyBank,
  ShieldCheck,
  Settings2,
  Sparkles,
  TrendingUp,
  Wallet,
  Zap,
} from "lucide-react";
import { Area, AreaChart, CartesianGrid, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts";
import { useAuth } from "@/features/authentication/auth-provider";
import { useInvestments, useInvestmentPlans, useWallet } from "@/features/earn/hooks";
import { InvestmentPaymentModal } from "@/features/earn/investment-payment-modal";
import { WithdrawalRequestModal } from "@/features/earn/withdrawal-request-modal";
import { useToast } from "@/features/shared/providers/toast-provider";
import {
  buildInvestmentGrowthSeries,
  buildRewardTimeline,
  calculateInvestment,
  calculateWithdrawal,
  deriveRoiPercent,
  formatCurrency,
  formatRoi,
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
import type { InvestmentPlan, InvestmentSummary } from "@/lib/api/types";
import type { EarningsSummary } from "@/features/investments/types";
import { displayName } from "@/lib/utils/format";

const PREFS_KEY = "coindistro.earn.preferences";

interface EarnPreferences {
  notifyDailyRewards: boolean;
  notifyWithdrawals: boolean;
  notifyReferrals: boolean;
  preferredPaymentMethod: "paystack" | "flutterwave";
  autoReinvest: boolean;
}

const defaultPrefs: EarnPreferences = {
  notifyDailyRewards: true,
  notifyWithdrawals: true,
  notifyReferrals: true,
  preferredPaymentMethod: "paystack",
  autoReinvest: false,
};

function formatDate(value?: string | null) {
  if (!value) return "—";
  return new Date(value).toLocaleDateString(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

function statusVariant(
  status: string,
): "success" | "warning" | "secondary" | "danger" | "outline" {
  if (status === "active" || status === "completed" || status === "paid") {
    return status === "active" || status === "paid" ? "success" : "secondary";
  }
  if (
    status === "pending" ||
    status === "pending_payment" ||
    status === "pending_review" ||
    status === "processing"
  ) {
    return "warning";
  }
  if (status === "failed" || status === "cancelled" || status === "rejected") {
    return "danger";
  }
  return "outline";
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
    roi_percent: deriveRoiPercent(earnedNgn, amountNgn),
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

function loadPreferences(): EarnPreferences {
  if (typeof window === "undefined") return defaultPrefs;
  try {
    const raw = window.localStorage.getItem(PREFS_KEY);
    if (!raw) return defaultPrefs;
    return { ...defaultPrefs, ...(JSON.parse(raw) as Partial<EarnPreferences>) };
  } catch {
    return defaultPrefs;
  }
}

function useAnimatedNumber(value: number, enabled = true) {
  const [display, setDisplay] = React.useState(value || 0);
  const displayRef = React.useRef(display);
  displayRef.current = display;

  React.useEffect(() => {
    if (!enabled || !Number.isFinite(value)) {
      setDisplay(value || 0);
      return;
    }
    let frame = 0;
    const start = displayRef.current;
    const delta = value - start;
    const duration = 700;
    const started = performance.now();
    const tick = (now: number) => {
      const progress = Math.min(1, (now - started) / duration);
      const eased = 1 - Math.pow(1 - progress, 3);
      setDisplay(start + delta * eased);
      if (progress < 1) frame = requestAnimationFrame(tick);
    };
    frame = requestAnimationFrame(tick);
    return () => cancelAnimationFrame(frame);
  }, [value, enabled]);

  return display;
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
  const animated = useAnimatedNumber(numeric ?? 0, numeric != null && !loading);
  const shown =
    loading
      ? "…"
      : numeric != null
        ? `${prefix}${new Intl.NumberFormat("en-NG", { maximumFractionDigits: 2 }).format(animated)}${suffix}`
        : value;

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

function StatTile({
  label,
  value,
  icon: Icon,
  accent = "text-primary",
}: {
  label: string;
  value: React.ReactNode;
  icon: React.ComponentType<{ className?: string }>;
  accent?: string;
}) {
  return (
    <div className="rounded-xl border border-border/60 bg-gradient-to-br from-muted/40 to-transparent px-4 py-3 transition hover:border-primary/30">
      <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
        <Icon className={`h-3.5 w-3.5 ${accent}`} />
        <span>{label}</span>
      </div>
      <p className="mt-1 text-lg font-semibold tabular-nums">{value}</p>
    </div>
  );
}

export function EarnDashboard() {
  const { user } = useAuth();
  const walletQ = useWallet();
  const legacyInvestmentsQ = useInvestments();
  const plansQ = useInvestmentPlans();
  const dashboardQ = useDashboard();
  const rateQ = useExchangeRate();
  const settingsQ = useInvestmentSettings();
  const paymentsQ = usePaymentHistory();
  const withdrawalsQ = useWithdrawalHistory();
  const rewardsQ = useRewardHistory();
  const qc = useQueryClient();
  const { toast } = useToast();

  const [investOpen, setInvestOpen] = React.useState(false);
  const [selectedPlan, setSelectedPlan] = React.useState<InvestmentPlan | null>(null);
  const [calcAmount, setCalcAmount] = React.useState("");
  const [paying, setPaying] = React.useState(false);
  const [withdrawOpen, setWithdrawOpen] = React.useState(false);
  const [withdrawing, setWithdrawing] = React.useState(false);
  const [copied, setCopied] = React.useState(false);
  const [prefs, setPrefs] = React.useState<EarnPreferences>(defaultPrefs);
  const [paymentError, setPaymentError] = React.useState<string | null>(null);
  const [settingsOpen, setSettingsOpen] = React.useState(false);
  const [provider, setProvider] = React.useState<"paystack" | "flutterwave">("paystack");

  React.useEffect(() => {
    const loaded = loadPreferences();
    setPrefs(loaded);
    setProvider(loaded.preferredPaymentMethod);
  }, []);

  const dashboard = dashboardQ.data;
  const settings = settingsQ.data;
  const rate = rateQ.data?.usd_to_ngn ?? dashboard?.exchange_rate ?? 0;
  const plans = Array.isArray(plansQ.data) ? plansQ.data : [];
  const investments: InvestmentSummary[] =
    dashboard?.investments?.map(normalizeInvestment) ??
    (legacyInvestmentsQ.data?.investments ?? []).map(normalizeInvestment);

  const minUsd = settings?.minimum_investment_usd ?? 30;
  const durationDays = settings?.max_business_days ?? 20;
  const dailyReward = settings?.daily_reward_ngn ?? 0;
  const settingsRoi = settings?.roi_percent ?? 0;
  const referralPercent = settings?.referral_percent ?? 0;
  const minReferrals = settings?.min_referrals_for_payout ?? 5;
  const processingHours = settings?.withdrawal_processing_hours ?? 24;
  const feePercent = settings?.early_withdrawal_fee_percent ?? 0;
  const penaltyPercent = settings?.early_withdrawal_penalty_percent ?? 0;

  React.useEffect(() => {
    if (!calcAmount && minUsd) setCalcAmount(String(minUsd));
  }, [minUsd, calcAmount]);

  const amountUsd = Math.max(Number(calcAmount) || 0, 0);
  const calc = calculateInvestment({
    amountUsd: Math.max(amountUsd, minUsd),
    exchangeRate: rate,
    dailyRewardNgn: dailyReward,
    durationBusinessDays: durationDays,
    roiPercent: settingsRoi,
  });

  const available =
    dashboard?.available_balance_ngn ??
    dashboard?.referral_info?.withdrawable_balance_ngn ??
    walletQ.data?.available_balance ??
    0;
  const locked = dashboard?.total_invested_ngn ?? walletQ.data?.locked_balance ?? 0;
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
  const withdrawalPreview = calculateWithdrawal(available, feePercent, penaltyPercent, early);
  const referralTargetEarnings = calc.amountNgn * (referralPercent / 100) * minReferrals;
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

  const savePrefs = (next: EarnPreferences) => {
    setPrefs(next);
    if (typeof window !== "undefined") {
      window.localStorage.setItem(PREFS_KEY, JSON.stringify(next));
    }
  };

  const openInvest = (plan?: InvestmentPlan | null) => {
    const resolved =
      plan ??
      plans.find((item) => item.enabled) ??
      ({
        id: "default",
        name: "CoinDistro Plan",
        minimum_amount: Math.max(amountUsd || minUsd, minUsd),
        maximum_amount: minUsd * 100,
        currency: "USD",
        roi_percent: calc.roiPercent,
        enabled: true,
      } satisfies InvestmentPlan);
    setSelectedPlan({
      ...resolved,
      minimum_amount: Math.max(amountUsd || resolved.minimum_amount || minUsd, minUsd),
    });
    setInvestOpen(true);
  };

  const createPayment = async (paymentProvider: "paystack" | "flutterwave", investAmount: number) => {
    setPaying(true);
    setPaymentError(null);
    try {
      const result =
        paymentProvider === "paystack"
          ? await investmentApi.initPaystackPayment(investAmount)
          : await investmentApi.initFlutterwavePayment(investAmount);
      const authorizationUrl =
        (result as { authorization_url?: string; authorizationUrl?: string }).authorization_url ??
        (result as { authorization_url?: string; authorizationUrl?: string }).authorizationUrl;
      if (!authorizationUrl) {
        throw new Error("Payment provider did not return a checkout URL");
      }
      setInvestOpen(false);
      setSelectedPlan(null);
      await Promise.all([
        qc.invalidateQueries({ queryKey: ["investments"] }),
        qc.invalidateQueries({ queryKey: ["wallet"] }),
        qc.invalidateQueries({ queryKey: ["investments", "dashboard"] }),
      ]);
      toast({
        message: `Your ${paymentProvider === "paystack" ? "Paystack" : "Flutterwave"} checkout is ready. Complete payment to activate your investment.`,
        variant: "success",
      });
      window.location.assign(authorizationUrl);
    } catch (error) {
      setPaymentError(error instanceof Error ? error.message : "Unable to start payment");
    } finally {
      setPaying(false);
    }
  };

  const copyReferral = async () => {
    const link = dashboard?.referral_info?.referral_link;
    if (!link) return;
    await navigator.clipboard.writeText(link);
    setCopied(true);
    window.setTimeout(() => setCopied(false), 1500);
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

  const activity = [
    ...(paymentsQ.data ?? []).slice(0, 3).map((row) => ({
      id: `pay-${row.id}`,
      title: "Investment Created",
      detail: `+$${row.amount_usd ?? 0}`,
      status: row.status,
      at: row.created_at,
    })),
    ...(rewardsQ.data ?? []).slice(0, 3).map((row) => ({
      id: `rew-${row.id}`,
      title: "Reward Credited",
      detail: `+${formatCurrency(row.amount_ngn)}`,
      status: row.status,
      at: row.created_at,
    })),
    ...(withdrawalsQ.data ?? []).slice(0, 2).map((row) => ({
      id: `wd-${row.id}`,
      title: "Withdrawal Requested",
      detail: formatCurrency(row.net_amount_ngn),
      status: row.status,
      at: row.created_at,
    })),
  ]
    .sort((a, b) => new Date(b.at).getTime() - new Date(a.at).getTime())
    .slice(0, 8);

  if (dashboard?.referral_earnings_ngn) {
    activity.unshift({
      id: "referral-lifetime",
      title: "Referral Bonus",
      detail: `+${formatCurrency(dashboard.referral_earnings_ngn)}`,
      status: "paid",
      at: new Date().toISOString(),
    });
  }

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
              {firstName}’s Investment Portfolio
            </h1>
            <p className="mt-2 max-w-2xl text-sm text-muted-foreground">
              Track capital, daily rewards, and withdrawals in one premium investment workspace.
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
              <MetricCard title="Today’s Reward"
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
                title="Current ROI"
                value={formatRoi(calc.roiPercent)}
                numeric={calc.roiPercent}
                suffix="%"
                icon={Zap}
                accent="border-primary/40"
                hint={`${durationDays} business days`}
              />
            </>
          )}
        </div>
      </section>

      {/* ─── Calculator + Exchange Rate ───────────────────── */}
      <div className="grid gap-6 xl:grid-cols-[1.35fr_1fr]">
        <Card className="border-primary/15 bg-card/80 shadow-xl backdrop-blur">
          <CardHeader>
            <div className="flex items-center gap-2">
              <Landmark className="h-5 w-5 text-primary" />
              <CardTitle className="text-lg">Investment Calculator</CardTitle>
            </div>
            <CardDescription>
              Live projections using your configured rate, daily reward, and duration. Minimum ${minUsd}.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-5">
            <div className="space-y-2">
              <Label htmlFor="calc-amount">Investment Amount (USD)</Label>
              <div className="flex flex-wrap gap-2">
                {[30, 50, 100, 200, 500, 1000].map((preset) => (
                  <Button
                    key={preset}
                    type="button"
                    variant={Number(calcAmount) === preset ? "primary" : "outline"}
                    size="sm"
                    onClick={() => setCalcAmount(String(preset))}
                  >
                    ${preset}
                  </Button>
                ))}
                <Button
                  type="button"
                  variant={![30, 50, 100, 200, 500, 1000].includes(Number(calcAmount)) ? "primary" : "outline"}
                  size="sm"
                  onClick={() => setCalcAmount(String(Math.max(amountUsd || minUsd, minUsd)))}
                >
                  Custom
                </Button>
              </div>
              <div className="relative">
                <span className="absolute left-3 top-1/2 -translate-y-1/2 text-lg font-semibold text-muted-foreground">$</span>
                <Input
                  id="calc-amount"
                  type="number"
                  min={minUsd}
                  step="1"
                  value={calcAmount}
                  onChange={(e) => setCalcAmount(e.target.value)}
                  className="h-12 pl-8 text-lg font-semibold"
                />
              </div>
              {amountUsd > 0 && amountUsd < minUsd ? (
                <p className="text-xs text-amber-600">Minimum investment is ${minUsd}.</p>
              ) : null}
            </div>

            <div className="rounded-xl border border-border/60 bg-muted/40 p-4">
              <div className="grid gap-3 sm:grid-cols-2">
                <div className="rounded-lg border bg-background/70 p-3">
                  <p className="text-xs text-muted-foreground">Investment</p>
                  <p className="mt-1 font-semibold">${Math.max(amountUsd, minUsd).toLocaleString()}</p>
                </div>
                <div className="rounded-lg border bg-background/70 p-3">
                  <p className="text-xs text-muted-foreground">Exchange rate</p>
                  <p className="mt-1 font-semibold">1 USD = {loading || !rate ? "…" : formatCurrency(rate)}</p>
                </div>
                <div className="rounded-lg border bg-background/70 p-3">
                  <p className="text-xs text-muted-foreground">NGN equivalent</p>
                  <p className="mt-1 font-semibold">{loading ? "…" : formatCurrency(calc.amountNgn)}</p>
                </div>
                <div className="rounded-lg border bg-background/70 p-3">
                  <p className="text-xs text-muted-foreground">Daily earnings</p>
                  <p className="mt-1 font-semibold">{loading ? "…" : formatCurrency(calc.dailyEarningsNgn)}</p>
                </div>
                <div className="rounded-lg border bg-background/70 p-3">
                  <p className="text-xs text-muted-foreground">Business days remaining</p>
                  <p className="mt-1 font-semibold">{loading ? "…" : String(calc.businessDaysRemaining)}</p>
                </div>
                <div className="rounded-lg border bg-background/70 p-3">
                  <p className="text-xs text-muted-foreground">Monthly earnings</p>
                  <p className="mt-1 font-semibold">{loading ? "…" : formatCurrency(calc.monthlyEarningsNgn)}</p>
                </div>
                <div className="rounded-lg border bg-background/70 p-3">
                  <p className="text-xs text-muted-foreground">Referral bonus potential</p>
                  <p className="mt-1 font-semibold">{loading ? "…" : formatCurrency(referralTargetEarnings)}</p>
                </div>
                <div className="rounded-lg border bg-background/70 p-3">
                  <p className="text-xs text-muted-foreground">Expected ROI</p>
                  <p className="mt-1 font-semibold">{loading ? "…" : formatRoi(calc.roiPercent)}</p>
                </div>
                <div className="rounded-lg border bg-background/70 p-3">
                  <p className="text-xs text-muted-foreground">Expected total payout</p>
                  <p className="mt-1 font-semibold">{loading ? "…" : formatCurrency(calc.totalPayoutNgn)}</p>
                </div>
                <div className="rounded-lg border bg-background/70 p-3">
                  <p className="text-xs text-muted-foreground">Processing time</p>
                  <p className="mt-1 font-semibold">{processingHours} hours</p>
                </div>
                <div className="rounded-lg border bg-background/70 p-3">
                  <p className="text-xs text-muted-foreground">Early withdrawal penalty</p>
                  <p className="mt-1 font-semibold">{feePercent + penaltyPercent}%</p>
                </div>
                <div className="rounded-lg border bg-background/70 p-3">
                  <p className="text-xs text-muted-foreground">Maturity date</p>
                  <p className="mt-1 font-semibold">{formatDate(activeInvestment?.matures_at)}</p>
                </div>
              </div>
            </div>

            <div className="grid gap-3 sm:grid-cols-2">
              <StatTile label="USD" value={`$${(Math.max(amountUsd, minUsd) || 0).toLocaleString()}`} icon={Wallet} />
              <StatTile label="NGN Equivalent" value={loading ? "…" : formatCurrency(calc.amountNgn)} icon={Landmark} accent="text-fuchsia-500" />
              <StatTile label="Daily Earnings" value={loading ? "…" : formatCurrency(calc.dailyEarningsNgn)} icon={Sparkles} accent="text-amber-500" />
              <StatTile label="Monthly Earnings" value={loading ? "…" : formatCurrency(calc.monthlyEarningsNgn)} icon={TrendingUp} accent="text-emerald-500" />
              <StatTile label="ROI %" value={loading ? "…" : formatRoi(calc.roiPercent)} icon={Zap} accent="text-primary" />
              <StatTile label="Total Withdrawal" value={loading ? "…" : formatCurrency(calc.totalPayoutNgn)} icon={Wallet} accent="text-cyan-500" />
              <StatTile label="Business Days Remaining" value={loading ? "…" : String(calc.businessDaysRemaining)} icon={CalendarDays} accent="text-violet-500" />
            </div>

            <div className="flex flex-col gap-2 sm:flex-row">
              <Button className="flex-1" size="lg" onClick={() => openInvest(null)}>
                <ArrowUpRight className="mr-2 h-4 w-4" /> Invest Now
              </Button>
              <Button variant="outline" size="lg" onClick={() => setWithdrawOpen(true)}>
                <Lock className="mr-2 h-4 w-4" /> Withdraw
              </Button>
            </div>
          </CardContent>
        </Card>

        {/* Exchange Rate + Trust */}
        <div className="space-y-6">
          <Card className="overflow-hidden border-cyan-500/20 bg-gradient-to-br from-cyan-500/10 via-card to-primary/10">
            <CardContent className="space-y-3 p-6">
              <div className="flex items-center justify-between">
                <Badge variant="secondary" className="gap-1">
                  <Zap className="h-3 w-3 text-cyan-500" /> Live Exchange Rate
                </Badge>
                <Badge variant="outline" className="gap-1">
                  <CheckCircle2 className="h-3 w-3 text-emerald-500" /> Updated Today
                </Badge>
              </div>
              <div className="flex items-end gap-3">
                <div>
                  <p className="text-sm text-muted-foreground">1 USD</p>
                  <p className="text-3xl font-bold">=</p>
                </div>
                <p className="text-4xl font-bold tabular-nums text-primary">
                  {loading || !rate ? "…" : formatCurrency(rate)}
                </p>
              </div>
              <p className="text-xs text-muted-foreground">
                Updated {rateQ.data?.updated_at ? formatDate(rateQ.data.updated_at) : "Today"}
              </p>
            </CardContent>
          </Card>

          <Card className="border-emerald-500/20">
            <CardHeader>
              <CardTitle className="text-base">Why investors trust CoinDistro</CardTitle>
            </CardHeader>
            <CardContent className="grid gap-3 text-sm">
              {[
                { label: "Capital Protected", icon: ShieldCheck },
                { label: "Daily Rewards", icon: Sparkles },
                { label: "24hr Withdrawal Processing", icon: Clock3 },
                { label: "Secure Payments", icon: CreditCard },
              ].map(({ label, icon: Icon }) => (
                <div key={label} className="flex items-center gap-2 rounded-lg bg-emerald-500/5 px-3 py-2 transition hover:bg-emerald-500/10">
                  <Icon className="h-4 w-4 text-emerald-500" />
                  <span>{label}</span>
                </div>
              ))}
            </CardContent>
          </Card>
        </div>
      </div>

      {/* ─── Summary Cards ───────────────────────────────── */}
      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <MetricCard title="Active Investments" value={String(dashboard?.active_investments ?? investments.filter((item) => item.status === "active").length)} numeric={dashboard?.active_investments ?? investments.filter((item) => item.status === "active").length} prefix="" icon={Wallet} accent="border-primary/40" hint="Live positions" />
        <MetricCard title="Today’s Earnings" value={formatCurrency(todayReward)} numeric={todayReward} prefix="₦" icon={Sparkles} accent="border-amber-400/30" hint="Business day payout" />
        <MetricCard title="Available Balance" value={formatCurrency(available)} numeric={available} prefix="₦" icon={Landmark} accent="border-cyan-400/30" hint="Withdrawable" />
        <MetricCard title="Pending Withdrawal" value={formatCurrency(pendingWithdrawal)} numeric={pendingWithdrawal} prefix="₦" icon={Clock3} accent="border-fuchsia-400/30" hint="Awaiting review" />
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
            <CardDescription>Track today’s credit and upcoming payouts.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3 text-sm">
            <div className="grid grid-cols-2 gap-3">
              <div className="rounded-xl border bg-amber-500/5 p-3">
                <p className="text-xs text-muted-foreground">Today’s Reward</p>
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
                <CheckCircle2 className="h-4 w-4" /> Today’s reward credited
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
              <AreaChart data={withdrawalSeries.length ? withdrawalSeries : [{ name: "No data", value: 0 }] }>
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

      {/* ─── Payment + Withdrawal ─────────────────────────── */}
      <div className="grid gap-6 lg:grid-cols-2">
        <Card className="border-primary/20">
          <CardHeader>
            <CardTitle className="text-base">Choose Payment Method</CardTitle>
            <CardDescription>Secure checkout with Paystack or Flutterwave.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid gap-3 sm:grid-cols-2">
              {(
                [
                  {
                    id: "paystack" as const,
                    name: "Paystack",
                    description: "Cards, bank transfer & USSD",
                    speed: "Instant redirect",
                    recommended: true,
                  },
                  {
                    id: "flutterwave" as const,
                    name: "Flutterwave",
                    description: "Cards & local payment rails",
                    speed: "Fast checkout",
                    recommended: false,
                  },
                ] as const
              ).map((item) => {
                const selected = provider === item.id;
                return (
                  <button
                    key={item.id}
                    type="button"
                    onClick={() => {
                      setProvider(item.id);
                      savePrefs({ ...prefs, preferredPaymentMethod: item.id });
                    }}
                    className={`rounded-2xl border p-4 text-left transition ${
                      selected
                        ? "border-primary bg-primary/10 shadow-lg shadow-primary/10"
                        : "border-border/70 hover:border-primary/40"
                    }`}
                  >
                    <div className="flex items-start justify-between gap-2">
                      <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-gradient-to-br from-[#7C3AED] to-[#06B6D4] text-sm font-bold text-white">
                        {item.name.slice(0, 1)}
                      </div>
                      {item.recommended ? <Badge variant="success">Recommended</Badge> : null}
                    </div>
                    <p className="mt-3 font-semibold">{item.name}</p>
                    <p className="mt-1 text-xs text-muted-foreground">{item.description}</p>
                    <p className="mt-2 text-xs font-medium text-primary">{item.speed}</p>
                  </button>
                );
              })}
            </div>
            <Button className="w-full" size="lg" onClick={() => openInvest(null)}>
              Invest Now · {provider === "paystack" ? "Paystack" : "Flutterwave"}
            </Button>
            {paymentError ? (
              <p className="text-sm text-destructive" role="alert">
                {paymentError}
              </p>
            ) : null}
          </CardContent>
        </Card>

        <Card className="border-amber-500/20">
          <CardHeader>
            <CardTitle className="text-base">Withdrawal</CardTitle>
            <CardDescription>Processing time: {processingHours} Hours</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4 text-sm">
            <div className="grid grid-cols-2 gap-3">
              <div className="rounded-xl border p-3">
                <p className="text-xs text-muted-foreground">Available Balance</p>
                <p className="mt-1 text-lg font-semibold">{formatCurrency(available)}</p>
              </div>
              <div className="rounded-xl border p-3">
                <p className="text-xs text-muted-foreground">Locked Capital</p>
                <p className="mt-1 text-lg font-semibold">{formatCurrency(locked)}</p>
              </div>
              <div className="rounded-xl border p-3">
                <p className="text-xs text-muted-foreground">Pending Withdrawal</p>
                <p className="mt-1 text-lg font-semibold">{formatCurrency(pendingWithdrawal)}</p>
              </div>
              <div className="rounded-xl border p-3">
                <p className="text-xs text-muted-foreground">Withdrawal Fee</p>
                <p className="mt-1 text-lg font-semibold">{formatCurrency(withdrawalPreview.fee)}</p>
              </div>
            </div>
            {early ? (
              <p className="rounded-lg border border-amber-500/30 bg-amber-500/10 p-3 text-amber-800 dark:text-amber-200">
                Early withdrawal incurs a penalty. Estimated deduction:{" "}
                <strong>{formatCurrency(withdrawalPreview.deductions)}</strong>
              </p>
            ) : null}
            <Button className="w-full" variant="outline" onClick={() => setWithdrawOpen(true)}>
              <Lock className="mr-2 h-4 w-4" /> Request Withdrawal
            </Button>
          </CardContent>
        </Card>
      </div>

      {/* ─── Referral + Activity ──────────────────────────── */}
      <div className="grid gap-6 lg:grid-cols-2">
        <Card className="border-fuchsia-500/20 bg-gradient-to-br from-fuchsia-500/5 to-transparent">
          <CardHeader>
            <CardTitle className="text-base">Referral Dashboard</CardTitle>
            <CardDescription>
              Invite {minReferrals} Friends · Earn {referralPercent}% · ≈ {formatCurrency(referralTargetEarnings)}
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
              <div className="flex-1 rounded-xl border bg-background/70 px-4 py-3 font-mono text-sm">
                {dashboard?.referral_info?.referral_code ?? user?.referral_code ?? "—"}
              </div>
              <Button variant="outline" onClick={() => void copyReferral()}>
                <Copy className="mr-2 h-4 w-4" />
                {copied ? "Copied" : "Copy"}
              </Button>
            </div>
            <div className="grid grid-cols-2 gap-3 text-sm">
              <div className="rounded-xl border p-3">
                <p className="text-xs text-muted-foreground">Invited Users</p>
                <p className="mt-1 text-lg font-semibold">
                  {dashboard?.referral_info?.total_referrals ?? 0}
                </p>
              </div>
              <div className="rounded-xl border p-3">
                <p className="text-xs text-muted-foreground">Qualified Referrals</p>
                <p className="mt-1 text-lg font-semibold">
                  {dashboard?.referral_info?.active_referrals ?? 0}
                </p>
              </div>
              <div className="rounded-xl border p-3">
                <p className="text-xs text-muted-foreground">Pending Rewards</p>
                <p className="mt-1 text-lg font-semibold">{formatCurrency(pendingWithdrawal)}</p>
              </div>
              <div className="rounded-xl border p-3">
                <p className="text-xs text-muted-foreground">Paid Rewards</p>
                <p className="mt-1 text-lg font-semibold">
                  {formatCurrency(dashboard?.referral_earnings_ngn)}
                </p>
              </div>
            </div>
            <div className="rounded-xl border border-primary/20 bg-primary/5 p-3 text-sm">
              <p className="text-muted-foreground">Lifetime Referral Earnings</p>
              <p className="text-2xl font-bold">{formatCurrency(dashboard?.referral_earnings_ngn)}</p>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-base">Recent Activity</CardTitle>
            <CardDescription>Investments, rewards, referrals, and withdrawals.</CardDescription>
          </CardHeader>
          <CardContent>
            {activity.length ? (
              <div className="space-y-3">
                {activity.map((item) => (
                  <div
                    key={item.id}
                    className="flex items-center justify-between gap-3 rounded-xl border border-border/60 px-3 py-3 transition hover:border-primary/30"
                  >
                    <div>
                      <p className="font-medium">{item.title}</p>
                      <p className="text-xs text-muted-foreground">{formatDate(item.at)}</p>
                    </div>
                    <div className="text-right">
                      <p className="font-semibold tabular-nums">{item.detail}</p>
                      <Badge variant={statusVariant(item.status)} className="mt-1 capitalize">
                        {item.status.replaceAll("_", " ")}
                      </Badge>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-sm text-muted-foreground">No activity yet. Make your first investment to begin.</p>
            )}
          </CardContent>
        </Card>
      </div>

      {/* ─── Empty state ─────────────────────────────────── */}
      {investments.length === 0 ? (
        <Card className="border-dashed border-primary/20 bg-gradient-to-br from-primary/5 via-background to-cyan-500/10">
          <CardContent className="flex flex-col items-center justify-center gap-4 px-8 py-12 text-center">
            <div className="rounded-full bg-primary/10 p-5 text-primary">
              <PiggyBank className="h-8 w-8" />
            </div>
            <div className="max-w-xl space-y-2">
              <h3 className="text-2xl font-semibold">Your investment journey starts here</h3>
              <p className="text-sm text-muted-foreground">
                Start with a small, confidence-building investment and let daily rewards build momentum over time.
              </p>
            </div>
            <Button size="lg" onClick={() => openInvest(null)}>
              <ArrowUpRight className="mr-2 h-4 w-4" /> Start Investing
            </Button>
          </CardContent>
        </Card>
      ) : null}

      {/* ─── Active investments list (compact) ────────────── */}
      {investments.length > 0 ? (
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Your Investments</CardTitle>
          </CardHeader>
          <CardContent className="grid gap-3 md:grid-cols-2">
            {investments.map((investment) => {
              const pct =
                investment.progress_pct ??
                getProgressPercentage(
                  investment.days_remaining ?? investment.lock_period_days,
                  investment.lock_period_days,
                );
              return (
                <div key={investment.id} className="rounded-2xl border p-4 transition hover:border-primary/30">
                  <div className="mb-3 flex items-start justify-between gap-2">
                    <div>
                      <p className="font-semibold">{investment.plan_name}</p>
                      <p className="text-xs text-muted-foreground">
                        {investment.lock_period_days} business days
                      </p>
                    </div>
                    <Badge variant={statusVariant(String(investment.status))} className="capitalize">
                      {String(investment.status).replaceAll("_", " ")}
                    </Badge>
                  </div>
                  <div className="grid grid-cols-2 gap-2 text-sm">
                    <span>
                      Capital<strong className="block">{formatCurrency(investment.amount_paid)}</strong>
                    </span>
                    <span>
                      Earned<strong className="block">{formatCurrency(investment.roi_cdt)}</strong>
                    </span>
                  </div>
                  {investment.status === "active" ? (
                    <div className="mt-3">
                      <Progress value={pct} />
                    </div>
                  ) : null}
                </div>
              );
            })}
          </CardContent>
        </Card>
      ) : null}

      {/* ─── Collapsible settings ─────────────────────────── */}
      <Card className="border-dashed">
        <CardHeader className="cursor-pointer" onClick={() => setSettingsOpen((open) => !open)}>
          <div className="flex items-center justify-between gap-3">
            <div className="flex items-center gap-2">
              <Settings2 className="h-4 w-4" />
              <CardTitle className="text-base">Preferences</CardTitle>
            </div>
            <Badge variant="outline">{settingsOpen ? "Hide" : "Show"}</Badge>
          </div>
          <CardDescription>Notifications, preferred payment, and auto-reinvest.</CardDescription>
        </CardHeader>
        {settingsOpen ? (
          <CardContent className="grid gap-6 md:grid-cols-3">
            <div className="space-y-4">
              <p className="text-sm font-medium">Investment</p>
              <div className="flex items-center justify-between gap-3">
                <Label htmlFor="auto-reinvest">Auto-reinvest matured capital</Label>
                <Switch
                  id="auto-reinvest"
                  checked={prefs.autoReinvest}
                  onCheckedChange={(checked) => savePrefs({ ...prefs, autoReinvest: checked })}
                />
              </div>
            </div>
            <div className="space-y-4">
              <p className="text-sm font-medium">Notifications</p>
              <div className="flex items-center justify-between gap-3">
                <Label htmlFor="notify-daily">Daily rewards</Label>
                <Switch
                  id="notify-daily"
                  checked={prefs.notifyDailyRewards}
                  onCheckedChange={(checked) => savePrefs({ ...prefs, notifyDailyRewards: checked })}
                />
              </div>
              <div className="flex items-center justify-between gap-3">
                <Label htmlFor="notify-withdrawals">Withdrawals</Label>
                <Switch
                  id="notify-withdrawals"
                  checked={prefs.notifyWithdrawals}
                  onCheckedChange={(checked) => savePrefs({ ...prefs, notifyWithdrawals: checked })}
                />
              </div>
              <div className="flex items-center justify-between gap-3">
                <Label htmlFor="notify-referrals">Referrals</Label>
                <Switch
                  id="notify-referrals"
                  checked={prefs.notifyReferrals}
                  onCheckedChange={(checked) => savePrefs({ ...prefs, notifyReferrals: checked })}
                />
              </div>
            </div>
            <div className="space-y-4">
              <p className="text-sm font-medium">Preferred payment</p>
              <Select
                value={prefs.preferredPaymentMethod}
                onValueChange={(value) => {
                  const next = value as EarnPreferences["preferredPaymentMethod"];
                  setProvider(next);
                  savePrefs({ ...prefs, preferredPaymentMethod: next });
                }}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Select provider" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="paystack">Paystack</SelectItem>
                  <SelectItem value="flutterwave">Flutterwave</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </CardContent>
        ) : null}
      </Card>

      {/* ─── Trust badges ─────────────────────────────────── */}
      <div className="flex flex-wrap items-center justify-center gap-4 text-xs text-muted-foreground">
        <span className="inline-flex items-center gap-1">
          <ShieldCheck className="h-3.5 w-3.5" /> Capital Protected
        </span>
        <span className="inline-flex items-center gap-1">
          <Sparkles className="h-3.5 w-3.5" /> Daily Rewards
        </span>
        <span className="inline-flex items-center gap-1">
          <Clock3 className="h-3.5 w-3.5" /> {processingHours}hr Withdrawals
        </span>
        <span className="inline-flex items-center gap-1">
          <Wallet className="h-3.5 w-3.5" /> Secure Payments
        </span>
      </div>

      <InvestmentPaymentModal
        open={investOpen && !!selectedPlan}
        planName={selectedPlan?.name}
        exchangeRate={rate}
        minimumAmount={minUsd}
        defaultAmount={selectedPlan?.minimum_amount ?? Math.max(amountUsd, minUsd)}
        preferredProvider={provider}
        roiPercent={calc.roiPercent}
        durationDays={durationDays}
        dailyRewardNgn={dailyReward}
        isSubmitting={paying}
        onClose={() => {
          setInvestOpen(false);
          setSelectedPlan(null);
          setPaymentError(null);
        }}
        onConfirm={createPayment}
      />

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