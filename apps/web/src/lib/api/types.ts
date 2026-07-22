/** Standard Coindistro API envelope. */
export interface ApiResponse<T = unknown> {
  success: boolean;
  message?: string;
  data?: T;
  meta?: {
    page?: number;
    per_page?: number;
    total?: number;
    total_pages?: number;
  };
  error?: {
    code: string;
    message: string;
    details?: unknown;
  };
}

export interface AuthUser {
  id: string;
  email: string;
  username?: string | null;
  display_name?: string | null;
  avatar_url?: string | null;
  country?: string | null;
  timezone?: string;
  roles?: string[];
  referral_code?: string;
  referred_by?: string | null;
  is_genesis?: boolean;
  genesis_number?: number | null;
  is_founder?: boolean;
  is_verified?: boolean;
  status?: string;
  last_login_at?: string | null;
  created_at?: string;
}

export interface AuthTokens {
  access_token: string;
  refresh_token: string;
  token_type?: string;
  expires_in?: number;
}

export interface AuthPayload extends AuthTokens {
  user: AuthUser;
}

export interface ReferralDashboard {
  referral_code: string;
  referral_link: string;
  invitation_credits: number;
  total_invites: number;
  successful_invites: number;
  pending_invites: number;
  conversion_rate: number;
  leaderboard_rank: number;
  rewards_earned: number;
  referral_tree?: unknown[];
}

export interface SessionInfo {
  id: string;
  browser?: string | null;
  operating_system?: string | null;
  device_name?: string | null;
  device_type?: string | null;
  ip_address?: string | null;
  country?: string | null;
  is_current: boolean;
  login_at: string;
  last_activity_at: string;
  expires_at: string;
}

export interface DeviceInfo {
  id: string;
  name?: string | null;
  browser?: string | null;
  operating_system?: string | null;
  device_type?: string | null;
  is_trusted: boolean;
  is_current: boolean;
  last_seen_at: string;
  first_seen_at: string;
}

export interface ActivityLog {
  id: string;
  action: string;
  ip_address?: string | null;
  device_id?: string | null;
  details?: Record<string, unknown>;
  created_at: string;
}

export interface Invitation {
  id: string;
  invitee_email: string;
  code: string;
  status: string;
  message?: string | null;
  expires_at: string;
  consumed_at?: string | null;
  created_at: string;
}

export interface EarnPortfolio {
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

export interface AdminUserSummary {
  id: string;
  email: string;
  username?: string | null;
  display_name?: string | null;
  status: string;
  is_verified: boolean;
  is_genesis: boolean;
  roles: string[];
  last_login_at?: string | null;
  created_at: string;
}

export interface PlatformStats {
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

export interface FeatureFlag {
  name: string;
  description?: string;
  enabled: boolean;
  environment?: string;
}

export interface SystemStatus {
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

export interface HealthResponse {
  status: string;
  timestamp: string;
  version: string;
  checks: Record<string, string>;
}

export class ApiError extends Error {
  status: number;
  code: string;
  details?: unknown;

  constructor(status: number, code: string, message: string, details?: unknown) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
    this.details = details;
  }
}

/** Roles that may access the admin control plane. */
export const ADMIN_ROLES = ["super_admin", "admin", "moderator"] as const;

export function isAdminRole(roles?: string[] | null): boolean {
  if (!roles?.length) return false;
  return roles.some((r) => (ADMIN_ROLES as readonly string[]).includes(r));
}

export function postLoginPath(roles?: string[] | null, next?: string | null): string {
  if (next && next.startsWith("/") && !next.startsWith("//")) {
    return next;
  }
  return isAdminRole(roles) ? "/admin" : "/app/dashboard";
}
