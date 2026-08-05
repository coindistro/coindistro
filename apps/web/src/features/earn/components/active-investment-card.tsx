"use client";

import { motion } from "framer-motion";
import { CalendarDays, TrendingUp, Wallet } from "lucide-react";
import { formatCurrency } from "@/features/earn/utils";
import type { InvestmentSummary } from "@/lib/api/types";

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
}: ActiveInvestmentCardProps) {
  if (!investment) {
    return (
      <div className="rounded-2xl border border-dashed border-border bg-card/40 p-8 text-center">
        <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-2xl bg-primary/10">
          <Wallet className="h-6 w-6 text-primary" />
        </div>
        <h3 className="mt-4 text-lg font-semibold text-foreground">No active investment</h3>
        <p className="mt-1 text-sm text-muted-foreground">
          Choose the Genesis Plan ($10 / ₦14,000 · 18% ROI) to start earning daily rewards.
        </p>
      </div>
    );
  }

  const statusLabel =
    investment.status === "active"
      ? "Active"
      : investment.status === "completed"
        ? "Completed"
        : "Pending";

  const statusColor =
    investment.status === "active"
      ? "bg-emerald-500/10 text-emerald-500"
      : investment.status === "completed"
        ? "bg-cyan-500/10 text-cyan-500"
        : "bg-amber-500/10 text-amber-500";

  const expectedPayout =
    (Number(investment.amount_paid) || 0) * (1 + Math.max(0, expectedRoi) / 100);

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

        <div className="mt-4 grid grid-cols-2 gap-3 sm:grid-cols-3">
          <div className="rounded-xl border border-border/60 bg-background/50 p-3">
            <p className="text-xs text-muted-foreground">Investment</p>
            <p className="mt-1 text-lg font-bold tabular-nums text-foreground">
              {formatCurrency(investment.amount_paid)}
            </p>
          </div>
          <div className="rounded-xl border border-border/60 bg-background/50 p-3">
            <p className="text-xs text-muted-foreground">Daily Reward</p>
            <p className="mt-1 text-lg font-bold tabular-nums text-amber-500">
              {formatCurrency(dailyReward)}
            </p>
          </div>
          <div className="rounded-xl border border-border/60 bg-background/50 p-3">
            <p className="text-xs text-muted-foreground">Remaining Days</p>
            <p className="mt-1 text-lg font-bold tabular-nums text-foreground">{daysRemaining}</p>
          </div>
          <div className="rounded-xl border border-border/60 bg-background/50 p-3">
            <p className="text-xs text-muted-foreground">Expected ROI</p>
            <p className="mt-1 text-lg font-bold tabular-nums text-primary">
              {expectedRoi.toFixed(1)}%
            </p>
          </div>
          <div className="rounded-xl border border-border/60 bg-background/50 p-3">
            <p className="text-xs text-muted-foreground">Expected Payout</p>
            <p className="mt-1 text-lg font-bold tabular-nums text-emerald-500">
              {formatCurrency(expectedPayout)}
            </p>
          </div>
          <div className="rounded-xl border border-border/60 bg-background/50 p-3">
            <p className="text-xs text-muted-foreground">Total Earned</p>
            <p className="mt-1 text-lg font-bold tabular-nums text-emerald-500">
              {formatCurrency(totalEarned)}
            </p>
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
            Withdrawal Processing: Up to {processingHours} Hours
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