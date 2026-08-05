"use client";

import Link from "next/link";
import {
  ArrowDownLeft,
  ArrowUpRight,
  Gift,
  LineChart,
  PiggyBank,
  Users,
} from "lucide-react";
import { Badge, Button } from "@coindistro/cds";
import { formatCurrency } from "@/features/earn/utils";
import { formatRelative } from "@/lib/utils/format";
import { SectionShell, GlassCard, SectionSkeleton } from "./section-shell";

export type OverviewTxType =
  | "deposit"
  | "withdrawal"
  | "investment"
  | "reward"
  | "referral"
  | "trade";

export interface OverviewTransaction {
  id: string;
  type: OverviewTxType;
  amount: number;
  currency: string;
  status: string;
  timestamp: string;
  label?: string;
}

const typeMeta: Record<
  OverviewTxType,
  { icon: typeof ArrowDownLeft; label: string; positive: boolean }
> = {
  deposit: { icon: ArrowDownLeft, label: "Deposit", positive: true },
  withdrawal: { icon: ArrowUpRight, label: "Withdrawal", positive: false },
  investment: { icon: PiggyBank, label: "Investment", positive: false },
  reward: { icon: Gift, label: "Reward", positive: true },
  referral: { icon: Users, label: "Referral", positive: true },
  trade: { icon: LineChart, label: "Trade", positive: false },
};

function statusVariant(status: string): "success" | "warning" | "danger" | "secondary" | "outline" {
  const s = status.toLowerCase();
  if (s.includes("complete") || s.includes("success") || s === "paid") return "success";
  if (s.includes("pending") || s.includes("process") || s.includes("review")) return "warning";
  if (s.includes("fail") || s.includes("reject") || s.includes("cancel")) return "danger";
  return "secondary";
}

export function RecentTransactionsSection({
  items,
  loading,
}: {
  items: OverviewTransaction[];
  loading?: boolean;
}) {
  return (
    <SectionShell
      id="transactions"
      title="Recent Transactions"
      description="Deposits, withdrawals, investments, rewards & more"
      actionHref="/app/wallet"
      actionLabel="View All"
    >
      {loading ? (
        <SectionSkeleton rows={3} />
      ) : items.length === 0 ? (
        <GlassCard className="border-dashed p-6 text-center">
          <p className="font-semibold text-foreground">No transactions yet</p>
          <p className="mt-1 text-sm text-muted-foreground">
            Your deposits, investments, and rewards will appear here.
          </p>
          <Button className="mt-4" size="sm" asChild>
            <Link href="/app/earn">Make your first investment</Link>
          </Button>
        </GlassCard>
      ) : (
        <GlassCard className="divide-y divide-border/50 p-0">
          <ul>
            {items.slice(0, 8).map((tx) => {
              const meta = typeMeta[tx.type] ?? typeMeta.deposit;
              const Icon = meta.icon;
              const amountLabel =
                tx.currency === "NGN"
                  ? formatCurrency(tx.amount)
                  : `${tx.currency} ${tx.amount.toLocaleString(undefined, { maximumFractionDigits: 4 })}`;
              return (
                <li
                  key={tx.id}
                  className="flex items-center gap-3 px-4 py-3 transition-colors hover:bg-muted/25"
                >
                  <div
                    className={`flex h-9 w-9 shrink-0 items-center justify-center rounded-full ${
                      meta.positive
                        ? "bg-emerald-500/15 text-emerald-500"
                        : "bg-muted text-muted-foreground"
                    }`}
                  >
                    <Icon className="h-4 w-4" aria-hidden />
                  </div>
                  <div className="min-w-0 flex-1">
                    <div className="flex flex-wrap items-center gap-2">
                      <p className="text-sm font-medium text-foreground">
                        {tx.label || meta.label}
                      </p>
                      <Badge variant={statusVariant(tx.status)} className="capitalize text-[10px]">
                        {tx.status.replace(/_/g, " ")}
                      </Badge>
                    </div>
                    <p className="text-xs text-muted-foreground">
                      {formatRelative(tx.timestamp)} · {tx.currency}
                    </p>
                  </div>
                  <p
                    className={`shrink-0 text-sm font-semibold tabular-nums ${
                      meta.positive ? "text-emerald-500" : "text-foreground"
                    }`}
                  >
                    {meta.positive ? "+" : "−"}
                    {amountLabel}
                  </p>
                </li>
              );
            })}
          </ul>
        </GlassCard>
      )}
    </SectionShell>
  );
}
