"use client";

import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@coindistro/cds";
import { ArrowRight, Check } from "lucide-react";
import type { InvestmentPlanConfig } from "@/features/earn/config/investment-plans";
import { formatCurrency } from "@/features/earn/utils";

interface InvestmentCardProps {
  plan: InvestmentPlanConfig;
  exchangeRate: number;
  onSelect: (plan: InvestmentPlanConfig) => void;
}

export function InvestmentCard({ plan, exchangeRate, onSelect }: InvestmentCardProps) {
  const ngnEquivalent = plan.usdAmount * exchangeRate;

  return (
    <Card className="group relative overflow-hidden border-primary/20 bg-gradient-to-br from-primary/5 via-card to-cyan-500/5 backdrop-blur-sm transition-all duration-300 hover:-translate-y-1 hover:border-primary/40 hover:shadow-xl hover:shadow-primary/10">
      <div className="absolute inset-0 bg-gradient-to-br from-primary/10 via-transparent to-cyan-400/5 opacity-0 transition-opacity group-hover:opacity-100" />

      <CardHeader className="relative">
        <CardTitle className="text-xl font-bold text-primary">{plan.name}</CardTitle>
        <p className="text-sm text-muted-foreground">{plan.description}</p>
      </CardHeader>

      <CardContent className="relative space-y-4">
        <div className="space-y-2">
          <div className="flex items-baseline justify-between">
            <span className="text-sm text-muted-foreground">Investment</span>
            <span className="text-2xl font-bold tabular-nums">${plan.usdAmount.toLocaleString()}</span>
          </div>
          <div className="flex items-baseline justify-between">
            <span className="text-sm text-muted-foreground">NGN Equivalent</span>
            <span className="text-lg font-semibold text-fuchsia-600 tabular-nums">
              {formatCurrency(ngnEquivalent)}
            </span>
          </div>
        </div>

        <div className="space-y-2 rounded-lg border border-border/50 bg-muted/30 p-3">
          <div className="flex justify-between text-sm">
            <span className="text-muted-foreground">Daily Reward</span>
            <span className="font-semibold text-amber-600">{formatCurrency(plan.dailyRewardNgn)}</span>
          </div>
          <div className="flex justify-between text-sm">
            <span className="text-muted-foreground">Working Days</span>
            <span className="font-semibold">{plan.workingDays}</span>
          </div>
          <div className="flex justify-between text-sm">
            <span className="text-muted-foreground">Monthly Reward</span>
            <span className="font-semibold text-emerald-600">{formatCurrency(plan.monthlyRewardNgn)}</span>
          </div>
          <div className="flex justify-between text-sm">
            <span className="text-muted-foreground">Referral Bonus</span>
            <span className="font-semibold text-cyan-600">{plan.referralBonusPercent}%</span>
          </div>
          <div className="flex justify-between text-sm">
            <span className="text-muted-foreground">Min. Referrals</span>
            <span className="font-semibold">{plan.minReferrals}</span>
          </div>
        </div>

        <ul className="space-y-1.5">
          {plan.features.map((feature) => (
            <li key={feature} className="flex items-center gap-2 text-xs text-muted-foreground">
              <Check className="h-3 w-3 text-emerald-500" />
              {feature}
            </li>
          ))}
        </ul>
      </CardContent>

      <CardFooter className="relative">
        <button
          type="button"
          onClick={() => onSelect(plan)}
          className="group/btn flex w-full items-center justify-center gap-2 rounded-xl bg-gradient-to-r from-primary to-cyan-500 px-4 py-3 font-semibold text-primary-foreground shadow-lg transition-all duration-300 hover:shadow-xl hover:shadow-primary/25"
        >
          Join Plan
          <ArrowRight className="h-4 w-4 transition-transform group-hover/btn:translate-x-1" />
        </button>
      </CardFooter>
    </Card>
  );
}