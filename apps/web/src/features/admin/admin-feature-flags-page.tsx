"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  Badge,
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  PageHeader,
  Skeleton,
  Switch,
} from "@coindistro/cds";
import * as adminApi from "@/features/admin/api";
import { useToast } from "@/features/shared/providers/toast-provider";

export function AdminFeatureFlagsPage() {
  const { toast } = useToast();
  const qc = useQueryClient();
  const flagsQ = useQuery({
    queryKey: ["admin", "features"],
    queryFn: adminApi.getAdminFeatures,
  });

  const toggleMut = useMutation({
    mutationFn: ({ name, enabled }: { name: string; enabled: boolean }) =>
      adminApi.setFeatureFlag(name, enabled),
    onSuccess: async (_, vars) => {
      toast({
        message: `${vars.name} ${vars.enabled ? "enabled" : "disabled"}`,
        variant: "success",
      });
      await qc.invalidateQueries({ queryKey: ["admin", "features"] });
      await qc.invalidateQueries({ queryKey: ["admin", "system"] });
    },
    onError: () => {
      toast({ message: "Failed to update flag", variant: "danger" });
    },
  });

  return (
    <div className="space-y-6 animate-cds-fade-in">
      <PageHeader
        title="Feature flags"
        description="Runtime feature flag management for Coindistro modules."
      />

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Flags</CardTitle>
        </CardHeader>
        <CardContent>
          {flagsQ.isLoading ? (
            <Skeleton className="h-40 w-full" />
          ) : !flagsQ.data?.length ? (
            <p className="text-sm text-muted-foreground">No feature flags registered.</p>
          ) : (
            <ul className="divide-y divide-border/60">
              {flagsQ.data.map((f) => (
                <li
                  key={f.name}
                  className="flex flex-col gap-2 py-3 sm:flex-row sm:items-center sm:justify-between"
                >
                  <div>
                    <div className="flex items-center gap-2">
                      <code className="text-sm font-medium">{f.name}</code>
                      <Badge variant={f.enabled ? "success" : "secondary"}>
                        {f.enabled ? "On" : "Off"}
                      </Badge>
                    </div>
                    {f.description ? (
                      <p className="text-xs text-muted-foreground">{f.description}</p>
                    ) : null}
                  </div>
                  <Switch
                    checked={f.enabled}
                    disabled={toggleMut.isPending}
                    onCheckedChange={(checked) =>
                      toggleMut.mutate({ name: f.name, enabled: checked })
                    }
                    aria-label={`Toggle ${f.name}`}
                  />
                </li>
              ))}
            </ul>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
