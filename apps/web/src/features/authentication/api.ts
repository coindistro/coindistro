import { api, setAuthTokens } from "@/lib/api/client";
import type { AuthPayload, AuthUser } from "@/lib/api/types";
import type {
  UpdateProfileInput,
  ChangePasswordInput,
  Session,
  Device,
  ReferralDashboard,
  Invitation,
  ActivityLog,
  UserProfile,
  ForgotPasswordInput,
  ResetPasswordInput,
  VerifyEmailInput,
} from "@/features/shared/api";

export interface LoginInput {
  email: string;
  password: string;
}

export interface RegisterInput {
  email: string;
  password: string;
  username?: string;
  display_name?: string;
  referral_code: string;
  country?: string;
}

export async function login(input: LoginInput): Promise<AuthPayload> {
  const data = await api.post<AuthPayload>("/api/v1/auth/login", input, {
    auth: false,
  });
  setAuthTokens(data.access_token, data.refresh_token);
  return data;
}

export async function register(input: RegisterInput): Promise<AuthPayload> {
  const data = await api.post<AuthPayload>("/api/v1/auth/register", input, {
    auth: false,
  });
  setAuthTokens(data.access_token, data.refresh_token);
  return data;
}

export async function logout(): Promise<void> {
  try {
    await api.post("/api/v1/auth/logout", {});
  } catch {
    // ignore — clear local session anyway
  } finally {
    setAuthTokens(null, null);
  }
}

export async function forgotPassword(input: ForgotPasswordInput): Promise<void> {
  await api.post("/api/v1/auth/forgot-password", input, { auth: false });
}

export async function resetPassword(input: ResetPasswordInput): Promise<void> {
  await api.post("/api/v1/auth/reset-password", input, { auth: false });
}

export async function verifyEmail(input: VerifyEmailInput): Promise<void> {
  await api.get(`/api/v1/auth/verify-email?token=${encodeURIComponent(input.token)}`, {
    auth: false,
  });
}

export async function resendVerification(): Promise<void> {
  await api.post("/api/v1/auth/resend-verification", {});
}

export async function changePassword(input: ChangePasswordInput): Promise<void> {
  await api.put("/api/v1/auth/change-password", input);
}

export async function getMe(): Promise<AuthUser> {
  return api.get<AuthUser>("/api/v1/users/me");
}

export async function getProfile(): Promise<UserProfile> {
  return api.get<UserProfile>("/api/v1/users/me");
}

export async function updateProfile(input: UpdateProfileInput): Promise<UserProfile> {
  return api.put<UserProfile>("/api/v1/users/me", input);
}

export async function getSessions(): Promise<Session[]> {
  return api.get<Session[]>("/api/v1/sessions");
}

export async function terminateSession(sessionId: string): Promise<void> {
  await api.delete(`/api/v1/sessions/${sessionId}`);
}

export async function terminateAllSessions(): Promise<void> {
  await api.post("/api/v1/sessions/terminate-all", {});
}

export async function getDevices(): Promise<Device[]> {
  return api.get<Device[]>("/api/v1/devices");
}

export async function removeDevice(deviceId: string): Promise<void> {
  await api.delete(`/api/v1/devices/${deviceId}`);
}

export async function getReferralDashboard(): Promise<ReferralDashboard> {
  return api.get<ReferralDashboard>("/api/v1/referrals/dashboard");
}

export async function getInvitations(): Promise<Invitation[]> {
  return api.get<Invitation[]>("/api/v1/invitations");
}

export async function sendInvitation(
  email: string,
  message?: string,
  role?: string
): Promise<Invitation> {
  return api.post<Invitation>("/api/v1/invitations", { email, message, role });
}

export async function getActivityLog(): Promise<ActivityLog[]> {
  return api.get<ActivityLog[]>("/api/v1/activity");
}

export async function checkEmailAvailability(email: string): Promise<boolean> {
  const res = await api.get<{ available: boolean }>("/api/v1/users/check-email", {
    query: { email },
  });
  return res.available;
}

export async function checkUsernameAvailability(username: string): Promise<boolean> {
  const res = await api.get<{ available: boolean }>("/api/v1/users/check-username", {
    query: { username },
  });
  return res.available;
}