"use client";

import * as React from "react";
import Link from "next/link";
import { motion, useReducedMotion } from "framer-motion";
import { PiggyBank } from "lucide-react";
import { Badge, Button, Progress } from "@coindistro/cds";
import { formatCurrency, formatWithdrawalNextAvailable, getWithdrawalCooldown } from "@/features/earn/utils";
import type { InvestmentSummary } from "@/lib/api/types";
import { SectionShell, GlassCard, SectionSkeleton } from "./section-shell";

export function ActiveInvestmentsSection({
  investments,
  loading,
  processingHours = 24,
  lastWithdrawalAt,
  todayEarningsNgn = 0,
}: {
  investments: InvestmentSummary[];
  loading?: boolean;
  processingHours?: number;
  lastWithdrawalAt?: string | null;
  todayEarningsNgn?: number;
}) {
  const reduceMotion = useReducedMotion();
  const cooldown = React.useMemo(
    () => getWithdrawalCooldown(lastWithdrawalAt),
    [lastWithdrawalAt],
  );

  const active = investments.filter((i) => i.status === "active" || i.status === "pending" || i.status === "pending_payment");
  const list = active.length ? active : investments.slice(0, 4);

  return (
    <SectionShell
      id="active-investments"
      title="Active Investments"
      description="Genesis plans and live reward progress"
      actionHref="/app/earn"
      actionLabel="Open Earn"
    >
      {loading ? (
        <SectionSkeleton rows={2} />
      ) : list.length === 0 ? (
        <GlassCard className="flex flex-col items-start gap-3 border-dashed bg-muted/20 p-6 sm:flex-row sm:items-center sm:justify-between">
          <div className="flex items-start gap-3">
            <div className="rounded-2xl bg-primary/10 p-3 text-primary">
              <PiggyBank className="h-6 w-6" aria-hidden />
            </div>
            <div>
              <p className="font-semibold text-foreground">No investments yet</p>
              <p className="mt-1 text-sm text-muted-foreground">
                Start with the Genesis Plan — $10 · 18% ROI · 20 business days.
              </p>
            </div>
          </div>
          <Button asChild>
            <Link href="/app/earn">View Investment Plans</Link>
          </Button>
        </GlassCard>
      ) : (
        <div className="grid gap-3 lg:grid-cols-2">
          {list.map((inv, index) => {
            const daily = inv.daily_reward_ngn ?? 0;
            const expectedPayout =
              (Number(inv.amount_paid) || 0) * (1 + Math.max(0, Number(inv.roi_percent) || 0) / 100);
            const progress = Math.min(100, Math.max(0, inv.progress_pct ?? 0));
            const daysRemaining = inv.days_remaining ?? inv.lock_period_days ?? 0;

            return (
              <motion.article
                key={inv.id}
                initial={reduceMotion ? false : { opacity: 0, y: 12 }}
                whileInView={reduceMotion ? undefined : { opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ delay: index * 0.05, duration: 0.3 }}
              >
                <GlassCard className="h-full border-primary/15 bg-gradient-to-br from-primary/10 via-card/90 to-card">
                  <div className="flex items-start justify-between gap-2">
                    <div>
                      <h3 className="text-base font-bold text-foreground">
                        {inv.plan_name || "Genesis Plan"}
                      </h3>
                      <p className="mt-0.5 text-xs text-muted-foreground">
                        {inv.lock_period_days} business days · ROI {inv.roi_percent}%
                      </p>
                    </div>
                    <Badge
                      variant={
                        inv.status === "active"
                          ? "success"
                          : inv.status === "completed"
                            ? "secondary"
                            : "warning"
                      }
                      className="capitalize"
                    >
                      {String(inv.status).replace(/_/g, " ")}
                    </Badge>
                  </div>

                  <div className="mt-4 grid grid-cols-2 gap-2 sm:grid-cols-3">
                    <Metric label="Investment" value={formatCurrency(inv.amount_paid)} />
                    <Metric label="ROI" value={`${inv.roi_percent}%`} />
                    <Metric label="Daily Reward" value={formatCurrency(daily)} />
                    <Metric label="Today's Reward" value={formatCurrency(todayEarningsNgn || daily)} />
                    <Metric label="Days Remaining" value={`${daysRemaining}`} />
                    <Metric label="Expected Payout" value={formatCurrency(expectedPayout)} />
                  </div>

                  <div className="mt-4">
                    <div className="mb-1.5 flex items-center justify-between text-xs text-muted-foreground">
                      <span>Progress</span>
                      <span className="font-medium text-foreground">{Math.round(progress)}%</span>
                    </div>
                    <Progress value={progress} className="h-2" />
                  </div>

                  <div className="mt-3 flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-muted-foreground">
                    <span>
                      Withdrawal:{" "}
                      {cooldown.available
                        ? "Available now"
                        : `in ${cooldown.daysRemaining} day${cooldown.daysRemaining === 1 ? "" : "s"}`}
                      {!cooldown.available && cooldown.nextAvailableAt
                        ? ` (${formatWithdrawalNextAvailable(cooldown.nextAvailableAt)})`
                        : ""}
                    </span>
                    <span>Processing: up to {processingHours}h</span>
                  </div>
                </GlassCard>
              </motion.article>
            );
          })}
        </div>
      )}
    </SectionShell>
  );
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-xl border border-border/50 bg-background/40 px-3 py-2">
      <p className="text-[11px] text-muted-foreground">{label}</p>
      <p className="mt-0.5 text-sm font-semibold tabular-nums text-foreground">{value}</p>
    </div>
  );
}
