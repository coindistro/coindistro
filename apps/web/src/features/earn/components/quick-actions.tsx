"use client";

import { motion } from "framer-motion";
import { PiggyBank, Wallet, Sparkles, Users } from "lucide-react";

interface QuickActionsProps {
  onInvest: () => void;
  onWithdraw: () => void;
  onRewards: () => void;
  onReferrals: () => void;
  /** When true, the Withdraw action is disabled (weekly cooldown). */
  withdrawDisabled?: boolean;
  /** Optional label override for the withdraw action when locked. */
  withdrawLabel?: string;
}

const actions = [
  { id: "invest", label: "Invest", icon: PiggyBank },
  { id: "withdraw", label: "Withdraw", icon: Wallet },
  { id: "rewards", label: "Rewards", icon: Sparkles },
  { id: "referrals", label: "Referrals", icon: Users },
] as const;

export function QuickActions({
  onInvest,
  onWithdraw,
  onRewards,
  onReferrals,
  withdrawDisabled = false,
  withdrawLabel,
}: QuickActionsProps) {
  const handlers: Record<string, () => void> = {
    invest: onInvest,
    withdraw: onWithdraw,
    rewards: onRewards,
    referrals: onReferrals,
  };

  return (
    <div className="grid grid-cols-4 gap-3">
      {actions.map(({ id, label, icon: Icon }) => {
        const disabled = id === "withdraw" && withdrawDisabled;
        const displayLabel = id === "withdraw" && withdrawLabel ? withdrawLabel : label;
        return (
          <motion.button
            key={id}
            type="button"
            whileHover={disabled ? undefined : { scale: 1.03 }}
            whileTap={disabled ? undefined : { scale: 0.96 }}
            onClick={disabled ? undefined : handlers[id]}
            disabled={disabled}
            className={`group flex flex-col items-center gap-1.5 rounded-2xl border border-border bg-card/80 px-2 py-3.5 backdrop-blur-sm transition-colors ${
              disabled
                ? "cursor-not-allowed opacity-50"
                : "hover:border-primary/40"
            }`}
            aria-label={displayLabel}
            aria-disabled={disabled}
          >
            <span className="rounded-xl bg-primary/10 p-2.5 text-primary transition-transform group-hover:scale-110">
              <Icon className="h-5 w-5" />
            </span>
            <span className="text-xs font-medium text-foreground">{displayLabel}</span>
          </motion.button>
        );
      })}
    </div>
  );
}
