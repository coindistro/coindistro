import { api, setAuthTokens } from "@/lib/api/client";
import type { AuthPayload, AuthUser } from "@/lib/api/types";

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

export async function forgotPassword(email: string): Promise<void> {
  await api.post("/api/v1/auth/forgot-password", { email }, { auth: false });
}

export async function resetPassword(token: string, password: string): Promise<void> {
  await api.post(
    "/api/v1/auth/reset-password",
    { token, password },
    { auth: false },
  );
}

export async function verifyEmail(token: string): Promise<void> {
  await api.get(`/api/v1/auth/verify-email?token=${encodeURIComponent(token)}`, {
    auth: false,
  });
}

export async function getMe(): Promise<AuthUser> {
  return api.get<AuthUser>("/api/v1/users/me");
}
