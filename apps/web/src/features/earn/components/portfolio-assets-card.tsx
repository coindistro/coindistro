"use client";

import * as React from "react";
import { Lock, Wallet } from "lucide-react";
import { formatCurrency } from "@/features/earn/utils";

export function PortfolioAssetsCard({
  availableNgn,
  lockedInvestmentNgn,
  profitEarnedNgn,
  withdrawableNgn,
  withdrawalsUnlocked,
  lockMessage,
  minReferrals,
  activeReferrals,
}: {
  availableNgn: number;
  lockedInvestmentNgn: number;
  profitEarnedNgn: number;
  withdrawableNgn: number;
  withdrawalsUnlocked: boolean;
  lockMessage?: string;
  minReferrals: number;
  activeReferrals: number;
}) {
  return (
    <section className="space-y-3">
      <div>
        <h2 className="text-lg font-bold text-foreground">Assets</h2>
        <p className="text-sm text-muted-foreground">
          Position breakdown — capital stays locked; profit unlocks with referrals
        </p>
      </div>
      <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
        <AssetTile
          label="Available Balance"
          value={formatCurrency(availableNgn)}
          hint="Free cash not locked in a plan"
          icon={<Wallet className="h-4 w-4 text-muted-foreground" />}
        />
        <AssetTile
          label="Locked Investment Capital"
          value={formatCurrency(lockedInvestmentNgn)}
          hint="Genesis plan principal (locked)"
          icon={<Lock className="h-4 w-4 text-fuchsia-400" />}
          accent="text-fuchsia-400"
        />
        <AssetTile
          label="Earned Profit"
          value={formatCurrency(profitEarnedNgn)}
          hint="Credited via earnings ledger"
          icon={<Wallet className="h-4 w-4 text-emerald-500" />}
          accent="text-emerald-500"
        />
        <div className="rounded-[1.25rem] border border-border/60 bg-card/80 p-4 shadow-sm backdrop-blur-sm">
          <div className="flex items-center gap-2 text-xs text-muted-foreground">
            <Lock className="h-4 w-4 text-amber-500" />
            Withdrawable Balance
          </div>
          {withdrawalsUnlocked ? (
            <p className="mt-2 text-xl font-bold tabular-nums text-primary">
              {formatCurrency(withdrawableNgn)}
            </p>
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
    </section>
  );
}

function AssetTile({
  label,
  value,
  hint,
  icon,
  accent,
}: {
  label: string;
  value: string;
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
      <p className="mt-1 text-[11px] text-muted-foreground">{hint}</p>
    </div>
  );
}
