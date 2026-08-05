"use client";

import Link from "next/link";
import { motion, useReducedMotion } from "framer-motion";
import {
  ArrowDownToLine,
  ArrowLeftRight,
  ArrowUpFromLine,
  LineChart,
  PiggyBank,
  Share2,
  ShoppingCart,
  Store,
} from "lucide-react";
import { SectionShell } from "./section-shell";

const ACTIONS = [
  { href: "/app/wallet", label: "Deposit", icon: ArrowDownToLine },
  { href: "/app/wallet", label: "Withdraw", icon: ArrowUpFromLine },
  { href: "/app/earn", label: "Invest", icon: PiggyBank },
  { href: "/app/trade", label: "Trade", icon: LineChart },
  { href: "/app/wallet", label: "Transfer", icon: ArrowLeftRight },
  { href: "/app/markets", label: "Buy Crypto", icon: ShoppingCart },
  { href: "/app/markets", label: "Sell Crypto", icon: Store },
  { href: "/app/referrals", label: "Refer Friends", icon: Share2 },
] as const;

export function QuickActionsGrid() {
  const reduceMotion = useReducedMotion();

  return (
    <SectionShell
      id="quick-actions"
      title="Quick Actions"
      description="Jump into the modules available today"
    >
      <div className="grid grid-cols-4 gap-2 sm:grid-cols-4 md:grid-cols-8">
        {ACTIONS.map(({ href, label, icon: Icon }, index) => (
          <motion.div
            key={label}
            initial={reduceMotion ? false : { opacity: 0, scale: 0.96 }}
            whileInView={reduceMotion ? undefined : { opacity: 1, scale: 1 }}
            viewport={{ once: true }}
            transition={{ delay: index * 0.03, duration: 0.25 }}
          >
            <Link
              href={href}
              aria-label={label}
              className="group flex h-full flex-col items-center gap-2 rounded-[1.15rem] border border-border/60 bg-card/70 px-2 py-3.5 text-center shadow-sm backdrop-blur-sm transition-colors hover:border-primary/40 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            >
              <motion.span
                whileHover={reduceMotion ? undefined : { scale: 1.08 }}
                whileTap={reduceMotion ? undefined : { scale: 0.94 }}
                className="rounded-xl bg-primary/10 p-2.5 text-primary transition-colors group-hover:bg-primary/15"
              >
                <Icon className="h-5 w-5" aria-hidden />
              </motion.span>
              <span className="text-[11px] font-medium leading-tight text-foreground sm:text-xs">
                {label}
              </span>
            </Link>
          </motion.div>
        ))}
      </div>
    </SectionShell>
  );
}
