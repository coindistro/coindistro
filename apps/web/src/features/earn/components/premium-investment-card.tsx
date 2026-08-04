"use client";

import * as React from "react";
import { motion, AnimatePresence } from "framer-motion";
import { ArrowUpRight, ChevronDown, Gift, TrendingUp, CalendarDays } from "lucide-react";
import type { InvestmentPlanConfig } from "@/features/earn/config/investment-plans";
import { formatCurrency } from "@/features/earn/utils";

interface PremiumInvestmentCardProps {
  plan: InvestmentPlanConfig;
  exchangeRate: number;
  onSelect: (plan: InvestmentPlanConfig) => void;
}

export function PremiumInvestmentCard({ plan, exchangeRate, onSelect }: PremiumInvestmentCardProps) {
  const [expanded, setExpanded] = React.useState(false);
  const ngnValue = plan.usdAmount * exchangeRate;

  return (
    <motion.div
      layout
      initial={{ opacity: 0, y: 24 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.4, ease: "easeOut" }}
      className="group relative overflow-hidden rounded-2xl border border-border bg-card/80 backdrop-blur-sm transition-colors hover:border-primary/40"
    >
      {/* Purple gradient glow */}
      <div className="pointer-events-none absolute -top-20 -right-20 h-40 w-40 rounded-full bg-primary/15 blur-3xl" />
      <div className="pointer-events-none absolute inset-0 bg-gradient-to-br from-primary/5 via-transparent to-transparent opacity-0 transition-opacity group-hover:opacity-100" />

      <div className="relative p-5">
        <div className="flex items-start justify-between gap-3">
          <div>
            <h3 className="text-lg font-bold text-foreground">{plan.name}</h3>
            <p className="mt-1 text-sm text-muted-foreground">{plan.description}</p>
          </div>
          <span className="rounded-lg bg-primary/10 px-3 py-1.5 text-sm font-bold tabular-nums text-primary">
            ${plan.usdAmount.toLocaleString()}
          </span>
        </div>

        {/* Key stats */}
        <div className="mt-4 grid grid-cols-3 gap-3">
          <div className="rounded-xl border border-border/60 bg-muted/40 p-3">
            <div className="flex items-center gap-1 text-xs text-muted-foreground">
              <TrendingUp className="h-3 w-3 text-amber-500" />
              Daily
            </div>
            <p className="mt-1 text-sm font-semibold text-foreground">{formatCurrency(plan.dailyRewardNgn)}</p>
          </div>
          <div className="rounded-xl border border-border/60 bg-muted/40 p-3">
            <div className="flex items-center gap-1 text-xs text-muted-foreground">
              <CalendarDays className="h-3 w-3 text-emerald-500" />
              Duration
            </div>
            <p className="mt-1 text-sm font-semibold text-foreground">{plan.workingDays}d</p>
          </div>
          <div className="rounded-xl border border-border/60 bg-muted/40 p-3">
            <div className="flex items-center gap-1 text-xs text-muted-foreground">
              <Gift className="h-3 w-3 text-cyan-500" />
              Referral
            </div>
            <p className="mt-1 text-sm font-semibold text-foreground">{plan.referralBonusPercent}%</p>
          </div>
        </div>

        <div className="mt-3 flex items-center justify-between rounded-xl border border-border/60 bg-background/50 px-4 py-3">
          <span className="text-xs text-muted-foreground">NGN Equivalent</span>
          <span className="text-sm font-semibold text-fuchsia-500">{formatCurrency(ngnValue)}</span>
        </div>

        {/* Expandable details */}
        <AnimatePresence initial={false}>
          {expanded ? (
            <motion.div
              initial={{ height: 0, opacity: 0 }}
              animate={{ height: "auto", opacity: 1 }}
              exit={{ height: 0, opacity: 0 }}
              transition={{ duration: 0.25, ease: "easeInOut" }}
              className="overflow-hidden"
            >
              <div className="mt-4 space-y-2 rounded-xl border border-border/60 bg-muted/30 p-4 text-sm">
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Monthly Reward</span>
                  <span className="font-semibold text-emerald-500">{formatCurrency(plan.monthlyRewardNgn)}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Referral Bonus</span>
                  <span className="font-semibold text-cyan-500">{plan.referralBonusPercent}%</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Min. Referrals</span>
                  <span className="font-semibold">{plan.minReferrals}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Estimated ROI</span>
                  <span className="font-semibold text-primary">
                    {((plan.monthlyRewardNgn / ngnValue) * 100).toFixed(1)}%
                  </span>
                </div>
              </div>
            </motion.div>
          ) : null}
        </AnimatePresence>

        {/* Actions */}
        <div className="mt-4 flex items-center gap-2">
          <button
            type="button"
            onClick={() => onSelect(plan)}
            className="flex flex-1 items-center justify-center gap-2 rounded-xl bg-gradient-to-r from-primary to-secondary px-4 py-3 text-sm font-semibold text-primary-foreground shadow-lg shadow-primary/20 transition-transform active:scale-[0.98]"
          >
            Invest Now
            <ArrowUpRight className="h-4 w-4" />
          </button>
          <button
            type="button"
            aria-label={expanded ? "Collapse plan details" : "Expand plan details"}
            onClick={() => setExpanded((value) => !value)}
            className="flex h-11 w-11 items-center justify-center rounded-xl border border-border bg-muted/40 text-muted-foreground transition-colors hover:border-primary/40 hover:text-primary"
          >
            <motion.span animate={{ rotate: expanded ? 180 : 0 }} transition={{ duration: 0.2 }}>
              <ChevronDown className="h-4 w-4" />
            </motion.span>
          </button>
        </div>
      </div>
    </motion.div>
  );
}