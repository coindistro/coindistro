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
} from "@coindistro/cds";
import * as adminApi from "@/features/admin/api";
import { formatRelative } from "@/lib/utils/format";

type TaskRow = {
  task_id?: string;
  name?: string;
  run_count?: number;
  error_count?: number;
  is_running?: boolean;
  last_run_at?: string;
  next_run_at?: string;
  last_error?: string;
};

export function AdminSchedulerPage() {
  const schedQ = useQuery({
    queryKey: ["admin", "scheduler"],
    queryFn: adminApi.getSchedulerStatus,
    refetchInterval: 15_000,
  });

  const data = schedQ.data ?? {};
  const tasks = (Array.isArray(data.tasks) ? data.tasks : []) as TaskRow[];

  return (
    <div className="space-y-6 animate-cds-fade-in">
      <PageHeader title="Scheduler" description="Recurring background task status." />
      <div className="grid gap-4 sm:grid-cols-3">
        <StatCard title="Status" value={String(data.status ?? "—")} />
        <StatCard title="Enabled" value={String(data.enabled ?? "—")} />
        <StatCard title="Tasks" value={tasks.length} />
      </div>
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Scheduled tasks</CardTitle>
        </CardHeader>
        <CardContent>
          {schedQ.isLoading ? (
            <Skeleton className="h-32 w-full" />
          ) : tasks.length === 0 ? (
            <p className="text-sm text-muted-foreground">No tasks registered.</p>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-left text-sm">
                <thead className="text-xs text-muted-foreground">
                  <tr className="border-b">
                    <th className="pb-2 font-medium">Task</th>
                    <th className="pb-2 font-medium">Runs</th>
                    <th className="pb-2 font-medium">Errors</th>
                    <th className="pb-2 font-medium">Last run</th>
                    <th className="pb-2 font-medium">State</th>
                  </tr>
                </thead>
                <tbody>
                  {tasks.map((t) => (
                    <tr key={t.task_id || t.name} className="border-b border-border/50">
                      <td className="py-2">
                        <div className="font-medium">{t.name}</div>
                        <div className="font-mono text-xs text-muted-foreground">
                          {t.task_id}
                        </div>
                      </td>
                      <td className="py-2">{t.run_count ?? 0}</td>
                      <td className="py-2">{t.error_count ?? 0}</td>
                      <td className="py-2 text-muted-foreground">
                        {formatRelative(t.last_run_at)}
                      </td>
                      <td className="py-2">
                        <Badge variant={t.is_running ? "info" : "secondary"}>
                          {t.is_running ? "Running" : "Idle"}
                        </Badge>
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
  );
}
