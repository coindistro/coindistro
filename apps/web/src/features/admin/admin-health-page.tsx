"use client";

import { useQuery } from "@tanstack/react-query";
import {
  Badge,
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  PageHeader,
  Skeleton,
  StatCard,
  StatusDot,
} from "@coindistro/cds";
import * as adminApi from "@/features/admin/api";
import { formatDate } from "@/lib/utils/format";

function dot(status?: string): "online" | "error" | "busy" | "offline" {
  const s = (status || "").toLowerCase();
  if (s.includes("healthy") || s === "running" || s === "ready") return "online";
  if (s.includes("degraded")) return "busy";
  if (s.includes("unhealthy") || s.includes("error")) return "error";
  return "offline";
}

export function AdminHealthPage() {
  const healthQ = useQuery({
    queryKey: ["health"],
    queryFn: adminApi.getHealth,
    refetchInterval: 15_000,
  });
  const systemQ = useQuery({
    queryKey: ["admin", "system"],
    queryFn: adminApi.getAdminSystem,
    refetchInterval: 15_000,
  });
  const workersQ = useQuery({
    queryKey: ["admin", "workers"],
    queryFn: adminApi.getWorkersStatus,
  });
  const schedQ = useQuery({
    queryKey: ["admin", "scheduler"],
    queryFn: adminApi.getSchedulerStatus,
  });

  const health = healthQ.data;
  const system = systemQ.data;

  return (
    <div className="space-y-6 animate-cds-fade-in">
      <PageHeader
        title="System health"
        description="Service health, readiness, and infrastructure dependencies."
        actions={
          <Badge variant={health?.status === "healthy" ? "success" : "warning"}>
            {health?.status ?? "…"}
          </Badge>
        }
      />

      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <StatCard title="Overall" value={health?.status ?? system?.status ?? "—"} />
        <StatCard title="Version" value={system?.version ?? health?.version ?? "—"} />
        <StatCard title="Environment" value={system?.environment ?? "—"} />
        <StatCard
          title="Checked"
          value={formatDate(health?.timestamp || system?.timestamp)}
        />
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Dependency checks</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {healthQ.isLoading ? (
              <Skeleton className="h-24 w-full" />
            ) : (
              Object.entries({
                database: system?.database ?? health?.checks?.database,
                redis: system?.redis ?? health?.checks?.redis,
                server: health?.checks?.server ?? "healthy",
                api: system?.api_status,
                backend: system?.backend,
                docker: system?.docker,
              }).map(([key, value]) => (
                <div key={key} className="flex items-center justify-between text-sm">
                  <span className="capitalize text-muted-foreground">{key}</span>
                  <span className="inline-flex items-center gap-2 font-medium">
                    <StatusDot status={dot(String(value))} />
                    {String(value ?? "—")}
                  </span>
                </div>
              ))
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-base">Workers & scheduler</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4 text-sm">
            {workersQ.isLoading || schedQ.isLoading ? (
              <Skeleton className="h-24 w-full" />
            ) : (
              <>
                <pre className="overflow-auto rounded-md bg-muted/40 p-3 text-xs">
                  {JSON.stringify(workersQ.data, null, 2)}
                </pre>
                <pre className="overflow-auto rounded-md bg-muted/40 p-3 text-xs">
                  {JSON.stringify(schedQ.data, null, 2)}
                </pre>
              </>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
