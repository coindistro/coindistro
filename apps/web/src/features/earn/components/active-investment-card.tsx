"use client";

import { motion } from "framer-motion";
import { CalendarDays, TrendingUp, Wallet } from "lucide-react";
import { formatCurrency } from "@/features/earn/utils";
import type { InvestmentSummary } from "@/lib/api/types";

export interface ActivePositionMetrics {
  capitalUsd: number;
  capitalNgn: number;
  profitUsd: number;
  profitNgn: number;
  currentValueUsd: number;
  currentValueNgn: number;
  roiPercent: number;
  exchangeRate: number;
}

interface ActiveInvestmentCardProps {
  investment: InvestmentSummary | null;
  totalDays: number;
  completedDays: number;
  daysRemaining: number;
  progress: number;
  dailyReward: number;
  totalEarned: number;
  expectedRoi: number;
  processingHours?: number;
  position?: ActivePositionMetrics | null;
}

function formatUsd(value: number) {
  return `$${value.toLocaleString(undefined, { maximumFractionDigits: 2 })}`;
}

export function ActiveInvestmentCard({
  investment,
  totalDays,
  completedDays,
  daysRemaining,
  progress,
  dailyReward,
  totalEarned,
  expectedRoi,
  processingHours = 24,
  position,
}: ActiveInvestmentCardProps) {
  if (!investment) {
    return (
      <div className="rounded-2xl border border-dashed border-border bg-card/40 p-8 text-center">
        <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-2xl bg-primary/10">
          <Wallet className="h-6 w-6 text-primary" />
        </div>
        <h3 className="mt-4 text-lg font-semibold text-foreground">No active investment</h3>
        <p className="mt-1 text-sm text-muted-foreground">
          Choose a Genesis plan to start earning daily rewards.
        </p>
      </div>
    );
  }

  const isActive = investment.status === "active";
  const statusLabel = isActive
    ? "Growing"
    : investment.status === "completed"
      ? "Completed"
      : "Pending";

  const statusColor = isActive
    ? "bg-emerald-500/10 text-emerald-500"
    : investment.status === "completed"
      ? "bg-cyan-500/10 text-cyan-500"
      : "bg-amber-500/10 text-amber-500";

  const capitalUsd = position?.capitalUsd ?? 0;
  const capitalNgn = position?.capitalNgn ?? (Number(investment.amount_paid) || 0);
  const profitUsd = position?.profitUsd ?? 0;
  const profitNgn = position?.profitNgn ?? totalEarned;
  const currentUsd = position?.currentValueUsd ?? capitalUsd + profitUsd;
  const currentNgn = position?.currentValueNgn ?? capitalNgn + profitNgn;
  const roi =
    position?.roiPercent ??
    (capitalUsd > 0 ? (profitUsd / capitalUsd) * 100 : expectedRoi);

  return (
    <motion.div
      initial={{ opacity: 0, y: 16 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.4 }}
      className="relative overflow-hidden rounded-2xl border border-primary/20 bg-gradient-to-br from-primary/10 via-card to-transparent p-5"
    >
      <div className="pointer-events-none absolute -top-16 -right-16 h-40 w-40 rounded-full bg-primary/20 blur-3xl" />

      <div className="relative">
        <div className="flex items-start justify-between gap-3">
          <div>
            <h3 className="text-lg font-bold text-foreground">
              {investment.plan_name || "Genesis Plan"}
            </h3>
            <p className="mt-1 text-sm text-muted-foreground">
              {formatDate(investment.started_at)} — {formatDate(investment.matures_at)}
            </p>
          </div>
          <span className={`rounded-lg px-3 py-1 text-xs font-semibold ${statusColor}`}>
            {statusLabel}
          </span>
        </div>

        <div className="mt-4 grid grid-cols-2 gap-3 sm:grid-cols-4">
          <div className="rounded-xl border border-border/60 bg-background/50 p-3">
            <p className="text-xs text-muted-foreground">Investment</p>
            <p className="mt-1 text-lg font-bold tabular-nums text-foreground">
              {capitalUsd > 0 ? formatUsd(capitalUsd) : formatCurrency(capitalNgn)}
            </p>
            {capitalUsd > 0 && (
              <p className="text-[11px] text-muted-foreground">{formatCurrency(capitalNgn)}</p>
            )}
          </div>
          <div className="rounded-xl border border-border/60 bg-background/50 p-3">
            <p className="text-xs text-muted-foreground">Current Value</p>
            <p className="mt-1 text-lg font-bold tabular-nums text-primary">
              {formatUsd(currentUsd)}
            </p>
            <p className="text-[11px] text-muted-foreground">{formatCurrency(currentNgn)}</p>
          </div>
          <div className="rounded-xl border border-border/60 bg-background/50 p-3">
            <p className="text-xs text-muted-foreground">Profit</p>
            <p className="mt-1 text-lg font-bold tabular-nums text-emerald-500">
              +{formatUsd(profitUsd)}
            </p>
            <p className="text-[11px] text-muted-foreground">+{formatCurrency(profitNgn)}</p>
          </div>
          <div className="rounded-xl border border-border/60 bg-background/50 p-3">
            <p className="text-xs text-muted-foreground">ROI</p>
            <p className="mt-1 text-lg font-bold tabular-nums text-amber-500">
              {roi.toFixed(2)}%
            </p>
            <p className="text-[11px] text-muted-foreground">On capital</p>
          </div>
        </div>

        <div className="mt-3 grid grid-cols-2 gap-3 sm:grid-cols-3">
          <div className="rounded-xl border border-border/40 bg-muted/20 px-3 py-2">
            <p className="text-[11px] text-muted-foreground">Daily Reward</p>
            <p className="text-sm font-semibold tabular-nums">{formatCurrency(dailyReward)}</p>
          </div>
          <div className="rounded-xl border border-border/40 bg-muted/20 px-3 py-2">
            <p className="text-[11px] text-muted-foreground">Plan ROI setting</p>
            <p className="text-sm font-semibold tabular-nums">{expectedRoi.toFixed(1)}%</p>
          </div>
          <div className="rounded-xl border border-border/40 bg-muted/20 px-3 py-2 sm:col-span-1 col-span-2">
            <p className="text-[11px] text-muted-foreground">Status</p>
            <p className="text-sm font-semibold">{statusLabel}</p>
          </div>
        </div>

        <div className="mt-4">
          <div className="mb-2 flex items-center justify-between text-sm">
            <span className="text-muted-foreground">Progress</span>
            <span className="font-semibold text-foreground">
              Day {completedDays} / {totalDays}
            </span>
          </div>
          <div className="h-2.5 overflow-hidden rounded-full bg-muted">
            <motion.div
              initial={{ width: 0 }}
              animate={{ width: `${Math.min(100, progress)}%` }}
              transition={{ duration: 0.8, ease: "easeOut" }}
              className="h-full rounded-full bg-gradient-to-r from-primary to-secondary"
            />
          </div>
          <div className="mt-2 flex items-center justify-between text-xs text-muted-foreground">
            <span className="flex items-center gap-1">
              <CalendarDays className="h-3 w-3" /> {daysRemaining} days remaining
            </span>
            <span className="flex items-center gap-1">
              <TrendingUp className="h-3 w-3" /> {Math.round(progress)}% complete
            </span>
          </div>
          <p className="mt-3 text-xs text-muted-foreground">
            Capital remains locked in the plan. Withdrawal Processing: Up to {processingHours} Hours
            after unlock.
          </p>
        </div>
      </div>
    </motion.div>
  );
}

function formatDate(value?: string | null) {
  if (!value) return "—";
  return new Date(value).toLocaleDateString(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}
