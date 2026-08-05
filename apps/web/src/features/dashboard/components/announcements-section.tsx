"use client";

import Link from "next/link";
import { Badge } from "@coindistro/cds";
import { formatRelative } from "@/lib/utils/format";
import { SectionShell, GlassCard, SectionSkeleton } from "./section-shell";

export interface AnnouncementItem {
  id: string;
  title: string;
  category: string;
  created_at: string;
  href?: string;
}

const FALLBACK_ANNOUNCEMENTS: AnnouncementItem[] = [
  {
    id: "a1",
    title: "Genesis Investment Plans are live — start from $10",
    category: "Investment",
    created_at: new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString(),
    href: "/app/earn",
  },
  {
    id: "a2",
    title: "Weekly withdrawals process within 24 hours",
    category: "Platform",
    created_at: new Date(Date.now() - 8 * 60 * 60 * 1000).toISOString(),
    href: "/app/earn",
  },
  {
    id: "a3",
    title: "Scheduled maintenance window this weekend",
    category: "System",
    created_at: new Date(Date.now() - 26 * 60 * 60 * 1000).toISOString(),
  },
  {
    id: "a4",
    title: "Referral rewards update for Genesis investors",
    category: "Referral",
    created_at: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000).toISOString(),
    href: "/app/referrals",
  },
  {
    id: "a5",
    title: "Markets module and CDT listing on the roadmap",
    category: "Listings",
    created_at: new Date(Date.now() - 3 * 24 * 60 * 60 * 1000).toISOString(),
    href: "/app/markets",
  },
];

export function AnnouncementsSection({
  items,
  loading,
}: {
  items?: AnnouncementItem[];
  loading?: boolean;
}) {
  const list = (items?.length ? items : FALLBACK_ANNOUNCEMENTS).slice(0, 5);

  return (
    <SectionShell
      id="announcements"
      title="Announcements"
      description="Latest platform updates"
      actionHref="/app/notifications"
      actionLabel="View All"
    >
      {loading ? (
        <SectionSkeleton rows={3} />
      ) : (
        <GlassCard className="divide-y divide-border/50 p-0">
          <ul>
            {list.map((item) => (
              <li key={item.id}>
                <Link
                  href={item.href || "/app/notifications"}
                  className="flex items-start justify-between gap-3 px-4 py-3.5 transition-colors hover:bg-muted/30 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-ring"
                >
                  <div className="min-w-0">
                    <div className="flex flex-wrap items-center gap-2">
                      <Badge variant="secondary" className="text-[10px]">
                        {item.category}
                      </Badge>
                      <span className="text-[11px] text-muted-foreground">
                        {formatRelative(item.created_at)}
                      </span>
                    </div>
                    <p className="mt-1.5 text-sm font-medium text-foreground">{item.title}</p>
                  </div>
                </Link>
              </li>
            ))}
          </ul>
        </GlassCard>
      )}
    </SectionShell>
  );
}
