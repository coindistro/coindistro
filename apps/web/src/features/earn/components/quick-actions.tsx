"use client";

import { motion } from "framer-motion";
import { PiggyBank, Wallet, Sparkles, Users } from "lucide-react";

interface QuickActionsProps {
  onInvest: () => void;
  onWithdraw: () => void;
  onRewards: () => void;
  onReferrals: () => void;
}

const actions = [
  { id: "invest", label: "Invest", icon: PiggyBank },
  { id: "withdraw", label: "Withdraw", icon: Wallet },
  { id: "rewards", label: "Rewards", icon: Sparkles },
  { id: "referrals", label: "Referrals", icon: Users },
] as const;

export function QuickActions({ onInvest, onWithdraw, onRewards, onReferrals }: QuickActionsProps) {
  const handlers: Record<string, () => void> = {
    invest: onInvest,
    withdraw: onWithdraw,
    rewards: onRewards,
    referrals: onReferrals,
  };

  return (
    <div className="grid grid-cols-4 gap-3">
      {actions.map(({ id, label, icon: Icon }) => (
        <motion.button
          key={id}
          type="button"
          whileHover={{ scale: 1.03 }}
          whileTap={{ scale: 0.96 }}
          onClick={handlers[id]}
          className="group flex flex-col items-center gap-1.5 rounded-2xl border border-border bg-card/80 px-2 py-3.5 backdrop-blur-sm transition-colors hover:border-primary/40"
          aria-label={label}
        >
          <span className="rounded-xl bg-primary/10 p-2.5 text-primary transition-transform group-hover:scale-110">
            <Icon className="h-5 w-5" />
          </span>
          <span className="text-xs font-medium text-foreground">{label}</span>
        </motion.button>
      ))}
    </div>
  );
}