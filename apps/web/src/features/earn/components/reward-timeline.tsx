"use client";

import { motion } from "framer-motion";
import { CheckCircle2, Circle, Clock3 } from "lucide-react";
import { formatCurrency } from "@/features/earn/utils";

interface RewardTimelineProps {
  days: number;
  rewardAmount: number;
  completedDays: number;
}

export function RewardTimeline({ days, rewardAmount, completedDays }: RewardTimelineProps) {
  return (
    <div className="relative space-y-1">
      {Array.from({ length: Math.min(days, 20) }, (_, index) => {
        const day = index + 1;
        const paid = day <= completedDays;
        const current = day === completedDays + 1;
        return (
          <motion.div
            key={day}
            initial={{ opacity: 0, x: -12 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ delay: day * 0.03, duration: 0.3 }}
            className={`flex items-center gap-3 rounded-xl border px-4 py-2.5 ${
              paid
                ? "border-emerald-500/30 bg-emerald-500/5"
                : current
                  ? "border-primary/40 bg-primary/10"
                  : "border-border/60 bg-muted/20"
            }`}
          >
            <span className="flex h-8 w-8 items-center justify-center">
              {paid ? (
                <CheckCircle2 className="h-5 w-5 text-emerald-500" />
              ) : current ? (
                <Clock3 className="h-5 w-5 text-primary" />
              ) : (
                <Circle className="h-4 w-4 text-muted-foreground/40" />
              )}
            </span>
            <div className="flex-1">
              <p className="text-sm font-medium text-foreground">Day {day}</p>
              <p className="text-xs text-muted-foreground">
                {paid ? "Paid" : current ? "In progress" : "Pending"}
              </p>
            </div>
            <span className={`text-sm font-semibold tabular-nums ${paid ? "text-emerald-500" : "text-muted-foreground"}`}>
              {formatCurrency(rewardAmount)}
            </span>
          </motion.div>
        );
      })}
    </div>
  );
}