"use client";

import { useQuery } from "@tanstack/react-query";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  PageHeader,
  Skeleton,
  StatCard,
} from "@coindistro/cds";
import * as adminApi from "@/features/admin/api";

export function AdminWorkersPage() {
  const workersQ = useQuery({
    queryKey: ["admin", "workers"],
    queryFn: adminApi.getWorkersStatus,
    refetchInterval: 15_000,
  });

  const data = workersQ.data ?? {};

  return (
    <div className="space-y-6 animate-cds-fade-in">
      <PageHeader title="Workers" description="Background worker pool status." />
      <div className="grid gap-4 sm:grid-cols-3">
        <StatCard title="Status" value={String(data.status ?? "—")} />
        <StatCard title="Workers" value={String(data.num_workers ?? "—")} />
        <StatCard
          title="Queue"
          value={
            data.queue_len != null
              ? `${data.queue_len}/${data.queue_cap ?? "?"}`
              : "—"
          }
        />
      </div>
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Raw status</CardTitle>
        </CardHeader>
        <CardContent>
          {workersQ.isLoading ? (
            <Skeleton className="h-32 w-full" />
          ) : (
            <pre className="overflow-auto rounded-md bg-muted/40 p-4 text-xs">
              {JSON.stringify(data, null, 2)}
            </pre>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
