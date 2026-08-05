"use client";

import * as React from "react";
import { Lock, Wallet, TrendingUp, Gift } from "lucide-react";
import { formatCurrency } from "@/features/earn/utils";

function formatUsd(value: number) {
  return `$${value.toLocaleString(undefined, { maximumFractionDigits: 2 })}`;
}

export function PortfolioAssetsCard({
  availableUsd,
  lockedUsd,
  portfolioUsd,
  capitalUsd,
  profitUsd,
  referralUsd,
  withdrawableUsd,
  withdrawalsUnlocked,
  lockMessage,
  minReferrals,
  activeReferrals,
  exchangeRate,
}: {
  availableUsd: number;
  lockedUsd: number;
  portfolioUsd: number;
  capitalUsd: number;
  profitUsd: number;
  referralUsd: number;
  withdrawableUsd: number;
  withdrawalsUnlocked: boolean;
  lockMessage?: string;
  minReferrals: number;
  activeReferrals: number;
  exchangeRate: number;
}) {
  const rate = exchangeRate > 0 ? exchangeRate : 1400;
  return (
    <section className="space-y-3">
      <div>
        <h2 className="text-lg font-bold text-foreground">Wallet</h2>
        <p className="text-sm text-muted-foreground">
          Available cash vs locked capital, profit, and referral rewards
        </p>
      </div>
      <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
        <AssetTile
          label="Available Balance"
          value={formatUsd(availableUsd)}
          sub={`≈ ${formatCurrency(availableUsd * rate)}`}
          hint="Withdrawable immediately (after unlock)"
          icon={<Wallet className="h-4 w-4 text-muted-foreground" />}
        />
        <AssetTile
          label="Locked Balance"
          value={formatUsd(lockedUsd)}
          sub={`≈ ${formatCurrency(lockedUsd * rate)}`}
          hint="Capital + locked profit/referrals"
          icon={<Lock className="h-4 w-4 text-fuchsia-400" />}
          accent="text-fuchsia-400"
        />
        <AssetTile
          label="Portfolio Value"
          value={formatUsd(portfolioUsd)}
          sub={`≈ ${formatCurrency(portfolioUsd * rate)}`}
          hint="Available + Locked"
          icon={<TrendingUp className="h-4 w-4 text-primary" />}
          accent="text-primary"
        />
        <div className="rounded-[1.25rem] border border-border/60 bg-card/80 p-4 shadow-sm backdrop-blur-sm">
          <div className="flex items-center gap-2 text-xs text-muted-foreground">
            <Lock className="h-4 w-4 text-amber-500" />
            Withdrawable Balance
          </div>
          {withdrawalsUnlocked ? (
            <>
              <p className="mt-2 text-xl font-bold tabular-nums text-primary">
                {formatUsd(withdrawableUsd)}
              </p>
              <p className="text-xs text-muted-foreground">≈ {formatCurrency(withdrawableUsd * rate)}</p>
            </>
          ) : (
            <>
              <p className="mt-2 text-xl font-bold uppercase tracking-wide text-amber-500">
                Locked
              </p>
              <p className="mt-2 text-xs leading-relaxed text-muted-foreground">
                {lockMessage ||
                  `Complete ${minReferrals} successful referrals to unlock withdrawals.`}
              </p>
              <p className="mt-1 text-xs font-semibold text-foreground">
                Progress: {activeReferrals} / {minReferrals}
              </p>
            </>
          )}
        </div>
      </div>

      <div className="grid gap-3 sm:grid-cols-3">
        <AssetTile
          label="Invested Capital"
          value={formatUsd(capitalUsd)}
          sub={`≈ ${formatCurrency(capitalUsd * rate)}`}
          hint="Locked in active plans"
          icon={<Lock className="h-4 w-4 text-muted-foreground" />}
        />
        <AssetTile
          label="Profit Earned"
          value={formatUsd(profitUsd)}
          sub={`≈ ${formatCurrency(profitUsd * rate)}`}
          hint="From earnings ledger"
          icon={<TrendingUp className="h-4 w-4 text-emerald-500" />}
          accent="text-emerald-500"
        />
        <AssetTile
          label="Referral Earnings"
          value={formatUsd(referralUsd)}
          sub={`≈ ${formatCurrency(referralUsd * rate)}`}
          hint="Referral rewards"
          icon={<Gift className="h-4 w-4 text-cyan-500" />}
          accent="text-cyan-500"
        />
      </div>
    </section>
  );
}

function AssetTile({
  label,
  value,
  sub,
  hint,
  icon,
  accent,
}: {
  label: string;
  value: string;
  sub?: string;
  hint: string;
  icon: React.ReactNode;
  accent?: string;
}) {
  return (
    <div className="rounded-[1.25rem] border border-border/60 bg-card/80 p-4 shadow-sm backdrop-blur-sm">
      <div className="flex items-center gap-2 text-xs text-muted-foreground">
        {icon}
        {label}
      </div>
      <p className={`mt-2 text-xl font-bold tabular-nums ${accent ?? "text-foreground"}`}>
        {value}
      </p>
      {sub ? <p className="text-xs text-muted-foreground">{sub}</p> : null}
      <p className="mt-1 text-[11px] text-muted-foreground">{hint}</p>
    </div>
  );
}
