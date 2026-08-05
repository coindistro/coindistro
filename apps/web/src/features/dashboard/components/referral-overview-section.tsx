"use client";

import * as React from "react";
import Link from "next/link";
import { Copy, Share2 } from "lucide-react";
import { Button } from "@coindistro/cds";
import { formatCurrency } from "@/features/earn/utils";
import { useToast } from "@/features/shared/providers/toast-provider";
import { SectionShell, GlassCard, SectionSkeleton } from "./section-shell";

export interface ReferralOverviewData {
  referralCode?: string | null;
  referralLink?: string | null;
  todayEarningsNgn?: number;
  weeklyEarningsNgn?: number;
  monthlyEarningsNgn?: number;
  totalReferrals?: number;
  rank?: number | null;
  rewardsEarned?: number;
}

export function ReferralOverviewSection({
  data,
  loading,
}: {
  data?: ReferralOverviewData | null;
  loading?: boolean;
}) {
  const { toast } = useToast();
  const code = data?.referralCode || "—";

  const copyCode = async () => {
    if (!data?.referralCode) return;
    try {
      await navigator.clipboard.writeText(data.referralCode);
      toast({ message: "Referral code copied", variant: "success" });
    } catch {
      toast({ message: "Could not copy code", variant: "danger" });
    }
  };

  const shareLink = async () => {
    const link = data?.referralLink || data?.referralCode;
    if (!link) return;
    try {
      if (navigator.share) {
        await navigator.share({ title: "Join CoinDistro", url: link, text: "Invest with me on CoinDistro" });
      } else {
        await navigator.clipboard.writeText(link);
        toast({ message: "Referral link copied", variant: "success" });
      }
    } catch {
      /* user cancelled share */
    }
  };

  return (
    <SectionShell
      id="referrals"
      title="Referral Overview"
      description="Grow with your network"
      actionHref="/app/referrals"
      actionLabel="Open Referrals"
    >
      {loading ? (
        <SectionSkeleton rows={2} />
      ) : !data?.referralCode && !data?.totalReferrals ? (
        <GlassCard className="border-dashed p-6 text-center">
          <p className="font-semibold text-foreground">No referrals yet</p>
          <p className="mt-1 text-sm text-muted-foreground">
            Share your code and earn when friends invest.
          </p>
          <Button className="mt-4" asChild>
            <Link href="/app/referrals">Invite friends</Link>
          </Button>
        </GlassCard>
      ) : (
        <GlassCard>
          <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <p className="text-xs text-muted-foreground">Referral Code</p>
              <p className="mt-1 font-mono text-2xl font-bold tracking-wide text-foreground">
                {code}
              </p>
            </div>
            <div className="flex flex-wrap gap-2">
              <Button type="button" variant="outline" size="sm" onClick={() => void copyCode()}>
                <Copy className="mr-2 h-4 w-4" aria-hidden />
                Copy
              </Button>
              <Button type="button" size="sm" onClick={() => void shareLink()}>
                <Share2 className="mr-2 h-4 w-4" aria-hidden />
                Share
              </Button>
            </div>
          </div>

          <div className="mt-5 grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-6">
            <Stat label="Today" value={formatCurrency(data?.todayEarningsNgn ?? 0)} />
            <Stat label="Weekly" value={formatCurrency(data?.weeklyEarningsNgn ?? 0)} />
            <Stat label="Monthly" value={formatCurrency(data?.monthlyEarningsNgn ?? data?.rewardsEarned ?? 0)} />
            <Stat label="Total Referrals" value={String(data?.totalReferrals ?? 0)} />
            <Stat label="Referral Rank" value={data?.rank != null ? `#${data.rank}` : "—"} />
            <Stat label="Rewards" value={formatCurrency(data?.rewardsEarned ?? 0)} />
          </div>
        </GlassCard>
      )}
    </SectionShell>
  );
}

function Stat({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-xl border border-border/50 bg-background/40 px-3 py-2.5">
      <p className="text-[11px] text-muted-foreground">{label}</p>
      <p className="mt-1 text-sm font-semibold tabular-nums text-foreground">{value}</p>
    </div>
  );
}
