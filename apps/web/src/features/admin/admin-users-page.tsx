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
} from "@coindistro/cds";
import * as adminApi from "@/features/admin/api";
import { formatRelative } from "@/lib/utils/format";

export function AdminUsersPage() {
  const usersQ = useQuery({
    queryKey: ["admin", "users"],
    queryFn: () => adminApi.getAdminUsers({ page: 1, per_page: 50 }),
  });

  return (
    <div className="space-y-6 animate-cds-fade-in">
      <PageHeader
        title="Users"
        description="Platform user directory from the Identity Service."
        actions={
          <Badge variant="secondary">
            {usersQ.data?.total ?? 0} total
          </Badge>
        }
      />

      <Card>
        <CardHeader>
          <CardTitle className="text-base">All users</CardTitle>
        </CardHeader>
        <CardContent>
          {usersQ.isLoading ? (
            <Skeleton className="h-48 w-full" />
          ) : usersQ.isError ? (
            <p className="text-sm text-destructive">Failed to load users.</p>
          ) : !usersQ.data?.users.length ? (
            <p className="text-sm text-muted-foreground">No users found.</p>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full min-w-[640px] text-left text-sm">
                <thead className="text-xs text-muted-foreground">
                  <tr className="border-b">
                    <th className="pb-2 pr-3 font-medium">User</th>
                    <th className="pb-2 pr-3 font-medium">Status</th>
                    <th className="pb-2 pr-3 font-medium">Verified</th>
                    <th className="pb-2 pr-3 font-medium">Genesis</th>
                    <th className="pb-2 pr-3 font-medium">Roles</th>
                    <th className="pb-2 font-medium">Joined</th>
                  </tr>
                </thead>
                <tbody>
                  {usersQ.data.users.map((u) => (
                    <tr key={u.id} className="border-b border-border/50 last:border-0">
                      <td className="py-2.5 pr-3">
                        <div className="font-medium">
                          {u.display_name || u.username || u.email}
                        </div>
                        <div className="text-xs text-muted-foreground">{u.email}</div>
                      </td>
                      <td className="py-2.5 pr-3 capitalize">{u.status}</td>
                      <td className="py-2.5 pr-3">
                        <Badge variant={u.is_verified ? "success" : "secondary"}>
                          {u.is_verified ? "Yes" : "No"}
                        </Badge>
                      </td>
                      <td className="py-2.5 pr-3">{u.is_genesis ? "Yes" : "—"}</td>
                      <td className="py-2.5 pr-3 text-xs">
                        {(u.roles || []).join(", ") || "user"}
                      </td>
                      <td className="py-2.5 text-muted-foreground">
                        {formatRelative(u.created_at)}
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
