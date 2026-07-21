"use client";

import * as React from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { ChevronRight, Home } from "lucide-react";
import { cn } from "@coindistro/cds";

const routeLabels: Record<string, string> = {
  app: "Dashboard",
  dashboard: "Dashboard",
  markets: "Markets",
  trade: "Trade",
  p2p: "P2P",
  earn: "Earn",
  academy: "Academy",
  signals: "Signals",
  "ai-bots": "AI Bots",
  wallet: "Wallet",
  merchant: "Merchant",
  pay: "Pay",
  referrals: "Referrals",
  notifications: "Notifications",
  profile: "Profile",
  settings: "Settings",
  admin: "Admin",
  users: "Users",
  genesis: "Genesis Members",
  invitations: "Invitations",
  "feature-flags": "Feature Flags",
  workers: "Workers",
  scheduler: "Scheduler",
  metrics: "Metrics",
  health: "System Health",
  audit: "Audit Logs",
  wallets: "Wallets",
  merchants: "Merchants",
};

function segmentLabel(segment: string): string {
  const decoded = decodeURIComponent(segment);
  return routeLabels[decoded] || routeLabels[decoded.toLowerCase()] || decoded.charAt(0).toUpperCase() + decoded.slice(1);
}

export function Breadcrumbs({ className }: { className?: string }) {
  const pathname = usePathname();

  const segments = React.useMemo(() => {
    const parts = pathname.split("/").filter(Boolean);
    return parts.map((segment, index) => {
      const href = "/" + parts.slice(0, index + 1).join("/");
      return {
        label: segmentLabel(segment),
        href,
        isLast: index === parts.length - 1,
      };
    });
  }, [pathname]);

  if (segments.length <= 1) return null;

  return (
    <nav aria-label="Breadcrumb" className={cn("flex items-center gap-1 text-sm text-muted-foreground", className)}>
      <Link href="/" className="hover:text-foreground transition-colors" aria-label="Home">
        <Home className="h-3.5 w-3.5" />
      </Link>
      {segments.map((s) => (
        <React.Fragment key={s.href}>
          <ChevronRight className="h-3.5 w-3.5 shrink-0" aria-hidden="true" />
          {s.isLast ? (
            <span className="font-medium text-foreground" aria-current="page">
              {s.label}
            </span>
          ) : (
            <Link href={s.href} className="hover:text-foreground transition-colors">
              {s.label}
            </Link>
          )}
        </React.Fragment>
      ))}
    </nav>
  );
}