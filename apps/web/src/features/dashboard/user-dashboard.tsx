"use client";

import * as React from "react";
import { useQuery } from "@tanstack/react-query";
import { useAuth } from "@/features/authentication/auth-provider";
import { useInvestments, useWallet } from "@/features/earn/hooks";
import * as earnApi from "@/features/earn/api";
import {
  useDashboard,
  useExchangeRate,
  useInvestmentSettings,
  usePaymentHistory,
  useRewardHistory,
  useWithdrawalHistory,
} from "@/features/investments";
import type {
  EarningsSummary,
  PaymentHistoryItem,
  RewardHistoryItem,
  WithdrawalHistoryItem,
} from "@/features/investments/types";
import { ENABLED_PLANS } from "@/features/earn/config/investment-plans";
import * as identityApi from "@/features/identity/api";
import { displayName, profileCompletion } from "@/lib/utils/format";
import type { InvestmentSummary, WalletTransaction } from "@/lib/api/types";
import {
  AccountStatusSection,
  ActiveInvestmentsSection,
  AnnouncementsSection,
  AssetsSection,
  ForYouSection,
  MarketsSection,
  OverviewHero,
  QuickActionsGrid,
  RecentTransactionsSection,
  ReferralOverviewSection,
  RewardsSection,
  TrendingEventsSection,
  type AssetWallet,
  type OverviewTransaction,
  type RewardActivityItem,
} from "@/features/dashboard/components";

function matchPlanName(amountUsd: number): string {
  return ENABLED_PLANS.find((p) => p.usdAmount === amountUsd)?.name ?? "Genesis";
}

function matchPlanRoi(amountUsd: number, fallbackRoi: number): number {
  return ENABLED_PLANS.find((p) => p.usdAmount === amountUsd)?.roiPercent ?? fallbackRoi;
}

function normalizeInvestment(
  item: InvestmentSummary | EarningsSummary,
  settingsRoi = 18,
): InvestmentSummary {
  if ("plan_name" in item) return item;
  return {
    id: item.id,
    plan_name: `${matchPlanName(item.amount_usd)} Plan`,
    amount_paid: item.amount_ngn,
    allocated_cdt: item.amount_ngn,
    roi_cdt: item.total_earned_ngn,
    roi_percent: matchPlanRoi(item.amount_usd, settingsRoi),
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

function mapWalletTxType(type: string): OverviewTransaction["type"] {
  const t = type.toLowerCase();
  if (t.includes("withdraw")) return "withdrawal";
  if (t.includes("deposit") || t.includes("credit")) return "deposit";
  if (t.includes("invest")) return "investment";
  if (t.includes("reward") || t.includes("roi")) return "reward";
  if (t.includes("refer")) return "referral";
  if (t.includes("trade")) return "trade";
  return "deposit";
}

function normalizeWalletTransactions(
  payload: { items?: WalletTransaction[] } | WalletTransaction[] | undefined,
): OverviewTransaction[] {
  const list = Array.isArray(payload) ? payload : payload?.items ?? [];
  return list.map((tx) => ({
    id: tx.id,
    type: mapWalletTxType(tx.type),
    amount: Math.abs(tx.amount),
    currency: "CDT",
    status: "completed",
    timestamp: tx.created_at,
    label: tx.description || tx.type,
  }));
}

export function UserDashboard() {
  const { user } = useAuth();

  const profileQ = useQuery({
    queryKey: ["users", "me"],
    queryFn: identityApi.getProfile,
    initialData: user ?? undefined,
  });
  const referralQ = useQuery({
    queryKey: ["referrals", "dashboard"],
    queryFn: identityApi.getReferralDashboard,
  });
  const activityQ = useQuery({
    queryKey: ["activity"],
    queryFn: identityApi.getActivityLog,
  });
  const sessionsQ = useQuery({
    queryKey: ["sessions"],
    queryFn: identityApi.getSessions,
  });

  const walletQ = useWallet();
  const legacyInvestmentsQ = useInvestments();
  const earningsDashQ = useDashboard();
  const rateQ = useExchangeRate();
  const settingsQ = useInvestmentSettings();
  const rewardsQ = useRewardHistory(1, 10);
  const paymentsQ = usePaymentHistory(1, 10);
  const withdrawalsQ = useWithdrawalHistory(1, 10);
  const walletTxQ = useQuery({
    queryKey: ["wallet", "transactions", 1, 10],
    queryFn: () => earnApi.getWalletTransactions(1, 10),
    staleTime: 30_000,
  });

  const [refreshing, setRefreshing] = React.useState(false);
  const [lastUpdated, setLastUpdated] = React.useState<Date | null>(null);

  const me = profileQ.data ?? user;
  const firstName = displayName(me).split(" ")[0] || "Investor";
  const completion = profileCompletion(me);
  const referrals = referralQ.data;
  const wallet = walletQ.data;
  const dashboard = earningsDashQ.data;
  const settings = settingsQ.data;
  const rate = rateQ.data?.usd_to_ngn ?? dashboard?.exchange_rate ?? 0;
  const settingsRoi = settings?.roi_percent ?? 18;
  const processingHours = settings?.withdrawal_processing_hours ?? 24;

  const investments: InvestmentSummary[] = React.useMemo(() => {
    if (dashboard?.investments?.length) {
      return dashboard.investments.map((item) => normalizeInvestment(item, settingsRoi));
    }
    return (legacyInvestmentsQ.data?.investments ?? []).map((item) =>
      normalizeInvestment(item, settingsRoi),
    );
  }, [dashboard?.investments, legacyInvestmentsQ.data?.investments, settingsRoi]);

  const portfolioUsd =
    dashboard?.total_invested_usd ??
    legacyInvestmentsQ.data?.total_invested ??
    wallet?.total_balance ??
    0;
  const portfolioNgn =
    dashboard?.total_invested_ngn ??
    (rate > 0 ? portfolioUsd * rate : 0);
  const todayEarnings = dashboard?.today_earnings_ngn ?? 0;
  const availableBalance =
    dashboard?.available_balance_ngn ??
    dashboard?.referral_info?.withdrawable_balance_ngn ??
    wallet?.available_balance ??
    0;
  const lockedInvestments =
    legacyInvestmentsQ.data?.locked_cdt ??
    wallet?.locked_balance ??
    dashboard?.total_invested_usd ??
    0;
  const referralEarnings =
    dashboard?.referral_earnings_ngn ??
    dashboard?.referral_info?.referral_earnings_ngn ??
    Number(referrals?.rewards_earned ?? 0);

  const heroLoading =
    earningsDashQ.isLoading && walletQ.isLoading && legacyInvestmentsQ.isLoading;

  const assetWallets: AssetWallet[] = React.useMemo(() => {
    const ngn = dashboard?.available_balance_ngn ?? 0;
    const cdt = wallet?.total_balance ?? 0;
    const usdFromInvested = dashboard?.total_invested_usd ?? 0;
    return [
      {
        currency: "USD",
        label: "USD Wallet",
        balance: usdFromInvested,
        symbol: "$",
        available: true,
      },
      {
        currency: "NGN",
        label: "NGN Wallet",
        balance: ngn,
        symbol: "₦",
        available: true,
      },
      {
        currency: "USDT",
        label: "USDT Wallet",
        balance: 0,
        symbol: "",
        available: false,
      },
      {
        currency: "BTC",
        label: "BTC Wallet",
        balance: 0,
        symbol: "",
        available: false,
      },
      {
        currency: "ETH",
        label: "ETH Wallet",
        balance: 0,
        symbol: "",
        available: false,
      },
      {
        currency: "CDT",
        label: "CDT Wallet",
        balance: cdt,
        symbol: "",
        available: true,
      },
    ];
  }, [dashboard?.available_balance_ngn, dashboard?.total_invested_usd, wallet?.total_balance]);

  const rewardActivity: RewardActivityItem[] = React.useMemo(() => {
    const list: RewardHistoryItem[] = rewardsQ.data ?? [];
    return list.map((r) => ({
      id: r.id,
      amount_ngn: r.amount_ngn,
      label: "Daily reward",
      created_at: r.created_at || r.reward_date || new Date().toISOString(),
      status: r.status,
    }));
  }, [rewardsQ.data]);

  const lifetimeRewards = React.useMemo(
    () => rewardActivity.reduce((sum, r) => sum + (r.amount_ngn || 0), 0) || todayEarnings,
    [rewardActivity, todayEarnings],
  );
  const pendingRewards = dashboard?.pending_withdrawal_ngn ?? 0;
  const claimedRewards = Math.max(0, lifetimeRewards - pendingRewards);

  const transactions: OverviewTransaction[] = React.useMemo(() => {
    const fromWallet = normalizeWalletTransactions(walletTxQ.data);
    const payments: OverviewTransaction[] = (paymentsQ.data ?? []).map((p: PaymentHistoryItem) => ({
      id: `pay-${p.id}`,
      type: "investment" as const,
      amount: p.amount_ngn ?? p.amount_usd ?? 0,
      currency: p.amount_ngn != null ? "NGN" : "USD",
      status: p.status || "pending",
      timestamp: p.created_at || new Date().toISOString(),
      label: p.provider ? `Investment · ${p.provider}` : "Investment",
    }));
    const withdrawals: OverviewTransaction[] = (withdrawalsQ.data ?? []).map(
      (w: WithdrawalHistoryItem) => ({
        id: `wd-${w.id}`,
        type: "withdrawal" as const,
        amount: w.amount_ngn ?? 0,
        currency: "NGN",
        status: w.status || "pending",
        timestamp: w.created_at || new Date().toISOString(),
        label: "Withdrawal",
      }),
    );
    const rewards: OverviewTransaction[] = rewardActivity.map((r) => ({
      id: `rw-${r.id}`,
      type: "reward" as const,
      amount: r.amount_ngn,
      currency: "NGN",
      status: r.status || "completed",
      timestamp: r.created_at,
      label: r.label,
    }));

    return [...fromWallet, ...payments, ...withdrawals, ...rewards]
      .sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime())
      .slice(0, 8);
  }, [walletTxQ.data, paymentsQ.data, withdrawalsQ.data, rewardActivity]);

  const announcements = React.useMemo(() => {
    const activity = activityQ.data ?? [];
    if (!activity.length) return undefined;
    return activity.slice(0, 5).map((a) => ({
      id: a.id,
      title: a.action.replace(/[._]/g, " ").replace(/\b\w/g, (c) => c.toUpperCase()),
      category: a.action.includes("invest")
        ? "Investment"
        : a.action.includes("security") || a.action.includes("login")
          ? "System"
          : "Platform",
      created_at: a.created_at,
      href: "/app/notifications",
    }));
  }, [activityQ.data]);

  const securityScore = React.useMemo(() => {
    let score = 35;
    if (me?.is_verified) score += 25;
    if (completion.percent >= 80) score += 20;
    else if (completion.percent >= 40) score += 10;
    if ((sessionsQ.data ?? []).some((s) => s.is_current)) score += 10;
    if (me?.is_genesis) score += 10;
    return Math.min(100, score);
  }, [me?.is_verified, me?.is_genesis, completion.percent, sessionsQ.data]);

  const handleRefresh = React.useCallback(() => {
    setRefreshing(true);
    void Promise.all([
      profileQ.refetch(),
      referralQ.refetch(),
      walletQ.refetch(),
      legacyInvestmentsQ.refetch(),
      earningsDashQ.refetch(),
      rateQ.refetch(),
      settingsQ.refetch(),
      rewardsQ.refetch(),
      paymentsQ.refetch(),
      withdrawalsQ.refetch(),
      walletTxQ.refetch(),
      activityQ.refetch(),
    ]).finally(() => {
      setLastUpdated(new Date());
      window.setTimeout(() => setRefreshing(false), 400);
    });
  }, [
    profileQ,
    referralQ,
    walletQ,
    legacyInvestmentsQ,
    earningsDashQ,
    rateQ,
    settingsQ,
    rewardsQ,
    paymentsQ,
    withdrawalsQ,
    walletTxQ,
    activityQ,
  ]);

  React.useEffect(() => {
    if (!lastUpdated && !heroLoading) {
      setLastUpdated(new Date());
    }
  }, [heroLoading, lastUpdated]);

  return (
    <div className="mx-auto max-w-7xl space-y-7 pb-10">
      {/* 1. Hero */}
      <OverviewHero
        firstName={firstName}
        portfolioUsd={portfolioUsd}
        portfolioNgn={portfolioNgn}
        todayEarningsNgn={todayEarnings}
        availableBalanceNgn={availableBalance}
        lockedInvestments={lockedInvestments}
        referralEarningsNgn={referralEarnings}
        exchangeRate={rate}
        loading={heroLoading}
        lastUpdated={lastUpdated}
        onRefresh={handleRefresh}
        refreshing={refreshing}
      />

      {/* 2. Assets */}
      <AssetsSection wallets={assetWallets} loading={walletQ.isLoading && earningsDashQ.isLoading} />

      {/* 3. Quick Actions */}
      <QuickActionsGrid />

      {/* 4. Active Investments */}
      <ActiveInvestmentsSection
        investments={investments}
        loading={earningsDashQ.isLoading && legacyInvestmentsQ.isLoading}
        processingHours={processingHours}
        lastWithdrawalAt={dashboard?.last_withdrawal_at}
        todayEarningsNgn={todayEarnings}
      />

      {/* 5–6. For You + Markets */}
      <div className="grid gap-7 xl:grid-cols-[1.35fr_1fr]">
        <ForYouSection />
        <MarketsSection />
      </div>

      {/* 7. Trending Events */}
      <TrendingEventsSection />

      {/* 8–9. Announcements + Referrals */}
      <div className="grid gap-7 lg:grid-cols-2">
        <AnnouncementsSection items={announcements} loading={activityQ.isLoading} />
        <ReferralOverviewSection
          loading={referralQ.isLoading}
          data={{
            referralCode: referrals?.referral_code || me?.referral_code,
            referralLink: referrals?.referral_link,
            todayEarningsNgn: referralEarnings > 0 ? Math.min(referralEarnings, todayEarnings || referralEarnings) : 0,
            weeklyEarningsNgn: referralEarnings,
            monthlyEarningsNgn: referralEarnings,
            totalReferrals: referrals?.successful_invites ?? referrals?.total_invites ?? 0,
            rank: referrals?.leaderboard_rank,
            rewardsEarned: Number(referrals?.rewards_earned ?? referralEarnings),
          }}
        />
      </div>

      {/* 10–11. Rewards + Transactions */}
      <div className="grid gap-7 lg:grid-cols-2">
        <RewardsSection
          loading={rewardsQ.isLoading && earningsDashQ.isLoading}
          todayReward={todayEarnings}
          pendingRewards={pendingRewards}
          claimedRewards={claimedRewards}
          lifetimeRewards={lifetimeRewards || referralEarnings}
          activity={rewardActivity}
        />
        <RecentTransactionsSection
          loading={walletTxQ.isLoading && paymentsQ.isLoading && withdrawalsQ.isLoading}
          items={transactions}
        />
      </div>

      {/* 12. Account Status */}
      <AccountStatusSection
        data={{
          verified: me?.is_verified,
          twoFactorEnabled: false,
          kycLevel: me?.is_verified ? "Level 1" : "Level 0",
          membership: me?.is_genesis
            ? `Genesis #${me.genesis_number ?? "—"}`
            : me?.is_founder
              ? "Founder"
              : "Standard",
          referralTier:
            (referrals?.successful_invites ?? 0) >= 10
              ? "Ambassador"
              : (referrals?.successful_invites ?? 0) >= 3
                ? "Builder"
                : "Starter",
          securityScore,
          profileCompletion: completion.percent,
          missingProfile: completion.missing,
        }}
      />
    </div>
  );
}
