"use client";

import Link from "next/link";
import { motion, useReducedMotion } from "framer-motion";
import { ArrowRight } from "lucide-react";
import { Badge } from "@coindistro/cds";
import { SectionShell, GlassCard } from "./section-shell";

export interface ForYouCard {
  id: string;
  title: string;
  description: string;
  href: string;
  cta: string;
  comingSoon?: boolean;
  tone?: "primary" | "amber" | "cyan" | "emerald";
}

const DEFAULT_CARDS: ForYouCard[] = [
  {
    id: "upgrade-tier",
    title: "Upgrade Investment Tier",
    description: "Move from Genesis to Starter for 21% ROI over 20 business days.",
    href: "/app/earn",
    cta: "View plans",
    tone: "primary",
  },
  {
    id: "complete-kyc",
    title: "Complete KYC",
    description: "Unlock higher limits and full CoinDistro banking features.",
    href: "/app/profile",
    cta: "Start verification",
    tone: "amber",
  },
  {
    id: "refer-friends",
    title: "Refer Friends",
    description: "Share your code and earn referral bonuses on investments.",
    href: "/app/referrals",
    cta: "Invite now",
    tone: "cyan",
  },
  {
    id: "trade-btc",
    title: "Trade BTC",
    description: "Spot trading is on the roadmap. Watch markets while you wait.",
    href: "/app/trade",
    cta: "Open trade",
    comingSoon: true,
    tone: "emerald",
  },
  {
    id: "buy-cdt",
    title: "Buy CDT",
    description: "Acquire CoinDistro Token when the public market listing opens.",
    href: "/app/markets",
    cta: "View markets",
    comingSoon: true,
  },
  {
    id: "stake-cdt",
    title: "Stake CDT",
    description: "Earn staking rewards once CDT staking goes live.",
    href: "/app/earn",
    cta: "Learn more",
    comingSoon: true,
  },
  {
    id: "usd-wallet",
    title: "Open USD Wallet",
    description: "Multi-currency wallets power deposits, withdrawals, and transfers.",
    href: "/app/wallet",
    cta: "Open wallet",
  },
  {
    id: "virtual-card",
    title: "Create Virtual Card",
    description: "CoinDistro Bank virtual cards for everyday spending.",
    href: "/app/settings",
    cta: "Coming soon",
    comingSoon: true,
  },
];

const toneBorder: Record<NonNullable<ForYouCard["tone"]>, string> = {
  primary: "hover:border-primary/40",
  amber: "hover:border-amber-500/40",
  cyan: "hover:border-cyan-500/40",
  emerald: "hover:border-emerald-500/40",
};

export function ForYouSection({ cards = DEFAULT_CARDS }: { cards?: ForYouCard[] }) {
  const reduceMotion = useReducedMotion();

  return (
    <SectionShell
      id="for-you"
      title="For You"
      description="Personalized opportunities on CoinDistro today"
    >
      <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
        {cards.map((card, index) => (
          <motion.div
            key={card.id}
            initial={reduceMotion ? false : { opacity: 0, y: 10 }}
            whileInView={reduceMotion ? undefined : { opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ delay: index * 0.03, duration: 0.28 }}
          >
            <GlassCard
              className={`flex h-full flex-col transition-colors ${
                card.tone ? toneBorder[card.tone] : "hover:border-primary/30"
              }`}
            >
              <div className="flex items-start justify-between gap-2">
                <h3 className="text-sm font-semibold text-foreground">{card.title}</h3>
                {card.comingSoon ? (
                  <Badge variant="outline" className="shrink-0 text-[10px]">
                    Coming Soon
                  </Badge>
                ) : null}
              </div>
              <p className="mt-2 flex-1 text-xs leading-relaxed text-muted-foreground">
                {card.description}
              </p>
              <Link
                href={card.href}
                className="mt-4 inline-flex items-center gap-1 text-xs font-semibold text-primary transition-colors hover:text-primary/80 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              >
                {card.cta}
                <ArrowRight className="h-3.5 w-3.5" aria-hidden />
              </Link>
            </GlassCard>
          </motion.div>
        ))}
      </div>
    </SectionShell>
  );
}
