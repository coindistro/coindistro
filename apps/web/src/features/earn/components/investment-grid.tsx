"use client";

import { InvestmentCard } from "@/features/earn/components/investment-card";
import type { InvestmentPlanConfig } from "@/features/earn/config/investment-plans";

interface InvestmentGridProps {
  plans: readonly InvestmentPlanConfig[];
  exchangeRate: number;
  onSelectPlan: (plan: InvestmentPlanConfig) => void;
}

export function InvestmentGrid({ plans, exchangeRate, onSelectPlan }: InvestmentGridProps) {
  return (
    <div className="grid gap-6 md:grid-cols-2 xl:grid-cols-3">
      {plans.map((plan) => (
        <InvestmentCard
          key={plan.id}
          plan={plan}
          exchangeRate={exchangeRate}
          onSelect={onSelectPlan}
        />
      ))}
    </div>
  );
}