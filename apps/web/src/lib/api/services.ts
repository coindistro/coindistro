/**
 * Coindistro API Service Layer
 * Maps to all backend Identity, Earn, and Admin endpoints
 */

import {
  api,
  type RequestOptions,
  setUnauthorizedHandler,
  loadAuthTokensFromStorage,
} from "@/lib/api/client";
import type {
  ApiResponse,
  AuthUser,
  AuthPayload,
  ReferralDashboard,
  SessionInfo,
  DeviceInfo,
  ActivityLog,
  Invitation,
  EarnPortfolio,
  AdminUserSummary,
  PlatformStats,
  SystemStatus,
  FeatureFlag,
  HealthResponse,
  ApiError,
} from "@/lib/api/types";

// Initialize token loading on client side
if (typeof window !== "undefined") {
  loadAuthTokensFromStorage();
}

// ─── Generic Request Wrapper ──────────────────────────────────

function handleApiError(error: unknown): never {
  if (error instanceof ApiError) throw error;
  if (error instanceof Error) throw new ApiError(0, "NETWORK_ERROR", error.message);
  throw new ApiError(0, "UNKNOWN_ERROR", "An unknown error occurred");
}

function unwrap<T>(promise: Promise<ApiResponse<T>>): Promise<T> {
  return promise.then((res) => {
    if (res.success && res.data !== undefined) return res.data;
    if (!res.success && res.error) throw new ApiError(0, res.error.code, res.error.message, res.error.details);
    return res.data as T;
  });
}

// ─── Authentication API ───────────────────────────────────────

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
  timezone?: string;
}

export async function login(input: LoginInput): Promise<AuthPayload> {
  return unwrap(
    api.post<AuthPayload>("/api/v1/auth/login", input, { auth: false })
  );
}

export async function register(input: RegisterInput): Promise<AuthPayload> {
  return unwrap(
    api.post<AuthPayload>("/api/v1/auth/register", input, { auth: false })
  );
}

export async function logout(): Promise<void> {
  await unwrap(api.post("/api/v1/auth/logout", {}));
}

export async function refreshToken(): Promise<AuthPayload> {
  return unwrap(api.post<AuthPayload>("/api/v1/auth/refresh", { auth: false }));
}

export async function verifyEmail(token: string): Promise<void> {
  await unwrap(
    api.get(`/api/v1/auth/verify-email?token=${encodeURIComponent(token)}`, { auth: false })
  );
}

export async function resendVerification(): Promise<void> {
  await unwrap(api.post("/api/v1/auth/resend-verification", {}));
}

export async function forgotPassword(email: string): Promise<void> {
  await unwrap(api.post("/api/v1/auth/forgot-password", { email }, { auth: false }));
}

export async function resetPassword(token: string, password: string): Promise<void> {
  await unwrap(
    api.post("/api/v1/auth/reset-password", { token, password }, { auth: false })
  );
}

export async function changePassword(currentPassword: string, newPassword: string): Promise<void> {
  await unwrap(api.put("/api/v1/auth/change-password", { current_password: currentPassword, new_password: newPassword }));
}

// ─── User Profile API ─────────────────────────────────────────

export interface ProfileResponse {
  id: string;
  email: string;
  username?: string | null;
  display_name?: string | null;
  avatar_url?: string | null;
  country?: string | null;
  timezone: string;
  referral_code: string;
  referred_by?: string | null;
  status: string;
  is_verified: boolean;
  is_genesis: boolean;
  genesis_number?: number | null;
  is_founder: boolean;
  roles: string[];
  last_login_at?: string | null;
  created_at: string;
}

export interface UpdateProfileInput {
  display_name?: string;
  username?: string;
  country?: string;
  timezone?: string;
  avatar_url?: string;
}

export async function getProfile(): Promise<ProfileResponse> {
  return unwrap(api.get<ProfileResponse>("/api/v1/users/me"));
}

export async function updateProfile(input: UpdateProfileInput): Promise<ProfileResponse> {
  return unwrap(api.put<ProfileResponse>("/api/v1/users/me", input));
}

// ─── Session & Device API ──────────────────────────────────────

export async function getSessions(): Promise<SessionInfo[]> {
  return unwrap(api.get<SessionInfo[]>("/api/v1/sessions"));
}

export async function terminateSession(sessionId: string): Promise<void> {
  await unwrap(api.delete(`/api/v1/sessions/${sessionId}`));
}

export async function terminateAllSessions(): Promise<void> {
  await unwrap(api.post("/api/v1/sessions/terminate-all", {}));
}

export async function getDevices(): Promise<DeviceInfo[]> {
  return unwrap(api.get<DeviceInfo[]>("/api/v1/devices"));
}

export async function removeDevice(deviceId: string): Promise<void> {
  await unwrap(api.delete(`/api/v1/devices/${deviceId}`));
}

// ─── Referral & Invitation API ────────────────────────────────

export async function getReferralDashboard(): Promise<ReferralDashboard> {
  return unwrap(api.get<ReferralDashboard>("/api/v1/referrals/dashboard"));
}

export async function getInvitations(): Promise<Invitation[]> {
  return unwrap(api.get<Invitation[]>("/api/v1/invitations"));
}

export interface SendInvitationInput {
  email: string;
  message?: string;
  role?: string;
}

export async function sendInvitation(input: SendInvitationInput): Promise<Invitation> {
  return unwrap(api.post<Invitation>("/api/v1/invitations", input));
}

// ─── Security & Activity API ──────────────────────────────────

export async function getActivityLog(): Promise<ActivityLog[]> {
  return unwrap(api.get<ActivityLog[]>("/api/v1/activity"));
}

// ─── Earn API ──────────────────────────────────────────────────

export interface EarnProduct {
  id: string;
  name: string;
  slug: string;
  description: string;
  category: string;
  supported_assets: string[];
  duration_days?: number | null;
  capacity_total?: number | null;
  capacity_used: number;
  status: string;
  risk_level: string;
  min_allocation: number;
  max_allocation?: number | null;
  reward_model: string;
  reward_apr: number;
  eligibility: Record<string, unknown>;
  rules: Record<string, unknown>;
  strategy_profiles: string[];
  featured: boolean;
  metadata: Record<string, unknown>;
  starts_at?: string | null;
  ends_at?: string | null;
  created_by?: string | null;
  created_at: string;
  updated_at: string;
}

export interface EarnProductListResponse {
  products: EarnProduct[];
  meta: {
    page: number;
    per_page: number;
    total: number;
    total_pages: number;
  };
}

export interface EarnPortfolioResponse {
  total_assets_in_earn: number;
  estimated_rewards: number;
  todays_rewards: number;
  lifetime_rewards: number;
  active_products: number;
  available_balance: number;
  locked_balance: number;
  allocation_by_product: Record<string, number>;
  allocation_by_asset: Record<string, number>;
}

export interface Participation {
  id: string;
  user_id: string;
  product_id: string;
  asset: string;
  allocated_amount: number;
  current_balance: number;
  estimated_rewards: number;
  accrued_rewards: number;
  lifetime_rewards: number;
  status: string;
  strategy_profile?: string | null;
  joined_at: string;
  lock_start_at?: string | null;
  lock_end_at?: string | null;
  completed_at?: string | null;
  exited_at?: string | null;
  metadata: Record<string, unknown>;
  product?: EarnProduct;
}

export interface Reward {
  id: string;
  user_id: string;
  product_id: string;
  participation_id?: string | null;
  asset: string;
  amount: number;
  reward_type: string;
  status: string;
  description?: string;
  period_start?: string | null;
  period_end?: string | null;
  granted_at?: string | null;
  metadata: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

export interface Transaction {
  id: string;
  user_id: string;
  product_id?: string | null;
  participation_id?: string | null;
  type: string;
  asset: string;
  amount: number;
  balance_after?: number | null;
  status: string;
  reference?: string | null;
  description?: string;
  metadata: Record<string, unknown>;
  created_at: string;
}

export interface JoinProductInput {
  asset: string;
  amount: number;
  strategy_profile?: string;
}

export interface AddFundsInput {
  amount: number;
}

export interface WithdrawInput {
  amount: number;
}

export interface ExitParticipationInput {
  reason?: string;
}

export async function listEarnProducts(params?: {
  category?: string;
  status?: string;
  featured?: boolean;
  asset?: string;
  page?: number;
  per_page?: number;
}): Promise<EarnProductListResponse> {
  const searchParams = new URLSearchParams();
  if (params?.category) searchParams.set("category", params.category);
  if (params?.status) searchParams.set("status", params.status);
  if (params?.featured) searchParams.set("featured", "true");
  if (params?.asset) searchParams.set("asset", params.asset);
  if (params?.page) searchParams.set("page", String(params.page));
  if (params?.per_page) searchParams.set("per_page", String(params.per_page));
  const query = searchParams.toString();
  return unwrap(api.get<EarnProductListResponse>(`/api/v1/earn/products${query ? `?${query}` : ""}`));
}

export async function getEarnProduct(id: string): Promise<EarnProduct> {
  return unwrap(api.get<EarnProduct>(`/api/v1/earn/products/${id}`));
}

export async function getEarnPortfolio(): Promise<EarnPortfolioResponse> {
  return unwrap(api.get<EarnPortfolioResponse>("/api/v1/earn/portfolio"));
}

export async function joinEarnProduct(productId: string, input: JoinProductInput): Promise<Participation> {
  return unwrap(api.post<Participation>(`/api/v1/earn/products/${productId}/join`, input));
}

export async function listParticipations(status?: string, page = 1, perPage = 20): Promise<{ participations: Participation[]; meta: { page: number; per_page: number; total: number; total_pages: number } }> {
  const searchParams = new URLSearchParams();
  if (status) searchParams.set("status", status);
  searchParams.set("page", String(page));
  searchParams.set("per_page", String(perPage));
  return unwrap(api.get(`/api/v1/earn/participations?${searchParams.toString()}`));
}

export async function getParticipation(participationId: string): Promise<Participation> {
  return unwrap(api.get<Participation>(`/api/v1/earn/participations/${participationId}`));
}

export async function addFundsToParticipation(participationId: string, input: AddFundsInput): Promise<Participation> {
  return unwrap(api.post<Participation>(`/api/v1/earn/participations/${participationId}/add-funds`, input));
}

export async function withdrawFromParticipation(participationId: string, input: WithdrawInput): Promise<Participation> {
  return unwrap(api.post<Participation>(`/api/v1/earn/participations/${participationId}/withdraw`, input));
}

export async function exitParticipation(participationId: string, input?: ExitParticipationInput): Promise<Participation> {
  return unwrap(api.post<Participation>(`/api/v1/earn/participations/${participationId}/exit`, input ?? {}));
}

export async function listRewards(page = 1, perPage = 20): Promise<{ rewards: Reward[]; meta: { page: number; per_page: number; total: number; total_pages: number } }> {
  const searchParams = new URLSearchParams();
  searchParams.set("page", String(page));
  searchParams.set("per_page", String(perPage));
  return unwrap(api.get(`/api/v1/earn/rewards?${searchParams.toString()}`));
}

export async function listTransactions(page = 1, perPage = 20): Promise<{ transactions: Transaction[]; meta: { page: number; per_page: number; total: number; total_pages: number } }> {
  const searchParams = new URLSearchParams();
  searchParams.set("page", String(page));
  searchParams.set("per_page", String(perPage));
  return unwrap(api.get(`/api/v1/earn/history?${searchParams.toString()}`));
}

export async function listLaunchpools(status?: string): Promise<unknown[]> {
  const searchParams = new URLSearchParams();
  if (status) searchParams.set("status", status);
  return unwrap(api.get(`/api/v1/earn/launchpool${searchParams.toString() ? `?${searchParams.toString()}` : ""}`));
}

export async function listLearnCampaigns(status?: string): Promise<unknown[]> {
  const searchParams = new URLSearchParams();
  if (status) searchParams.set("status", status);
  return unwrap(api.get(`/api/v1/earn/learn${searchParams.toString() ? `?${searchParams.toString()}` : ""}`));
}

export async function completeLearnCampaign(campaignId: string, metadata?: Record<string, unknown>): Promise<unknown> {
  return unwrap(api.post(`/api/v1/earn/learn/${campaignId}/complete`, { metadata }));
}

export async function getReferralRewards(): Promise<unknown> {
  return unwrap(api.get("/api/v1/earn/referral/rewards"));
}

// ─── Admin API ─────────────────────────────────────────────────

export interface AdminDashboardSystemStatus {
  status: string;
  api_status: string;
  database: string;
  redis: string;
  backend: string;
  docker: string;
  version: string;
  environment: string;
  app_name: string;
  timestamp: string;
  workers: Record<string, unknown>;
  scheduler: Record<string, unknown>;
  feature_flags: FeatureFlag[];
}

export interface AdminDashboardStats {
  total_users: number;
  verified_users: number;
  genesis_members: number;
  active_users: number;
  total_referrals: number;
  total_invitations: number;
  recent_registrations?: AdminUserSummary[];
  recent_logins?: AdminUserSummary[];
  recent_activity?: ActivityLog[];
  genesis_config?: {
    id: string;
    max_genesis_members: number;
    current_genesis_count: number;
    is_active: boolean;
  };
}

export interface AdminUsersListResponse {
  users: AdminUserSummary[];
  meta: {
    page: number;
    per_page: number;
    total: number;
    total_pages: number;
  };
}

export async function getAdminSystemStatus(): Promise<AdminDashboardSystemStatus> {
  return unwrap(api.get<AdminDashboardSystemStatus>("/api/v1/admin/system"));
}

export async function getAdminDashboardStats(): Promise<AdminDashboardStats> {
  return unwrap(api.get<AdminDashboardStats>("/api/v1/admin/dashboard/stats"));
}

export async function getAdminUsers(page = 1, perPage = 20, status?: string): Promise<AdminUsersListResponse> {
  const searchParams = new URLSearchParams();
  searchParams.set("page", String(page));
  searchParams.set("per_page", String(perPage));
  if (status) searchParams.set("status", status);
  return unwrap(api.get<AdminUsersListResponse>(`/api/v1/admin/users?${searchParams.toString()}`));
}

export async function getAdminRecentRegistrations(limit = 10): Promise<AdminUserSummary[]> {
  return unwrap(api.get<AdminUserSummary[]>(`/api/v1/admin/registrations/recent?limit=${limit}`));
}

export async function getAdminRecentLogins(limit = 10): Promise<AdminUserSummary[]> {
  return unwrap(api.get<AdminUserSummary[]>(`/api/v1/admin/logins/recent?limit=${limit}`));
}

export async function getAdminActivityLog(limit = 20): Promise<ActivityLog[]> {
  return unwrap(api.get<ActivityLog[]>(`/api/v1/admin/activity?limit=${limit}`));
}

export async function getAdminGenesisStats(): Promise<{
  id: string;
  max_genesis_members: number;
  current_genesis_count: number;
  is_active: boolean;
}> {
  return unwrap(api.get("/api/v1/admin/genesis/stats"));
}

export async function getAdminReferralStats(): Promise<{
  total_referrals: number;
  total_invitations: number;
}> {
  return unwrap(api.get("/api/v1/admin/referrals/stats"));
}

export async function getAdminEarnStats(): Promise<{ message: string }> {
  return unwrap(api.get("/api/v1/admin/earn/stats"));
}

export async function getAdminRoles(): Promise<Record<string, string[]>> {
  return unwrap(api.get("/api/v1/admin/roles"));
}

export async function getAdminRolePermissions(role: string): Promise<{ role: string; permissions: string[] }> {
  return unwrap(api.get(`/api/v1/admin/roles/${role}/permissions`));
}

export async function getAdminFeatureFlags(): Promise<FeatureFlag[]> {
  return unwrap(api.get<FeatureFlag[]>("/api/v1/admin/features"));
}

export async function setAdminFeatureFlag(flag: string, enabled: boolean): Promise<{ flag: string; enabled: boolean }> {
  return unwrap(api.put(`/api/v1/admin/features/${flag}`, { enabled }));
}

export async function getAdminWorkers(): Promise<Record<string, unknown>> {
  return unwrap(api.get("/api/v1/admin/workers"));
}

export async function getAdminScheduler(): Promise<{ enabled: boolean; status: string; task_count: number; tasks: unknown[] }> {
  return unwrap(api.get("/api/v1/admin/scheduler"));
}

// ─── Health & Features ────────────────────────────────────────

export async function healthCheck(): Promise<HealthResponse> {
  return unwrap(api.get<HealthResponse>("/health"));
}

export async function getFeatureFlags(): Promise<FeatureFlag[]> {
  const res = await unwrap(api.get<{ flags: FeatureFlag[] }>("/api/v1/features"));
  return res.flags ?? [];
}

// ─── User Availability Checks ──────────────────────────────────

export async function checkEmailAvailability(email: string): Promise<{ available: boolean }> {
  return unwrap(api.get(`/api/v1/users/check-email?email=${encodeURIComponent(email)}`, { auth: false }));
}

export async function checkUsernameAvailability(username: string): Promise<{ available: boolean }> {
  return unwrap(api.get(`/api/v1/users/check-username?username=${encodeURIComponent(username)}`, { auth: false }));
}

// ─── Auth Utilities ────────────────────────────────────────────

export function setupUnauthorizedRedirect(handler: () => void) {
  setUnauthorizedHandler(handler);
}

export function getAccessToken(): string | null {
  // This is handled by the api client internally
  return null; // Not exposed directly, use api client
}