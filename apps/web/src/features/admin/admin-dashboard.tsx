"use client";

import * as React from "react";
import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import {
  Badge,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  PageHeader,
  Skeleton,
  StatCard,
  StatusDot,
} from "@coindistro/cds";
import {
  Activity,
  Database,
  Flag,
  HeartPulse,
  Server,
  Sparkles,
  Users,
} from "lucide-react";
import * as adminApi from "@/features/admin/api";
import { formatRelative, humanizeAction } from "@/lib/utils/format";

function healthDot(status?: string): "online" | "error" | "busy" | "offline" {
  const s = (status || "").toLowerCase();
  if (s === "healthy" || s === "running" || s === "ready") return "online";
  if (s === "degraded" || s === "busy") return "busy";
  if (s === "unhealthy" || s === "error") return "error";
  return "offline";
}

export function AdminDashboard() {
  const systemQ = useQuery({
    queryKey: ["admin", "system"],
    queryFn: adminApi.getAdminSystem,
    refetchInterval: 30_000,
  });
  const statsQ = useQuery({
    queryKey: ["admin", "stats"],
    queryFn: adminApi.getPlatformStats,
    refetchInterval: 30_000,
  });
  const healthQ = useQuery({
    queryKey: ["health"],
    queryFn: adminApi.getHealth,
    refetchInterval: 30_000,
  });

  const system = systemQ.data;
  const stats = statsQ.data;
  const health = healthQ.data;
  const loading = systemQ.isLoading || statsQ.isLoading;

  const workerStatus = String(
    (system?.workers as { status?: string } | undefined)?.status ?? "disabled",
  );
  const schedulerStatus = String(
    (system?.scheduler as { status?: string } | undefined)?.status ?? "disabled",
  );
  const flags = system?.feature_flags ?? [];
  const enabledFlags = flags.filter((f) => f.enabled).length;

  return (
    <div className="space-y-6 animate-cds-fade-in">
      <PageHeader
        title="Admin overview"
        description="Live platform control plane — users, health, and operations."
        actions={
          <div className="flex flex-wrap items-center gap-2">
            <Badge variant={system?.api_status === "healthy" ? "success" : "warning"}>
              API {system?.api_status ?? "…"}
            </Badge>
            <Badge variant="secondary">{system?.environment ?? "…"}</Badge>
            <Badge variant="outline">v{system?.version ?? health?.version ?? "—"}</Badge>
          </div>
        }
      />

      {/* Primary KPIs */}
      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        {loading ? (
          Array.from({ length: 4 }).map((_, i) => (
            <Card key={i}>
              <CardContent className="space-y-3 p-6">
                <Skeleton className="h-4 w-24" />
                <Skeleton className="h-8 w-16" />
              </CardContent>
            </Card>
          ))
        ) : (
          <>
            <StatCard
              title="Total users"
              value={stats?.total_users ?? 0}
              description={`${stats?.active_users ?? 0} active`}
              icon={<Users className="h-4 w-4" />}
            />
            <StatCard
              title="Verified users"
              value={stats?.verified_users ?? 0}
              description="Email verified"
              icon={<Activity className="h-4 w-4" />}
            />
            <StatCard
              title="Genesis members"
              value={stats?.genesis_members ?? 0}
              description={
                stats?.genesis_config
                  ? `${stats.genesis_config.current_genesis_count}/${stats.genesis_config.max_genesis_members} slots`
                  : "Genesis program"
              }
              icon={<Sparkles className="h-4 w-4" />}
            />
            <StatCard
              title="System status"
              value={system?.status ?? health?.status ?? "unknown"}
              description={`${system?.environment ?? ""} · v${system?.version ?? ""}`}
              icon={<HeartPulse className="h-4 w-4" />}
            />
          </>
        )}
      </div>

      {/* Infrastructure */}
      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        {[
          { label: "API status", value: system?.api_status, icon: Server },
          { label: "Database", value: system?.database ?? health?.checks?.database, icon: Database },
          { label: "Redis", value: system?.redis ?? health?.checks?.redis, icon: Database },
          { label: "Backend", value: system?.backend ?? "healthy", icon: Server },
        ].map((item) => (
          <Card key={item.label}>
            <CardContent className="flex items-center gap-3 p-4">
              <item.icon className="h-5 w-5 text-muted-foreground" />
              <div className="min-w-0 flex-1">
                <p className="text-xs text-muted-foreground">{item.label}</p>
                <p className="truncate text-sm font-medium capitalize">{item.value ?? "—"}</p>
              </div>
              <StatusDot status={healthDot(item.value)} />
            </CardContent>
          </Card>
        ))}
      </div>

      <div className="grid gap-4 lg:grid-cols-3">
        {/* Feature flags + workers */}
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Feature flags</CardTitle>
            <CardDescription>
              {enabledFlags} enabled of {flags.length}
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-2">
            {systemQ.isLoading ? (
              <Skeleton className="h-24 w-full" />
            ) : flags.length === 0 ? (
              <p className="text-sm text-muted-foreground">No flags loaded.</p>
            ) : (
              <ul className="max-h-48 space-y-2 overflow-auto text-sm">
                {flags.slice(0, 8).map((f) => (
                  <li key={f.name} className="flex items-center justify-between gap-2">
                    <span className="truncate font-mono text-xs">{f.name}</span>
                    <Badge variant={f.enabled ? "success" : "secondary"}>
                      {f.enabled ? "On" : "Off"}
                    </Badge>
                  </li>
                ))}
              </ul>
            )}
            <Link
              href="/admin/feature-flags"
              className="inline-flex items-center gap-1 text-sm text-primary hover:underline"
            >
              <Flag className="h-3.5 w-3.5" /> Manage flags
            </Link>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-base">Workers & scheduler</CardTitle>
            <CardDescription>Background job health</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4 text-sm">
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">Workers</span>
              <span className="inline-flex items-center gap-2 font-medium capitalize">
                <StatusDot status={healthDot(workerStatus)} />
                {workerStatus}
              </span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">Scheduler</span>
              <span className="inline-flex items-center gap-2 font-medium capitalize">
                <StatusDot status={healthDot(schedulerStatus)} />
                {schedulerStatus}
              </span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">Docker</span>
              <span className="font-medium capitalize">{system?.docker ?? "unknown"}</span>
            </div>
            <div className="flex gap-2 pt-1">
              <Link href="/admin/workers" className="text-primary hover:underline">
                Workers
              </Link>
              <span className="text-muted-foreground">·</span>
              <Link href="/admin/scheduler" className="text-primary hover:underline">
                Scheduler
              </Link>
              <span className="text-muted-foreground">·</span>
              <Link href="/admin/health" className="text-primary hover:underline">
                Health
              </Link>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-base">Earn & referrals</CardTitle>
            <CardDescription>Platform growth metrics</CardDescription>
          </CardHeader>
          <CardContent className="grid grid-cols-2 gap-3 text-sm">
            <div>
              <p className="text-xs text-muted-foreground">Total referrals</p>
              <p className="text-xl font-bold">{stats?.total_referrals ?? 0}</p>
            </div>
            <div>
              <p className="text-xs text-muted-foreground">Invitations</p>
              <p className="text-xl font-bold">{stats?.total_invitations ?? 0}</p>
            </div>
            <div className="col-span-2 text-xs text-muted-foreground">
              Earn product analytics are available under Admin → Earn when the module is enabled.
            </div>
            <Link href="/admin/referrals" className="text-sm text-primary hover:underline">
              Referral admin
            </Link>
            <Link href="/admin/earn" className="text-sm text-primary hover:underline">
              Earn admin
            </Link>
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-4 lg:grid-cols-2">
        {/* Recent registrations */}
        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <div>
              <CardTitle className="text-base">Recent registrations</CardTitle>
              <CardDescription>Newest accounts</CardDescription>
            </div>
            <Link href="/admin/users" className="text-sm text-primary hover:underline">
              All users
            </Link>
          </CardHeader>
          <CardContent>
            {statsQ.isLoading ? (
              <Skeleton className="h-40 w-full" />
            ) : !stats?.recent_registrations?.length ? (
              <p className="text-sm text-muted-foreground">No registrations yet.</p>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-left text-sm">
                  <thead className="text-xs text-muted-foreground">
                    <tr className="border-b">
                      <th className="pb-2 font-medium">User</th>
                      <th className="pb-2 font-medium">Status</th>
                      <th className="pb-2 font-medium">Joined</th>
                    </tr>
                  </thead>
                  <tbody>
                    {stats.recent_registrations.map((u) => (
                      <tr key={u.id} className="border-b border-border/50 last:border-0">
                        <td className="py-2">
                          <div className="font-medium">{u.display_name || u.username || u.email}</div>
                          <div className="text-xs text-muted-foreground">{u.email}</div>
                        </td>
                        <td className="py-2">
                          <Badge variant={u.is_verified ? "success" : "secondary"} className="text-[10px]">
                            {u.status}
                          </Badge>
                        </td>
                        <td className="py-2 text-muted-foreground">{formatRelative(u.created_at)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </CardContent>
        </Card>

        {/* Recent logins */}
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Recent logins</CardTitle>
            <CardDescription>Latest successful authentications</CardDescription>
          </CardHeader>
          <CardContent>
            {statsQ.isLoading ? (
              <Skeleton className="h-40 w-full" />
            ) : !stats?.recent_logins?.length ? (
              <p className="text-sm text-muted-foreground">No logins recorded yet.</p>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-left text-sm">
                  <thead className="text-xs text-muted-foreground">
                    <tr className="border-b">
                      <th className="pb-2 font-medium">User</th>
                      <th className="pb-2 font-medium">Roles</th>
                      <th className="pb-2 font-medium">When</th>
                    </tr>
                  </thead>
                  <tbody>
                    {stats.recent_logins.map((u) => (
                      <tr key={u.id} className="border-b border-border/50 last:border-0">
                        <td className="py-2">
                          <div className="font-medium">{u.email}</div>
                        </td>
                        <td className="py-2 text-xs text-muted-foreground">
                          {(u.roles || []).join(", ") || "user"}
                        </td>
                        <td className="py-2 text-muted-foreground">
                          {formatRelative(u.last_login_at)}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Audit activity */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <div>
            <CardTitle className="text-base">Audit activity</CardTitle>
            <CardDescription>Recent security and identity events</CardDescription>
          </div>
          <Link href="/admin/audit" className="text-sm text-primary hover:underline">
            Audit logs
          </Link>
        </CardHeader>
        <CardContent>
          {!stats?.recent_activity?.length ? (
            <p className="text-sm text-muted-foreground">No audit events yet.</p>
          ) : (
            <ul className="divide-y divide-border/60">
              {stats.recent_activity.slice(0, 10).map((a) => (
                <li key={a.id} className="flex flex-wrap items-center justify-between gap-2 py-2 text-sm">
                  <span className="font-medium">{humanizeAction(a.action)}</span>
                  <span className="text-xs text-muted-foreground">
                    {a.ip_address ? `${a.ip_address} · ` : ""}
                    {formatRelative(a.created_at)}
                  </span>
                </li>
              ))}
            </ul>
          )}
        </CardContent>
      </Card>

      {/* Meta footer */}
      <div className="flex flex-wrap gap-3 text-xs text-muted-foreground">
        <span>Environment: {system?.environment ?? "—"}</span>
        <span>·</span>
        <span>Version: {system?.version ?? "—"}</span>
        <span>·</span>
        <span>Updated: {formatRelative(system?.timestamp || health?.timestamp)}</span>
      </div>
    </div>
  );
}
