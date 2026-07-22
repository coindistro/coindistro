import { api, getAccessToken } from "@/lib/api/client";
import type {
  FeatureFlag,
  HealthResponse,
  PlatformStats,
  SystemStatus,
  AdminUserSummary,
  ApiResponse,
} from "@/lib/api/types";
import { appConfig } from "@/lib/config";

export async function getAdminSystem(): Promise<SystemStatus> {
  return api.get<SystemStatus>("/api/v1/admin/system");
}

export async function getPlatformStats(): Promise<PlatformStats> {
  return api.get<PlatformStats>("/api/v1/admin/stats");
}

export async function getAdminFeatures(): Promise<FeatureFlag[]> {
  const data = await api.get<{ flags: FeatureFlag[] }>("/api/v1/admin/features");
  return data?.flags ?? [];
}

export async function setFeatureFlag(flag: string, enabled: boolean): Promise<void> {
  await api.put(`/api/v1/admin/features/${encodeURIComponent(flag)}`, { enabled });
}

export async function getAdminUsers(params?: {
  page?: number;
  per_page?: number;
  status?: string;
}): Promise<{ users: AdminUserSummary[]; total: number; page: number; per_page: number }> {
  const q = new URLSearchParams();
  if (params?.page) q.set("page", String(params.page));
  if (params?.per_page) q.set("per_page", String(params.per_page));
  if (params?.status) q.set("status", params.status);
  const qs = q.toString();
  const path = `/api/v1/admin/users${qs ? `?${qs}` : ""}`;

  const token = getAccessToken();
  const res = await fetch(`${appConfig.apiBaseUrl}${path}`, {
    headers: {
      Accept: "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
  });
  const json = (await res.json()) as ApiResponse<AdminUserSummary[]>;
  if (!res.ok || json.success === false) {
    throw new Error(json.error?.message || "Failed to load users");
  }
  return {
    users: json.data ?? [],
    total: json.meta?.total ?? (json.data?.length ?? 0),
    page: json.meta?.page ?? params?.page ?? 1,
    per_page: json.meta?.per_page ?? params?.per_page ?? 20,
  };
}

export async function getHealth(): Promise<HealthResponse> {
  const res = await fetch(`${appConfig.apiBaseUrl}/health`, {
    headers: { Accept: "application/json" },
  });
  return (await res.json()) as HealthResponse;
}

export async function getWorkersStatus(): Promise<Record<string, unknown>> {
  return api.get<Record<string, unknown>>("/api/v1/admin/workers");
}

export async function getSchedulerStatus(): Promise<Record<string, unknown>> {
  return api.get<Record<string, unknown>>("/api/v1/admin/scheduler");
}
