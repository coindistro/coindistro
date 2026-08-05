"use client";

import * as React from "react";
import { motion, useReducedMotion } from "framer-motion";
import { Gift, Lock, RefreshCw, Sparkles, Wallet } from "lucide-react";
import { Skeleton } from "@coindistro/cds";
import { useCountUp } from "@/features/earn/components/use-count-up";
import { formatCurrency, greetingForHour } from "@/features/earn/utils";

export interface OverviewHeroProps {
  firstName: string;
  portfolioUsd: number;
  portfolioNgn: number;
  todayEarningsNgn: number;
  availableBalanceNgn: number;
  lockedInvestments: number;
  referralEarningsNgn: number;
  exchangeRate: number;
  loading?: boolean;
  lastUpdated?: Date | null;
  onRefresh?: () => void;
  refreshing?: boolean;
}

function formatCompact(value: number, currency: "USD" | "NGN") {
  if (currency === "USD") {
    if (value >= 1_000_000) return `$${(value / 1_000_000).toFixed(2)}M`;
    if (value >= 1_000) return `$${(value / 1_000).toFixed(2)}K`;
    return `$${value.toLocaleString(undefined, { maximumFractionDigits: 2 })}`;
  }
  return formatCurrency(value);
}

export function OverviewHero({
  firstName,
  portfolioUsd,
  portfolioNgn,
  todayEarningsNgn,
  availableBalanceNgn,
  lockedInvestments,
  referralEarningsNgn,
  exchangeRate,
  loading,
  lastUpdated,
  onRefresh,
  refreshing,
}: OverviewHeroProps) {
  const reduceMotion = useReducedMotion();
  const [currency, setCurrency] = React.useState<"USD" | "NGN">("USD");
  const greeting = greetingForHour();

  const portfolioValue = currency === "USD" ? portfolioUsd : portfolioNgn;
  const animatedPortfolio = useCountUp(portfolioValue, 800, !loading && !reduceMotion);
  const animatedToday = useCountUp(todayEarningsNgn, 800, !loading && !reduceMotion);
  const animatedAvailable = useCountUp(availableBalanceNgn, 800, !loading && !reduceMotion);
  const animatedReferral = useCountUp(referralEarningsNgn, 800, !loading && !reduceMotion);
  const animatedLocked = useCountUp(lockedInvestments, 800, !loading && !reduceMotion);

  return (
    <motion.section
      initial={reduceMotion ? false : { opacity: 0, y: 16 }}
      animate={reduceMotion ? undefined : { opacity: 1, y: 0 }}
      transition={{ duration: 0.4 }}
      className="relative overflow-hidden rounded-[1.25rem] border border-primary/20 bg-gradient-to-br from-primary/20 via-card to-card p-5 shadow-[0_12px_40px_rgba(88,28,135,0.18)] sm:p-7"
      aria-label="Portfolio overview"
    >
      <div className="pointer-events-none absolute -right-20 -top-24 h-64 w-64 rounded-full bg-primary/25 blur-3xl" />
      <div className="pointer-events-none absolute -bottom-28 -left-16 h-56 w-56 rounded-full bg-secondary/15 blur-3xl" />

      <div className="relative">
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div>
            <p className="text-sm text-muted-foreground">
              {greeting}, <span className="font-medium text-foreground">{firstName}</span>
            </p>
            <h1 className="mt-1 text-2xl font-bold tracking-tight text-foreground sm:text-3xl">
              Portfolio Value
            </h1>
          </div>

          <div className="flex items-center gap-2">
            <div
              className="flex rounded-xl border border-border/70 bg-background/50 p-1 backdrop-blur-sm"
              role="group"
              aria-label="Currency display"
            >
              {(["USD", "NGN"] as const).map((c) => (
                <button
                  key={c}
                  type="button"
                  onClick={() => setCurrency(c)}
                  aria-pressed={currency === c}
                  className={`rounded-lg px-3 py-1.5 text-xs font-semibold transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring ${
                    currency === c
                      ? "bg-primary text-primary-foreground"
                      : "text-muted-foreground hover:text-foreground"
                  }`}
                >
                  {c}
                </button>
              ))}
            </div>
            <button
              type="button"
              onClick={onRefresh}
              disabled={refreshing}
              aria-label="Refresh portfolio data"
              className="flex h-9 w-9 items-center justify-center rounded-xl border border-border/70 bg-background/50 text-muted-foreground transition-colors hover:text-primary focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-60"
            >
              <RefreshCw className={`h-4 w-4 ${refreshing ? "animate-spin" : ""}`} />
            </button>
          </div>
        </div>

        <div className="mt-5">
          {loading ? (
            <Skeleton className="h-12 w-56" />
          ) : (
            <p className="text-4xl font-bold tabular-nums tracking-tight text-foreground sm:text-5xl">
              {formatCompact(animatedPortfolio, currency)}
            </p>
          )}
          <p className="mt-2 text-sm text-muted-foreground">
            {currency === "USD"
              ? `≈ ${formatCurrency(portfolioNgn)} NGN`
              : `≈ $${portfolioUsd.toLocaleString(undefined, { maximumFractionDigits: 2 })} USD`}
            {exchangeRate > 0 ? (
              <span className="ml-2 text-xs">· 1 USD = {formatCurrency(exchangeRate)}</span>
            ) : null}
          </p>
          {lastUpdated ? (
            <p className="mt-1 text-xs text-muted-foreground">
              Last updated{" "}
              {lastUpdated.toLocaleTimeString(undefined, {
                hour: "2-digit",
                minute: "2-digit",
              })}
            </p>
          ) : null}
        </div>

        <div className="mt-6 grid grid-cols-2 gap-3 lg:grid-cols-4">
          <HeroStat
            icon={<Sparkles className="h-3.5 w-3.5 text-amber-500" />}
            label="Today's Earnings"
            value={loading ? "…" : formatCurrency(animatedToday)}
            tone="text-amber-500"
          />
          <HeroStat
            icon={<Wallet className="h-3.5 w-3.5 text-primary" />}
            label="Available Balance"
            value={loading ? "…" : formatCurrency(animatedAvailable)}
            tone="text-primary"
          />
          <HeroStat
            icon={<Lock className="h-3.5 w-3.5 text-fuchsia-400" />}
            label="Locked Investments"
            value={
              loading
                ? "…"
                : currency === "USD"
                  ? `$${animatedLocked.toLocaleString(undefined, { maximumFractionDigits: 2 })}`
                  : formatCurrency(animatedLocked * (exchangeRate || 0))
            }
            tone="text-fuchsia-400"
          />
          <HeroStat
            icon={<Gift className="h-3.5 w-3.5 text-cyan-400" />}
            label="Referral Earnings"
            value={loading ? "…" : formatCurrency(animatedReferral)}
            tone="text-cyan-400"
          />
        </div>
      </div>
    </motion.section>
  );
}

function HeroStat({
  icon,
  label,
  value,
  tone,
}: {
  icon: React.ReactNode;
  label: string;
  value: string;
  tone: string;
}) {
  return (
    <div className="rounded-2xl border border-border/50 bg-background/45 p-3.5 backdrop-blur-sm">
      <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
        {icon}
        {label}
      </div>
      <p className={`mt-1.5 text-base font-bold tabular-nums sm:text-lg ${tone}`}>{value}</p>
    </div>
  );
}
