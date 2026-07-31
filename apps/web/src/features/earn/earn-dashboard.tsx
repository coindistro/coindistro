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
  PageHeader,
  Progress,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
  StatCard,
  Switch,
} from "@coindistro/cds";
import {
  ArrowRight,
  CheckCircle2,
  Clock3,
  Copy,
  Gift,
  PiggyBank,
  Settings2,
  Sparkles,
  Wallet,
} from "lucide-react";
import { useInvestments, useInvestmentPlans, useWallet } from "@/features/earn/hooks";
import { InvestmentPaymentModal } from "@/features/earn/investment-payment-modal";
import { WithdrawalRequestModal } from "@/features/earn/withdrawal-request-modal";
import {
  calculateInvestment,
  formatCurrency,
  getProgressPercentage,
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
  if (status === "pending" || status === "pending_payment" || status === "pending_review" || status === "processing") {
    return "warning";
  }
  if (status === "failed" || status === "cancelled" || status === "rejected") {
    return "danger";
  }
  return "outline";
}

function normalizeInvestment(item: InvestmentSummary | EarningsSummary): InvestmentSummary {
  if ("plan_name" in item) return item;
  return {
    id: item.id,
    plan_name: "Investment plan",
    amount_paid: item.amount_ngn,
    allocated_cdt: item.amount_ngn,
    roi_cdt: item.total_earned_ngn,
    roi_percent: 0,
    status: item.status,
    lock_period_days: item.max_business_days,
    days_remaining: item.remaining_days,
    progress_pct: item.progress_pct,
    started_at: item.started_at,
    matures_at: item.maturity_date,
    created_at: item.created_at,
  };
}

function InvestmentCard({
  investment,
  dailyRewardNgn,
}: {
  investment: InvestmentSummary;
  dailyRewardNgn: number;
}) {
  const progress =
    investment.progress_pct ??
    getProgressPercentage(
      investment.days_remaining ?? investment.lock_period_days,
      investment.lock_period_days,
    );
  const nextPayout = investment.status === "active" ? dailyRewardNgn : 0;

  return (
    <div className="rounded-lg border p-4">
      <div className="mb-3 flex items-start justify-between gap-2">
        <div>
          <p className="font-semibold">{investment.plan_name || "Investment plan"}</p>
          <p className="text-xs text-muted-foreground">
            {investment.lock_period_days} business days
          </p>
        </div>
        <Badge variant={statusVariant(String(investment.status))} className="capitalize">
          {String(investment.status).replaceAll("_", " ")}
        </Badge>
      </div>
      <div className="grid grid-cols-2 gap-3 text-sm">
        <div>
          <p className="text-xs text-muted-foreground">Invested</p>
          <p className="font-medium">{formatCurrency(investment.amount_paid)}</p>
        </div>
        <div>
          <p className="text-xs text-muted-foreground">Total earnings</p>
          <p className="font-medium">{formatCurrency(investment.roi_cdt)}</p>
        </div>
        <div>
          <p className="text-xs text-muted-foreground">Next payout</p>
          <p className="font-medium">{formatCurrency(nextPayout)}</p>
        </div>
        <div>
          <p className="text-xs text-muted-foreground">Maturity date</p>
          <p className="font-medium">{formatDate(investment.matures_at)}</p>
        </div>
        <div>
          <p className="text-xs text-muted-foreground">Days remaining</p>
          <p className="font-medium">{investment.days_remaining ?? "—"}</p>
        </div>
        <div>
          <p className="text-xs text-muted-foreground">Investment status</p>
          <p className="font-medium capitalize">{String(investment.status).replaceAll("_", " ")}</p>
        </div>
      </div>
      {investment.status === "active" && (
        <div className="mt-3">
          <div className="mb-1 flex justify-between text-xs text-muted-foreground">
            <span>Investment progress</span>
            <span>{Math.round(progress)}%</span>
          </div>
          <Progress value={progress} />
        </div>
      )}
    </div>
  );
}

function HistoryCard({
  title,
  rows,
  amount,
}: {
  title: string;
  rows: Array<{
    id: string;
    status: string;
    created_at: string;
    amount_ngn?: number;
    amount_usd?: number;
    fee_ngn?: number;
    net_amount_ngn?: number;
  }>;
  amount: (row: {
    amount_ngn?: number;
    amount_usd?: number;
    fee_ngn?: number;
    net_amount_ngn?: number;
  }) => string;
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">{title}</CardTitle>
      </CardHeader>
      <CardContent>
        {rows.length ? (
          <div className="space-y-3">
            {rows.slice(0, 8).map((row) => (
              <div
                key={row.id}
                className="flex items-center justify-between gap-3 border-b pb-3 text-sm last:border-0 last:pb-0"
              >
                <div>
                  <p className="font-medium">{amount(row)}</p>
                  <p className="text-xs text-muted-foreground">{formatDate(row.created_at)}</p>
                </div>
                <Badge variant={statusVariant(row.status)} className="capitalize">
                  {row.status.replaceAll("_", " ")}
                </Badge>
              </div>
            ))}
          </div>
        ) : (
          <p className="text-sm text-muted-foreground">No records yet.</p>
        )}
      </CardContent>
    </Card>
  );
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

export function EarnDashboard() {
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

  const [selectedPlan, setSelectedPlan] = React.useState<InvestmentPlan | null>(null);
  const [paying, setPaying] = React.useState(false);
  const [withdrawOpen, setWithdrawOpen] = React.useState(false);
  const [withdrawing, setWithdrawing] = React.useState(false);
  const [copied, setCopied] = React.useState(false);
  const [prefs, setPrefs] = React.useState<EarnPreferences>(defaultPrefs);
  const [paymentError, setPaymentError] = React.useState<string | null>(null);

  React.useEffect(() => {
    setPrefs(loadPreferences());
  }, []);

  const dashboard = dashboardQ.data;
  const settings = settingsQ.data;
  const rate =
    rateQ.data?.usd_to_ngn ?? dashboard?.exchange_rate ?? 1600;
  const plans = Array.isArray(plansQ.data) ? plansQ.data : [];
  const investments: InvestmentSummary[] =
    dashboard?.investments?.map(normalizeInvestment) ??
    (legacyInvestmentsQ.data?.investments ?? []).map(normalizeInvestment);

  const minUsd = settings?.minimum_investment_usd ?? 30;
  const durationDays = settings?.max_business_days ?? 20;
  const dailyReward = settings?.daily_reward_ngn ?? 0;
  const roiPercent = settings?.roi_percent ?? selectedPlan?.roi_percent ?? 0;
  const amount = selectedPlan?.minimum_amount ?? minUsd;
  const calc = calculateInvestment({
    amountUsd: amount,
    exchangeRate: rate,
    dailyRewardNgn: dailyReward,
    durationBusinessDays: durationDays,
    roiPercent: selectedPlan?.roi_percent ?? roiPercent,
  });

  const available =
    dashboard?.available_balance_ngn ??
    dashboard?.referral_info?.withdrawable_balance_ngn ??
    walletQ.data?.available_balance ??
    0;
  const locked =
    dashboard?.total_invested_ngn ?? walletQ.data?.locked_balance ?? 0;
  const withdrawalBalance =
    dashboard?.referral_info?.withdrawable_balance_ngn ?? available;
  const early = investments.some((item) => item.status === "active");
  const activeInvestment = investments.find((item) => item.status === "active");
  const daysRemaining = activeInvestment?.days_remaining;
  const nextMaturity = activeInvestment?.matures_at;

  const savePrefs = (next: EarnPreferences) => {
    setPrefs(next);
    if (typeof window !== "undefined") {
      window.localStorage.setItem(PREFS_KEY, JSON.stringify(next));
    }
  };

  const createPayment = async (provider: "paystack" | "flutterwave", investAmount: number) => {
    setPaying(true);
    setPaymentError(null);
    try {
      const result =
        provider === "paystack"
          ? await investmentApi.initPaystackPayment(investAmount)
          : await investmentApi.initFlutterwavePayment(investAmount);
      if (!result.authorization_url) {
        throw new Error("Payment provider did not return a checkout URL");
      }
      window.location.assign(result.authorization_url);
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
      // Earnings balance withdrawal uses amount_ngn; investment_id is reserved for principal early/normal exits.
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

  const loading = dashboardQ.isLoading || rateQ.isLoading || settingsQ.isLoading;
  const totalEarnings =
    (dashboard?.monthly_earnings_ngn ?? 0) > 0
      ? (dashboard?.monthly_earnings_ngn ?? 0) * 12
      : investments.reduce((sum, item) => sum + (item.roi_cdt || 0), 0);

  const stats: Array<[string, string, React.ComponentType<{ className?: string }>]> = [
    ["Portfolio value", formatCurrency(dashboard?.total_invested_ngn), Wallet],
    ["Total invested", `$${dashboard?.total_invested_usd ?? 0}`, PiggyBank],
    ["Today's earnings", formatCurrency(dashboard?.today_earnings_ngn), Sparkles],
    ["Pending earnings", formatCurrency(dashboard?.pending_withdrawal_ngn), Clock3],
    ["Total earnings", formatCurrency(totalEarnings), Gift],
    ["Referral earnings", formatCurrency(dashboard?.referral_earnings_ngn), Gift],
    ["Available balance", formatCurrency(available), Wallet],
    ["Locked balance", formatCurrency(locked), PiggyBank],
    ["Withdrawal balance", formatCurrency(withdrawalBalance), Wallet],
    ["Days remaining", daysRemaining != null ? String(daysRemaining) : "—", Clock3],
    ["Next payout", formatCurrency(activeInvestment ? dailyReward : 0), Sparkles],
    ["Maturity date", formatDate(nextMaturity), Clock3],
  ];

  return (
    <div className="space-y-6 animate-cds-fade-in">
      <PageHeader
        title="Earn"
        description="Invest, track daily rewards, and manage your portfolio."
        actions={
          <Button asChild size="sm">
            <Link href="/app/dashboard">
              <ArrowRight className="mr-2 h-4 w-4" /> Back to dashboard
            </Link>
          </Button>
        }
      />

      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        {stats.map(([title, value, Icon]) => (
          <StatCard
            key={title}
            title={title}
            value={loading ? "…" : value}
            description="Updated from your account"
            icon={<Icon className="h-4 w-4" />}
          />
        ))}
      </div>

      <div className="grid gap-6 xl:grid-cols-[1.4fr_1fr]">
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Investment plans</CardTitle>
            <CardDescription>
              Minimum investment is ${minUsd}. Values use the current exchange rate: 1 USD ={" "}
              {formatCurrency(rate)}.
            </CardDescription>
          </CardHeader>
          <CardContent className="grid gap-3 md:grid-cols-2">
            {plans.length ? (
              plans
                .filter((plan) => plan.enabled)
                .map((plan) => {
                  const projection = calculateInvestment({
                    amountUsd: Math.max(plan.minimum_amount, minUsd),
                    exchangeRate: rate,
                    dailyRewardNgn: dailyReward,
                    durationBusinessDays: durationDays,
                    roiPercent: plan.roi_percent,
                  });
                  return (
                    <div key={plan.id} className="rounded-lg border p-4">
                      <div className="flex items-start justify-between gap-2">
                        <div>
                          <p className="font-semibold">{plan.name}</p>
                          <p className="mt-1 text-xs text-muted-foreground">
                            {plan.description || "Flexible growth plan"}
                          </p>
                        </div>
                        <Badge variant="success">Live</Badge>
                      </div>
                      <div className="mt-4 grid grid-cols-2 gap-3 text-sm">
                        <span>
                          USD
                          <strong className="block">${plan.minimum_amount.toLocaleString()}</strong>
                        </span>
                        <span>
                          NGN equivalent
                          <strong className="block">
                            {formatCurrency(plan.minimum_amount * rate)}
                          </strong>
                        </span>
                        <span>
                          ROI
                          <strong className="block">{plan.roi_percent}%</strong>
                        </span>
                        <span>
                          Duration
                          <strong className="block">{durationDays} days</strong>
                        </span>
                        <span>
                          Daily payout
                          <strong className="block">{formatCurrency(dailyReward)}</strong>
                        </span>
                        <span>
                          Total payout
                          <strong className="block">
                            {formatCurrency(projection.totalPayoutNgn)}
                          </strong>
                        </span>
                      </div>
                      <Button className="mt-4 w-full" onClick={() => setSelectedPlan(plan)}>
                        Invest now
                      </Button>
                    </div>
                  );
                })
            ) : (
              <div className="rounded-lg border p-4 md:col-span-2">
                <p className="font-semibold">Default investment plan</p>
                <p className="mt-1 text-xs text-muted-foreground">
                  Configured minimum of ${minUsd} with {durationDays} business days.
                </p>
                <div className="mt-4 grid grid-cols-2 gap-3 text-sm">
                  <span>
                    USD<strong className="block">${minUsd}</strong>
                  </span>
                  <span>
                    NGN equivalent
                    <strong className="block">{formatCurrency(minUsd * rate)}</strong>
                  </span>
                  <span>
                    ROI<strong className="block">{roiPercent}%</strong>
                  </span>
                  <span>
                    Daily payout
                    <strong className="block">{formatCurrency(dailyReward)}</strong>
                  </span>
                  <span>
                    Monthly earnings
                    <strong className="block">{formatCurrency(calc.monthlyEarningsNgn)}</strong>
                  </span>
                  <span>
                    Total payout
                    <strong className="block">{formatCurrency(calc.totalPayoutNgn)}</strong>
                  </span>
                </div>
                <Button
                  className="mt-4 w-full"
                  onClick={() =>
                    setSelectedPlan({
                      id: "default",
                      name: "Default plan",
                      minimum_amount: minUsd,
                      maximum_amount: minUsd * 100,
                      currency: "USD",
                      roi_percent: roiPercent,
                      enabled: true,
                    })
                  }
                >
                  Invest now
                </Button>
              </div>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-base">Investment summary</CardTitle>
            <CardDescription>Projected using configured rates — never hardcoded in the UI.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3 text-sm">
            {[
              ["Daily earnings", formatCurrency(calc.dailyEarningsNgn)],
              ["Monthly earnings", formatCurrency(calc.monthlyEarningsNgn)],
              ["ROI", `${calc.roiPercent}%`],
              ["Total expected payout", formatCurrency(calc.totalPayoutNgn)],
              ["Maturity window", `${durationDays} business days`],
            ].map(([label, value]) => (
              <div key={label} className="flex justify-between border-b pb-2 last:border-0">
                <span className="text-muted-foreground">{label}</span>
                <strong>{value}</strong>
              </div>
            ))}
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader className="flex flex-col gap-3 space-y-0 sm:flex-row sm:items-start sm:justify-between">
          <div>
            <CardTitle className="text-base">Your investments</CardTitle>
            <CardDescription>
              Progress, next payout, maturity date, days remaining, and status.
            </CardDescription>
          </div>
          <Button variant="outline" onClick={() => setWithdrawOpen(true)}>
            Request withdrawal
          </Button>
        </CardHeader>
        <CardContent>
          {investments.length ? (
            <div className="grid gap-3 md:grid-cols-2">
              {investments.map((item) => (
                <InvestmentCard
                  key={item.id}
                  investment={item}
                  dailyRewardNgn={dailyReward}
                />
              ))}
            </div>
          ) : (
            <p className="text-sm text-muted-foreground">
              No investments yet. Choose a plan above to get started.
            </p>
          )}
        </CardContent>
      </Card>

      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Referral rewards</CardTitle>
            <CardDescription>Share your link and track referral earnings.</CardDescription>
          </CardHeader>
          <CardContent>
            {dashboard?.referral_info ? (
              <>
                <div className="grid grid-cols-2 gap-3 text-sm">
                  <span>
                    Referral code
                    <strong className="block">{dashboard.referral_info.referral_code}</strong>
                  </span>
                  <span>
                    Total referrals
                    <strong className="block">{dashboard.referral_info.total_referrals}</strong>
                  </span>
                  <span>
                    Active referrals
                    <strong className="block">{dashboard.referral_info.active_referrals}</strong>
                  </span>
                  <span>
                    Referral earnings
                    <strong className="block">
                      {formatCurrency(dashboard.referral_info.referral_earnings_ngn)}
                    </strong>
                  </span>
                </div>
                <Button variant="outline" className="mt-4" onClick={() => void copyReferral()}>
                  <Copy className="mr-2 h-4 w-4" />
                  {copied ? "Copied" : "Copy referral link"}
                </Button>
              </>
            ) : (
              <p className="text-sm text-muted-foreground">
                Referral information is not available yet.
              </p>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-base">Daily tasks</CardTitle>
            <CardDescription>Reward activities coming soon.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3 text-sm">
            <div className="flex justify-between">
              <span>Today&apos;s reward</span>
              <strong>{formatCurrency(0)}</strong>
            </div>
            <div className="flex items-center gap-2">
              <CheckCircle2 className="h-4 w-4 text-muted-foreground" />
              Completed: 0
            </div>
            <div className="flex items-center gap-2">
              <Clock3 className="h-4 w-4 text-muted-foreground" />
              Pending: Watch YouTube
            </div>
            <p className="text-xs text-muted-foreground">
              Future tasks will appear here when enabled. Backend logic is not implemented yet.
            </p>
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-6 lg:grid-cols-2 xl:grid-cols-4">
        <HistoryCard
          title="Investment history"
          rows={(paymentsQ.data ?? []).map((row) => ({ ...row }))}
          amount={(row) => `$${row.amount_usd ?? 0} · ${formatCurrency(row.amount_ngn)}`}
        />
        <HistoryCard
          title="Withdrawal history"
          rows={(withdrawalsQ.data ?? []).map((row) => ({ ...row }))}
          amount={(row) =>
            `${formatCurrency(row.net_amount_ngn)} · fee ${formatCurrency(row.fee_ngn)}`
          }
        />
        <HistoryCard
          title="Daily rewards"
          rows={(rewardsQ.data ?? []).map((row) => ({ ...row }))}
          amount={(row) => formatCurrency(row.amount_ngn)}
        />
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Referral rewards history</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3 text-sm">
            <div className="flex justify-between">
              <span className="text-muted-foreground">Lifetime earnings</span>
              <strong>{formatCurrency(dashboard?.referral_earnings_ngn)}</strong>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground">Active referrals</span>
              <strong>{dashboard?.referral_info?.active_referrals ?? 0}</strong>
            </div>
            <Badge variant="secondary">Synced from referral wallet</Badge>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Settings2 className="h-4 w-4" />
            <CardTitle className="text-base">Settings</CardTitle>
          </div>
          <CardDescription>
            Investment preferences, notification preferences, and preferred payment method.
          </CardDescription>
        </CardHeader>
        <CardContent className="grid gap-6 md:grid-cols-3">
          <div className="space-y-4">
            <p className="text-sm font-medium">Investment preferences</p>
            <div className="flex items-center justify-between gap-3">
              <Label htmlFor="auto-reinvest">Auto-reinvest matured capital</Label>
              <Switch
                id="auto-reinvest"
                checked={prefs.autoReinvest}
                onCheckedChange={(checked) => savePrefs({ ...prefs, autoReinvest: checked })}
              />
            </div>
            <div className="space-y-2">
              <Label>Configured minimum</Label>
              <Input value={`$${minUsd}`} readOnly />
            </div>
          </div>

          <div className="space-y-4">
            <p className="text-sm font-medium">Notification preferences</p>
            <div className="flex items-center justify-between gap-3">
              <Label htmlFor="notify-daily">Daily rewards</Label>
              <Switch
                id="notify-daily"
                checked={prefs.notifyDailyRewards}
                onCheckedChange={(checked) =>
                  savePrefs({ ...prefs, notifyDailyRewards: checked })
                }
              />
            </div>
            <div className="flex items-center justify-between gap-3">
              <Label htmlFor="notify-withdrawals">Withdrawals</Label>
              <Switch
                id="notify-withdrawals"
                checked={prefs.notifyWithdrawals}
                onCheckedChange={(checked) =>
                  savePrefs({ ...prefs, notifyWithdrawals: checked })
                }
              />
            </div>
            <div className="flex items-center justify-between gap-3">
              <Label htmlFor="notify-referrals">Referrals</Label>
              <Switch
                id="notify-referrals"
                checked={prefs.notifyReferrals}
                onCheckedChange={(checked) =>
                  savePrefs({ ...prefs, notifyReferrals: checked })
                }
              />
            </div>
          </div>

          <div className="space-y-4">
            <p className="text-sm font-medium">Preferred payment method</p>
            <Select
              value={prefs.preferredPaymentMethod}
              onValueChange={(value) =>
                savePrefs({
                  ...prefs,
                  preferredPaymentMethod: value as EarnPreferences["preferredPaymentMethod"],
                })
              }
            >
              <SelectTrigger>
                <SelectValue placeholder="Select provider" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="paystack">Paystack</SelectItem>
                <SelectItem value="flutterwave">Flutterwave</SelectItem>
              </SelectContent>
            </Select>
            <p className="text-xs text-muted-foreground">
              Used as the default provider when you open the payment modal.
            </p>
          </div>
        </CardContent>
      </Card>

      <InvestmentPaymentModal
        open={!!selectedPlan}
        planName={selectedPlan?.name}
        exchangeRate={rate}
        minimumAmount={minUsd}
        defaultAmount={selectedPlan?.minimum_amount ?? minUsd}
        preferredProvider={prefs.preferredPaymentMethod}
        roiPercent={selectedPlan?.roi_percent ?? roiPercent}
        durationDays={durationDays}
        dailyRewardNgn={dailyReward}
        isSubmitting={paying}
        onClose={() => {
          setSelectedPlan(null);
          setPaymentError(null);
        }}
        onConfirm={createPayment}
      />
      {paymentError && selectedPlan && (
        <p className="text-sm text-destructive" role="alert">
          {paymentError}
        </p>
      )}

      <WithdrawalRequestModal
        open={withdrawOpen}
        availableBalance={available}
        processingHours={settings?.withdrawal_processing_hours ?? 24}
        feePercent={settings?.early_withdrawal_fee_percent ?? 0}
        penaltyPercent={settings?.early_withdrawal_penalty_percent ?? 0}
        earlyWithdrawal={early}
        isSubmitting={withdrawing}
        onClose={() => setWithdrawOpen(false)}
        onConfirm={submitWithdrawal}
      />
    </div>
  );
}
