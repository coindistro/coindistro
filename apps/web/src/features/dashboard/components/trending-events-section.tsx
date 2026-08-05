"use client";

import * as React from "react";
import Link from "next/link";
import { motion, useReducedMotion, AnimatePresence } from "framer-motion";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { SectionShell, GlassCard } from "./section-shell";

export interface TrendingEvent {
  id: string;
  title: string;
  description: string;
  href: string;
  cta: string;
  accent: string;
}

const DEFAULT_EVENTS: TrendingEvent[] = [
  {
    id: "genesis",
    title: "Genesis Investment",
    description: "Start from $10 with 18% ROI over 20 business days.",
    href: "/app/earn",
    cta: "Invest now",
    accent: "from-violet-500/30 to-fuchsia-500/10",
  },
  {
    id: "referral",
    title: "Referral Campaign",
    description: "Invite friends and earn on successful investments.",
    href: "/app/referrals",
    cta: "Share link",
    accent: "from-cyan-500/30 to-blue-500/10",
  },
  {
    id: "trade-comp",
    title: "Trade Competition",
    description: "Compete when spot trading launches on CoinDistro.",
    href: "/app/trade",
    cta: "Get ready",
    accent: "from-amber-500/25 to-orange-500/10",
  },
  {
    id: "cdt-launch",
    title: "CDT Launch",
    description: "CoinDistro Token listing and ecosystem utilities ahead.",
    href: "/app/markets",
    cta: "Learn more",
    accent: "from-emerald-500/25 to-teal-500/10",
  },
  {
    id: "cashback",
    title: "Cashback Week",
    description: "Rewards and cashback moments across the ecosystem.",
    href: "/app/earn",
    cta: "View rewards",
    accent: "from-pink-500/25 to-rose-500/10",
  },
  {
    id: "quiz",
    title: "Crypto Quiz",
    description: "Learn, play, and earn Academy rewards.",
    href: "/app/academy",
    cta: "Open Academy",
    accent: "from-indigo-500/25 to-purple-500/10",
  },
];

export function TrendingEventsSection({ events = DEFAULT_EVENTS }: { events?: TrendingEvent[] }) {
  const reduceMotion = useReducedMotion();
  const [index, setIndex] = React.useState(0);
  const count = events.length;

  React.useEffect(() => {
    if (reduceMotion || count <= 1) return;
    const id = window.setInterval(() => {
      setIndex((prev) => (prev + 1) % count);
    }, 5000);
    return () => window.clearInterval(id);
  }, [count, reduceMotion]);

  const visible = [
    events[index % count],
    events[(index + 1) % count],
    events[(index + 2) % count],
  ].filter(Boolean);

  return (
    <SectionShell
      id="trending-events"
      title="Trending Events"
      description="Campaigns and opportunities across the platform"
    >
      <div className="relative">
        <div className="mb-3 flex justify-end gap-2">
          <button
            type="button"
            aria-label="Previous event"
            onClick={() => setIndex((prev) => (prev - 1 + count) % count)}
            className="flex h-8 w-8 items-center justify-center rounded-xl border border-border/60 bg-card/80 text-muted-foreground transition-colors hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          >
            <ChevronLeft className="h-4 w-4" />
          </button>
          <button
            type="button"
            aria-label="Next event"
            onClick={() => setIndex((prev) => (prev + 1) % count)}
            className="flex h-8 w-8 items-center justify-center rounded-xl border border-border/60 bg-card/80 text-muted-foreground transition-colors hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          >
            <ChevronRight className="h-4 w-4" />
          </button>
        </div>

        <div className="grid gap-3 md:grid-cols-3">
          <AnimatePresence mode="popLayout" initial={false}>
            {visible.map((event) => (
              <motion.div
                key={`${event.id}-${index}`}
                initial={reduceMotion ? false : { opacity: 0, x: 16 }}
                animate={reduceMotion ? undefined : { opacity: 1, x: 0 }}
                exit={reduceMotion ? undefined : { opacity: 0, x: -16 }}
                transition={{ duration: 0.3 }}
              >
                <GlassCard
                  className={`relative overflow-hidden bg-gradient-to-br ${event.accent} p-5`}
                >
                  <div className="pointer-events-none absolute -right-8 -top-8 h-24 w-24 rounded-full bg-white/5 blur-2xl" />
                  <p className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                    Event
                  </p>
                  <h3 className="mt-2 text-base font-bold text-foreground">{event.title}</h3>
                  <p className="mt-2 text-sm text-muted-foreground">{event.description}</p>
                  <Link
                    href={event.href}
                    className="mt-4 inline-flex text-sm font-semibold text-primary focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                  >
                    {event.cta} →
                  </Link>
                </GlassCard>
              </motion.div>
            ))}
          </AnimatePresence>
        </div>
      </div>
    </SectionShell>
  );
}
