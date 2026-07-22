import { api } from "@/lib/api/client";
import type {
  ActivityLog,
  AuthUser,
  DeviceInfo,
  Invitation,
  ReferralDashboard,
  SessionInfo,
} from "@/lib/api/types";

export type UpdateProfileInput = {
  display_name?: string;
  username?: string;
  country?: string;
  timezone?: string;
  avatar_url?: string;
};

export async function getProfile(): Promise<AuthUser> {
  return api.get<AuthUser>("/api/v1/users/me");
}

export async function updateProfile(input: UpdateProfileInput): Promise<AuthUser> {
  return api.put<AuthUser>("/api/v1/users/me", input);
}

export async function getReferralDashboard(): Promise<ReferralDashboard> {
  return api.get<ReferralDashboard>("/api/v1/referrals/dashboard");
}

export async function getSessions(): Promise<SessionInfo[]> {
  const data = await api.get<SessionInfo[] | { sessions: SessionInfo[] }>("/api/v1/sessions");
  if (Array.isArray(data)) return data;
  return data?.sessions ?? [];
}

export async function terminateSession(id: string): Promise<void> {
  await api.delete(`/api/v1/sessions/${id}`);
}

export async function terminateAllSessions(): Promise<void> {
  await api.post("/api/v1/sessions/terminate-all", {});
}

export async function getDevices(): Promise<DeviceInfo[]> {
  const data = await api.get<DeviceInfo[] | { devices: DeviceInfo[] }>("/api/v1/devices");
  if (Array.isArray(data)) return data;
  return data?.devices ?? [];
}

export async function removeDevice(id: string): Promise<void> {
  await api.delete(`/api/v1/devices/${id}`);
}

export async function getActivityLog(): Promise<ActivityLog[]> {
  const data = await api.get<ActivityLog[] | { activity: ActivityLog[] }>("/api/v1/activity");
  if (Array.isArray(data)) return data;
  return data?.activity ?? [];
}

export async function getInvitations(): Promise<Invitation[]> {
  const data = await api.get<Invitation[] | { invitations: Invitation[] }>("/api/v1/invitations");
  if (Array.isArray(data)) return data;
  return data?.invitations ?? [];
}

export async function sendInvitation(email: string, message?: string): Promise<Invitation> {
  return api.post<Invitation>("/api/v1/invitations", { email, message });
}

export async function changePassword(
  current_password: string,
  new_password: string,
): Promise<void> {
  await api.put("/api/v1/auth/change-password", { current_password, new_password });
}
