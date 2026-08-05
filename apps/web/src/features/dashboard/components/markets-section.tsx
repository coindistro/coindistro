"use client";

import * as React from "react";
import Link from "next/link";
import { motion, useReducedMotion } from "framer-motion";
import { Button } from "@coindistro/cds";
import { SectionShell, GlassCard } from "./section-shell";
import { MiniSparkline } from "./mini-sparkline";

/** Curated snapshot until live Markets API ships. Presentation only. */
export interface MarketTicker {
  symbol: string;
  name: string;
  price: number;
  change24h: number;
  sparkline: number[];
  future?: boolean;
}

const DEFAULT_MARKETS: MarketTicker[] = [
  { symbol: "BTC", name: "Bitcoin", price: 97420.12, change24h: 1.84, sparkline: [92, 93, 91, 95, 97, 96, 98, 99, 97, 100] },
  { symbol: "ETH", name: "Ethereum", price: 3521.4, change24h: -0.62, sparkline: [80, 82, 79, 78, 81, 80, 77, 79, 78, 76] },
  { symbol: "BNB", name: "BNB", price: 612.3, change24h: 0.91, sparkline: [70, 71, 72, 71, 73, 74, 73, 75, 74, 76] },
  { symbol: "SOL", name: "Solana", price: 178.55, change24h: 3.12, sparkline: [50, 52, 55, 54, 58, 60, 59, 63, 65, 68] },
  { symbol: "XRP", name: "XRP", price: 2.41, change24h: -1.24, sparkline: [60, 58, 59, 57, 56, 55, 54, 53, 52, 51] },
  { symbol: "ADA", name: "Cardano", price: 0.82, change24h: 0.44, sparkline: [40, 41, 40, 42, 41, 43, 42, 44, 43, 45] },
  { symbol: "DOGE", name: "Dogecoin", price: 0.28, change24h: 2.05, sparkline: [30, 32, 31, 34, 33, 36, 35, 37, 38, 40] },
  { symbol: "CDT", name: "CoinDistro", price: 0, change24h: 0, sparkline: [20, 20, 21, 20, 22, 21, 20, 21, 20, 20], future: true },
];

function formatPrice(price: number, future?: boolean) {
  if (future) return "—";
  if (price >= 1000) return `$${price.toLocaleString(undefined, { maximumFractionDigits: 2 })}`;
  if (price >= 1) return `$${price.toFixed(2)}`;
  return `$${price.toFixed(4)}`;
}

export function MarketsSection({ markets = DEFAULT_MARKETS }: { markets?: MarketTicker[] }) {
  const reduceMotion = useReducedMotion();
  const top = markets.slice(0, 10);

  return (
    <SectionShell
      id="markets"
      title="Markets"
      description="Top assets at a glance"
      actionHref="/app/markets"
      actionLabel="View Markets"
    >
      <GlassCard className="overflow-hidden p-0">
        <div className="divide-y divide-border/50">
          <div className="grid grid-cols-[1.2fr_1fr_0.8fr_auto] gap-2 px-4 py-2.5 text-[11px] font-medium uppercase tracking-wide text-muted-foreground sm:grid-cols-[1.4fr_1fr_0.9fr_90px]">
            <span>Asset</span>
            <span className="text-right">Price</span>
            <span className="text-right">24h</span>
            <span className="hidden text-right sm:block">Trend</span>
          </div>
          {top.map((m, index) => {
            const positive = m.change24h >= 0;
            return (
              <motion.div
                key={m.symbol}
                initial={reduceMotion ? false : { opacity: 0 }}
                whileInView={reduceMotion ? undefined : { opacity: 1 }}
                viewport={{ once: true }}
                transition={{ delay: index * 0.02 }}
                className="grid grid-cols-[1.2fr_1fr_0.8fr_auto] items-center gap-2 px-4 py-3 transition-colors hover:bg-muted/30 sm:grid-cols-[1.4fr_1fr_0.9fr_90px]"
              >
                <div className="min-w-0">
                  <div className="flex items-center gap-2">
                    <span className="font-semibold text-foreground">{m.symbol}</span>
                    {m.future ? (
                      <span className="rounded bg-primary/10 px-1.5 py-0.5 text-[10px] font-medium text-primary">
                        Future
                      </span>
                    ) : null}
                  </div>
                  <p className="truncate text-xs text-muted-foreground">{m.name}</p>
                </div>
                <p className="text-right text-sm font-medium tabular-nums text-foreground">
                  {formatPrice(m.price, m.future)}
                </p>
                <p
                  className={`text-right text-sm font-semibold tabular-nums ${
                    m.future
                      ? "text-muted-foreground"
                      : positive
                        ? "text-emerald-500"
                        : "text-rose-500"
                  }`}
                >
                  {m.future ? "—" : `${positive ? "+" : ""}${m.change24h.toFixed(2)}%`}
                </p>
                <div className="hidden justify-end sm:flex">
                  <MiniSparkline data={m.sparkline} positive={positive || !!m.future} />
                </div>
              </motion.div>
            );
          })}
        </div>
        <div className="border-t border-border/50 p-3">
          <Button variant="outline" size="sm" className="w-full" asChild>
            <Link href="/app/markets">View Markets</Link>
          </Button>
        </div>
      </GlassCard>
    </SectionShell>
  );
}
