"use client";

import * as React from "react";
import Link from "next/link";
import { motion, useReducedMotion } from "framer-motion";
import { ArrowDownLeft, ArrowLeftRight, ArrowUpRight, Eye } from "lucide-react";
import { SectionShell, GlassCard, SectionSkeleton } from "./section-shell";

export interface AssetWallet {
  currency: string;
  label: string;
  balance: number;
  symbol: string;
  available?: boolean;
}

export function AssetsSection({
  wallets,
  loading,
}: {
  wallets: AssetWallet[];
  loading?: boolean;
}) {
  const reduceMotion = useReducedMotion();

  return (
    <SectionShell
      id="assets"
      title="Assets"
      description="Multi-currency wallet overview"
      actionHref="/app/wallet"
      actionLabel="View Wallet"
    >
      {loading ? (
        <SectionSkeleton rows={2} />
      ) : (
        <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-5">
          {wallets.map((wallet, index) => (
            <motion.div
              key={wallet.currency}
              initial={reduceMotion ? false : { opacity: 0, y: 10 }}
              whileInView={reduceMotion ? undefined : { opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ delay: index * 0.04, duration: 0.3 }}
            >
              <GlassCard className="group h-full transition-colors hover:border-primary/30">
                <div className="flex items-start justify-between gap-2">
                  <div>
                    <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
                      {wallet.label}
                    </p>
                    <p className="mt-2 text-xl font-bold tabular-nums text-foreground">
                      {wallet.symbol}
                      {wallet.balance.toLocaleString(undefined, {
                        maximumFractionDigits: wallet.currency === "BTC" || wallet.currency === "ETH" ? 6 : 2,
                      })}
                    </p>
                    <p className="mt-0.5 text-xs text-muted-foreground">{wallet.currency}</p>
                  </div>
                  <span className="rounded-lg bg-primary/10 px-2 py-1 text-[10px] font-semibold text-primary">
                    {wallet.available === false ? "Soon" : "Live"}
                  </span>
                </div>
                <div className="mt-4 grid grid-cols-2 gap-1.5">
                  <AssetAction href="/app/wallet" icon={<ArrowDownLeft className="h-3 w-3" />} label="Deposit" />
                  <AssetAction href="/app/wallet" icon={<ArrowUpRight className="h-3 w-3" />} label="Withdraw" />
                  <AssetAction href="/app/wallet" icon={<ArrowLeftRight className="h-3 w-3" />} label="Transfer" />
                  <AssetAction href="/app/wallet" icon={<Eye className="h-3 w-3" />} label="View" />
                </div>
              </GlassCard>
            </motion.div>
          ))}
        </div>
      )}
    </SectionShell>
  );
}

function AssetAction({
  href,
  icon,
  label,
}: {
  href: string;
  icon: React.ReactNode;
  label: string;
}) {
  return (
    <Link
      href={href}
      className="inline-flex items-center justify-center gap-1 rounded-xl border border-border/50 bg-background/40 px-2 py-1.5 text-[11px] font-medium text-muted-foreground transition-colors hover:border-primary/40 hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
    >
      {icon}
      {label}
    </Link>
  );
}
