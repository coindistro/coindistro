"use client";

import * as React from "react";
import { useQuery } from "@tanstack/react-query";
import {
  Badge,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  EmptyState,
  PageHeader,
  Skeleton,
} from "@coindistro/cds";
import { Bell, Gift, KeyRound, MonitorSmartphone, Shield, Sparkles } from "lucide-react";
import * as identityApi from "@/features/identity/api";
import { formatRelative, humanizeAction } from "@/lib/utils/format";

function iconForAction(action: string) {
  const a = action.toLowerCase();
  if (a.includes("login")) return <Shield className="h-4 w-4 text-success" />;
  if (a.includes("password")) return <KeyRound className="h-4 w-4 text-warning" />;
  if (a.includes("device")) return <MonitorSmartphone className="h-4 w-4 text-info" />;
  if (a.includes("referral") || a.includes("invite")) return <Gift className="h-4 w-4 text-primary" />;
  if (a.includes("genesis")) return <Sparkles className="h-4 w-4 text-primary" />;
  return <Bell className="h-4 w-4 text-muted-foreground" />;
}

function categoryForAction(action: string): string {
  const a = action.toLowerCase();
  if (a.includes("login")) return "Successful Login";
  if (a.includes("password")) return "Password Changes";
  if (a.includes("device")) return "New Device";
  if (a.includes("referral") || a.includes("invite")) return "Referral Joined";
  if (a.includes("genesis")) return "Genesis Granted";
  return "System";
}

const SYSTEM_NOTICES = [
  {
    id: "sys-1",
    title: "System notifications",
    body: "Platform announcements and maintenance notices will appear here.",
    category: "System Notifications",
  },
];

export function NotificationsPage() {
  const activityQ = useQuery({
    queryKey: ["activity"],
    queryFn: identityApi.getActivityLog,
  });

  const items = activityQ.data ?? [];

  return (
    <div className="space-y-6 animate-cds-fade-in">
      <PageHeader
        title="Notifications"
        description="Security, referrals, genesis, and system alerts from your account activity."
      />

      <div className="flex flex-wrap gap-2">
        {[
          "Successful Login",
          "Password Changes",
          "New Device",
          "Referral Joined",
          "Genesis Granted",
          "System Notifications",
        ].map((c) => (
          <Badge key={c} variant="secondary">
            {c}
          </Badge>
        ))}
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Inbox</CardTitle>
          <CardDescription>
            Derived from live activity events. Global toasts fire for critical auth actions.
          </CardDescription>
        </CardHeader>
        <CardContent>
          {activityQ.isLoading ? (
            <div className="space-y-3">
              <Skeleton className="h-16 w-full" />
              <Skeleton className="h-16 w-full" />
              <Skeleton className="h-16 w-full" />
            </div>
          ) : items.length === 0 ? (
            <EmptyState
              icon={<Bell className="h-8 w-8" />}
              title="No notifications yet"
              description="Successful logins, password changes, new devices, referrals, and genesis grants will show up here."
            />
          ) : (
            <ul className="divide-y divide-border/60">
              {items.map((a) => (
                <li key={a.id} className="flex gap-3 py-3">
                  <div className="mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-muted">
                    {iconForAction(a.action)}
                  </div>
                  <div className="min-w-0 flex-1">
                    <div className="flex flex-wrap items-center gap-2">
                      <p className="font-medium">{humanizeAction(a.action)}</p>
                      <Badge variant="outline" className="text-[10px]">
                        {categoryForAction(a.action)}
                      </Badge>
                    </div>
                    <p className="text-xs text-muted-foreground">
                      {a.ip_address ? `${a.ip_address} · ` : ""}
                      {formatRelative(a.created_at)}
                    </p>
                  </div>
                </li>
              ))}
              {SYSTEM_NOTICES.map((n) => (
                <li key={n.id} className="flex gap-3 py-3">
                  <div className="mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-muted">
                    <Bell className="h-4 w-4 text-muted-foreground" />
                  </div>
                  <div>
                    <div className="flex flex-wrap items-center gap-2">
                      <p className="font-medium">{n.title}</p>
                      <Badge variant="outline" className="text-[10px]">
                        {n.category}
                      </Badge>
                    </div>
                    <p className="text-sm text-muted-foreground">{n.body}</p>
                  </div>
                </li>
              ))}
            </ul>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
