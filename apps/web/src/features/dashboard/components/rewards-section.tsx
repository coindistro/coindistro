"use client";

import Link from "next/link";
import { Gift } from "lucide-react";
import { Button } from "@coindistro/cds";
import { formatCurrency } from "@/features/earn/utils";
import { formatRelative } from "@/lib/utils/format";
import { SectionShell, GlassCard, SectionSkeleton } from "./section-shell";

export interface RewardActivityItem {
  id: string;
  amount_ngn: number;
  label: string;
  created_at: string;
  status?: string;
}

export function RewardsSection({
  todayReward = 0,
  pendingRewards = 0,
  claimedRewards = 0,
  lifetimeRewards = 0,
  activity = [],
  loading,
}: {
  todayReward?: number;
  pendingRewards?: number;
  claimedRewards?: number;
  lifetimeRewards?: number;
  activity?: RewardActivityItem[];
  loading?: boolean;
}) {
  return (
    <SectionShell
      id="rewards"
      title="Rewards"
      description="Daily earnings and reward activity"
      actionHref="/app/earn"
      actionLabel="Open Earn"
    >
      {loading ? (
        <SectionSkeleton rows={2} />
      ) : (
        <div className="grid gap-3 lg:grid-cols-[1.1fr_1fr]">
          <div className="grid grid-cols-2 gap-3">
            <RewardStat label="Today's Reward" value={formatCurrency(todayReward)} />
            <RewardStat label="Pending Rewards" value={formatCurrency(pendingRewards)} />
            <RewardStat label="Claimed Rewards" value={formatCurrency(claimedRewards)} />
            <RewardStat label="Lifetime Rewards" value={formatCurrency(lifetimeRewards)} />
          </div>

          <GlassCard className="min-h-[180px]">
            <p className="text-sm font-semibold text-foreground">Recent Reward Activity</p>
            {activity.length === 0 ? (
              <div className="mt-6 flex flex-col items-center text-center">
                <div className="rounded-2xl bg-primary/10 p-3 text-primary">
                  <Gift className="h-5 w-5" aria-hidden />
                </div>
                <p className="mt-3 text-sm font-medium text-foreground">No rewards yet</p>
                <p className="mt-1 text-xs text-muted-foreground">
                  Active investments credit rewards every business day.
                </p>
                <Button className="mt-4" size="sm" variant="outline" asChild>
                  <Link href="/app/earn">Start investing</Link>
                </Button>
              </div>
            ) : (
              <ul className="mt-3 space-y-2.5">
                {activity.slice(0, 5).map((item) => (
                  <li
                    key={item.id}
                    className="flex items-center justify-between gap-2 rounded-xl border border-border/40 bg-background/30 px-3 py-2 text-sm"
                  >
                    <div className="min-w-0">
                      <p className="truncate font-medium text-foreground">{item.label}</p>
                      <p className="text-[11px] text-muted-foreground">
                        {formatRelative(item.created_at)}
                        {item.status ? ` · ${item.status}` : ""}
                      </p>
                    </div>
                    <p className="shrink-0 font-semibold tabular-nums text-emerald-500">
                      +{formatCurrency(item.amount_ngn)}
                    </p>
                  </li>
                ))}
              </ul>
            )}
          </GlassCard>
        </div>
      )}
    </SectionShell>
  );
}

function RewardStat({ label, value }: { label: string; value: string }) {
  return (
    <GlassCard>
      <p className="text-xs text-muted-foreground">{label}</p>
      <p className="mt-2 text-lg font-bold tabular-nums text-foreground">{value}</p>
    </GlassCard>
  );
}
