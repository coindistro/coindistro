// ─── Coindistro Frontend API Service ────────────────────
// Centralized API calls for all backend endpoints

import { api } from "@/lib/api/client";
import type {
  // Auth
  LoginInput,
  RegisterInput,
  ForgotPasswordInput,
  ResetPasswordInput,
  VerifyEmailInput,
  ChangePasswordInput,
  AuthPayload,
  AuthUser,
  // Profile
  UserProfile,
  UpdateProfileInput,
  // Sessions
  Session,
  // Devices
  Device,
  // Activity
  ActivityLog,
  // Referrals
  ReferralDashboard,
  Invitation,
  // Admin
  PlatformStats,
  SystemStatus,
  AdminUserSummary,
  FeatureFlags,
  // Earn
  EarnProduct,
  EarnProductListFilter,
  PortfolioOverview,
  Participation,
  Reward,
  Transaction,
  LaunchpoolCampaign,
  LearnCampaign,
  LearnCompletion,
  ReferralRewardSummary,
  ProductAnalytics,
  PaginatedResponse,
  JoinProductInput,
  AddFundsInput,
  WithdrawInput,
  ExitParticipationInput,
  CompleteLearnInput,
} from "@/features/shared/api-types";

// ============================================================
// AUTHENTICATION
// ============================================================

export async function login(input: LoginInput): Promise<AuthPayload> {
  const data = await api.post<AuthPayload>("/api/v1/auth/login", input, { auth: false });
  return data;
}

export async function register(input: RegisterInput): Promise<AuthPayload> {
  const data = await api.post<AuthPayload>("/api/v1/auth/register", input, { auth: false });
  return data;
}

export async function logout(): Promise<void> {
  try {
    await api.post("/api/v1/auth/logout", {});
  } catch {
    // ignore — clear local session anyway
  }
}

export async function forgotPassword(input: ForgotPasswordInput): Promise<void> {
  await api.post("/api/v1/auth/forgot-password", input, { auth: false });
}

export async function resetPassword(input: ResetPasswordInput): Promise<void> {
  await api.post("/api/v1/auth/reset-password", input, { auth: false });
}

export async function verifyEmail(input: VerifyEmailInput): Promise<void> {
  await api.get(`/api/v1/auth/verify-email?token=${encodeURIComponent(input.token)}`, { auth: false });
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

// ============================================================
// USER PROFILE
// ============================================================

export async function getProfile(): Promise<UserProfile> {
  return api.get<UserProfile>("/api/v1/users/me");
}

export async function updateProfile(input: UpdateProfileInput): Promise<UserProfile> {
  return api.put<UserProfile>("/api/v1/users/me", input);
}

export async function checkEmailAvailability(email: string): Promise<{ available: boolean }> {
  return api.get<{ available: boolean }>(`/api/v1/users/check-email?email=${encodeURIComponent(email)}`);
}

export async function checkUsernameAvailability(username: string): Promise<{ available: boolean }> {
  return api.get<{ available: boolean }>(`/api/v1/users/check-username?username=${encodeURIComponent(username)}`);
}

// ============================================================
// SESSIONS
// ============================================================

export async function getSessions(): Promise<Session[]> {
  return api.get<Session[]>("/api/v1/sessions");
}

export async function terminateSession(sessionId: string): Promise<void> {
  await api.delete(`/api/v1/sessions/${sessionId}`);
}

export async function terminateAllSessions(): Promise<void> {
  await api.post("/api/v1/sessions/terminate-all", {});
}

// ============================================================
// DEVICES
// ============================================================

export async function getDevices(): Promise<Device[]> {
  return api.get<Device[]>("/api/v1/devices");
}

export async function removeDevice(deviceId: string): Promise<void> {
  await api.delete(`/api/v1/devices/${deviceId}`);
}

// ============================================================
// ACTIVITY
// ============================================================

export async function getActivityLog(): Promise<ActivityLog[]> {
  return api.get<ActivityLog[]>("/api/v1/activity");
}

// ============================================================
// REFERRALS
// ============================================================

export async function getReferralDashboard(): Promise<ReferralDashboard> {
  return api.get<ReferralDashboard>("/api/v1/referrals/dashboard");
}

// ============================================================
// INVITATIONS
// ============================================================

export async function getInvitations(): Promise<Invitation[]> {
  return api.get<Invitation[]>("/api/v1/invitations");
}

export interface SendInvitationInput {
  email: string;
  message?: string;
  role?: string;
}

export async function sendInvitation(input: SendInvitationInput): Promise<Invitation> {
  return api.post<Invitation>("/api/v1/invitations", input);
}

// ============================================================
// ADMIN - PLATFORM STATS
// ============================================================

export async function getAdminDashboardStats(): Promise<PlatformStats> {
  return api.get<PlatformStats>("/api/v1/admin/dashboard/stats");
}

export async function getAdminUsers(params?: { status?: string; page?: number; per_page?: number }): Promise<PaginatedResponse<AdminUserSummary>> {
  const search = new URLSearchParams();
  if (params?.status) search.set("status", params.status);
  if (params?.page) search.set("page", String(params.page));
  if (params?.per_page) search.set("per_page", String(params.per_page));
  return api.get<PaginatedResponse<AdminUserSummary>>(`/api/v1/admin/users?${search.toString()}`);
}

export async function getAdminRecentRegistrations(limit = 10): Promise<AdminUserSummary[]> {
  return api.get<AdminUserSummary[]>(`/api/v1/admin/registrations/recent?limit=${limit}`);
}

export async function getAdminRecentLogins(limit = 10): Promise<AdminUserSummary[]> {
  return api.get<AdminUserSummary[]>(`/api/v1/admin/logins/recent?limit=${limit}`);
}

export async function getAdminActivityLog(limit = 20): Promise<ActivityLog[]> {
  return api.get<ActivityLog[]>(`/api/v1/admin/activity?limit=${limit}`);
}

export async function getAdminGenesisStats(): Promise<FeatureFlags> {
  // Returns genesis config via platform stats
  return api.get<FeatureFlags>("/api/v1/admin/genesis/stats");
}

export async function getAdminReferralStats(): Promise<{ total_referrals: number; total_invitations: number }> {
  return api.get("/api/v1/admin/referrals/stats");
}

export async function getAdminEarnStats(): Promise<{ message: string }> {
  return api.get("/api/v1/admin/earn/stats");
}

// ============================================================
// ADMIN - SYSTEM STATUS
// ============================================================

export async function getSystemStatus(): Promise<SystemStatus> {
  return api.get<SystemStatus>("/api/v1/admin/system");
}

export async function getWorkersStatus(): Promise<SystemStatus["workers"]> {
  return api.get("/api/v1/admin/workers");
}

export async function getSchedulerStatus(): Promise<SystemStatus["scheduler"]> {
  return api.get("/api/v1/admin/scheduler");
}

export async function getFeatureFlags(): Promise<FeatureFlags> {
  return api.get<FeatureFlags>("/api/v1/admin/features");
}

export async function updateFeatureFlag(flag: string, enabled: boolean): Promise<{ flag: string; enabled: boolean }> {
  return api.put(`/api/v1/admin/features/${flag}`, { enabled });
}

export async function getRBACRoles(): Promise<{ roles: string[] }> {
  return api.get("/api/v1/admin/roles");
}

export async function getRolePermissions(role: string): Promise<{ role: string; permissions: string[] }> {
  return api.get(`/api/v1/admin/roles/${role}/permissions`);
}

// ============================================================
// EARN - PRODUCTS
// ============================================================

export async function listEarnProducts(filter?: EarnProductListFilter): Promise<PaginatedResponse<EarnProduct>> {
  const search = new URLSearchParams();
  if (filter?.category) search.set("category", filter.category);
  if (filter?.status) search.set("status", filter.status);
  if (filter?.featured !== undefined) search.set("featured", String(filter.featured));
  if (filter?.asset) search.set("asset", filter.asset);
  if (filter?.page) search.set("page", String(filter.page));
  if (filter?.per_page) search.set("per_page", String(filter.per_page));
  return api.get<PaginatedResponse<EarnProduct>>(`/api/v1/earn/products?${search.toString()}`);
}

export async function getEarnProduct(id: string): Promise<EarnProduct> {
  return api.get<EarnProduct>(`/api/v1/earn/products/${id}`);
}

// ============================================================
// EARN - PORTFOLIO
// ============================================================

export async function getPortfolioOverview(): Promise<PortfolioOverview> {
  return api.get<PortfolioOverview>("/api/v1/earn/portfolio");
}

// ============================================================
// EARN - PARTICIPATIONS
// ============================================================

export async function joinProduct(productId: string, input: JoinProductInput): Promise<Participation> {
  return api.post<Participation>(`/api/v1/earn/products/${productId}/join`, input);
}

export async function listParticipations(status?: string): Promise<PaginatedResponse<Participation>> {
  const search = new URLSearchParams();
  if (status) search.set("status", status);
  return api.get<PaginatedResponse<Participation>>(`/api/v1/earn/participations?${search.toString()}`);
}

export async function getParticipation(id: string): Promise<Participation> {
  return api.get<Participation>(`/api/v1/earn/participations/${id}`);
}

export async function addFunds(participationId: string, input: AddFundsInput): Promise<Participation> {
  return api.post<Participation>(`/api/v1/earn/participations/${participationId}/add-funds`, input);
}

export async function withdraw(participationId: string, input: WithdrawInput): Promise<Participation> {
  return api.post<Participation>(`/api/v1/earn/participations/${participationId}/withdraw`, input);
}

export async function exitParticipation(participationId: string, input?: ExitParticipationInput): Promise<Participation> {
  return api.post<Participation>(`/api/v1/earn/participations/${participationId}/exit`, input ?? {});
}

// ============================================================
// EARN - REWARDS & HISTORY
// ============================================================

export async function listRewards(): Promise<PaginatedResponse<Reward>> {
  return api.get<PaginatedResponse<Reward>>("/api/v1/earn/rewards");
}

export async function listTransactions(): Promise<PaginatedResponse<Transaction>> {
  return api.get<PaginatedResponse<Transaction>>("/api/v1/earn/history");
}

// ============================================================
// EARN - LAUNCHPOOL / LEARN / REFERRAL
// ============================================================

export async function listLaunchpools(status?: string): Promise<LaunchpoolCampaign[]> {
  const search = new URLSearchParams();
  if (status) search.set("status", status);
  return api.get<LaunchpoolCampaign[]>(`/api/v1/earn/launchpool?${search.toString()}`);
}

export async function listLearnCampaigns(status?: string): Promise<LearnCampaign[]> {
  const search = new URLSearchParams();
  if (status) search.set("status", status);
  return api.get<LearnCampaign[]>(`/api/v1/earn/learn?${search.toString()}`);
}

export async function completeLearnCampaign(campaignId: string, input?: CompleteLearnInput): Promise<LearnCompletion> {
  return api.post<LearnCompletion>(`/api/v1/earn/learn/${campaignId}/complete`, input ?? {});
}

export async function getReferralRewards(): Promise<ReferralRewardSummary> {
  return api.get<ReferralRewardSummary>("/api/v1/earn/referral/rewards");
}

// ============================================================
// EARN - ADMIN
// ============================================================

export async function adminListProducts(filter?: EarnProductListFilter): Promise<PaginatedResponse<EarnProduct>> {
  const search = new URLSearchParams();
  if (filter?.category) search.set("category", filter.category);
  if (filter?.status) search.set("status", filter.status);
  if (filter?.page) search.set("page", String(filter.page));
  if (filter?.per_page) search.set("per_page", String(filter.per_page));
  return api.get<PaginatedResponse<EarnProduct>>(`/api/v1/earn/admin/products?${search.toString()}`);
}

export async function adminGetProductAnalytics(productId: string): Promise<ProductAnalytics> {
  return api.get<ProductAnalytics>(`/api/v1/earn/admin/products/${productId}/analytics`);
}

export async function adminListParticipants(productId: string, page = 1, perPage = 20): Promise<PaginatedResponse<Participation>> {
  return api.get<PaginatedResponse<Participation>>(`/api/v1/earn/admin/products/${productId}/participants?page=${page}&per_page=${perPage}`);
}