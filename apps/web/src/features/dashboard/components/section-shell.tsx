"use client";

import * as React from "react";
import Link from "next/link";
import { motion, useReducedMotion } from "framer-motion";
import { ArrowRight } from "lucide-react";
import { cn } from "@coindistro/cds";

export function SectionShell({
  title,
  description,
  actionHref,
  actionLabel,
  children,
  className,
  id,
}: {
  title: string;
  description?: string;
  actionHref?: string;
  actionLabel?: string;
  children: React.ReactNode;
  className?: string;
  id?: string;
}) {
  const reduceMotion = useReducedMotion();

  return (
    <motion.section
      id={id}
      initial={reduceMotion ? false : { opacity: 0, y: 14 }}
      whileInView={reduceMotion ? undefined : { opacity: 1, y: 0 }}
      viewport={{ once: true, margin: "-40px" }}
      transition={{ duration: 0.35, ease: "easeOut" }}
      className={cn("space-y-3", className)}
      aria-labelledby={id ? `${id}-title` : undefined}
    >
      <div className="flex items-start justify-between gap-3">
        <div>
          <h2
            id={id ? `${id}-title` : undefined}
            className="text-lg font-bold tracking-tight text-foreground"
          >
            {title}
          </h2>
          {description ? (
            <p className="mt-0.5 text-sm text-muted-foreground">{description}</p>
          ) : null}
        </div>
        {actionHref && actionLabel ? (
          <Link
            href={actionHref}
            className="inline-flex items-center gap-1 rounded-lg px-2 py-1 text-xs font-semibold text-primary transition-colors hover:bg-primary/10 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          >
            {actionLabel}
            <ArrowRight className="h-3.5 w-3.5" aria-hidden />
          </Link>
        ) : null}
      </div>
      {children}
    </motion.section>
  );
}

export function GlassCard({
  children,
  className,
  as: Comp = "div",
}: {
  children: React.ReactNode;
  className?: string;
  as?: "div" | "article" | "li";
}) {
  return (
    <Comp
      className={cn(
        "rounded-[1.25rem] border border-border/60 bg-card/80 p-4 shadow-[0_8px_30px_rgb(0,0,0,0.12)] backdrop-blur-md",
        className,
      )}
    >
      {children}
    </Comp>
  );
}

export function SectionSkeleton({ rows = 3 }: { rows?: number }) {
  return (
    <div className="space-y-3" aria-hidden>
      {Array.from({ length: rows }).map((_, i) => (
        <div
          key={i}
          className="h-20 animate-pulse rounded-[1.25rem] border border-border/40 bg-muted/40"
        />
      ))}
    </div>
  );
}
