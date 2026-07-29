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

/** @deprecated Use Wallet + InvestmentDashboard (Genesis Investor Program). */
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

/** Internal CDT wallet balances (GET /api/v1/wallet). */
export interface Wallet {
  id: string;
  user_id: string;
  available_balance: number;
  locked_balance: number;
  staking_balance: number;
  total_balance: number;
  updated_at: string;
}

export type InvestmentStatus =
  | "pending"
  | "active"
  | "completed"
  | "failed"
  | "cancelled"
  | string;

/** Summary row for a Genesis investment. */
export interface InvestmentSummary {
  id: string;
  plan_name: string;
  amount_paid: number;
  allocated_cdt: number;
  roi_cdt: number;
  roi_percent: number;
  status: InvestmentStatus;
  lock_period_days: number;
  days_remaining?: number;
  progress_pct?: number;
  started_at?: string | null;
  matures_at?: string | null;
  created_at: string;
}

/** Dashboard payload (GET /api/v1/earn/investments). */
export interface InvestmentDashboard {
  total_invested: number;
  locked_cdt: number;
  available_cdt: number;
  total_roi_earned: number;
  active_investments: number;
  completed_investments: number;
  upcoming_maturity?: string | null;
  investments: InvestmentSummary[];
}

export interface InvestmentPlan {
  id: string;
  name: string;
  description?: string;
  minimum_amount: number;
  maximum_amount: number;
  currency: string;
  roi_percent: number;
  enabled: boolean;
  created_at?: string;
  updated_at?: string;
}

export interface WalletTransaction {
  id: string;
  wallet_id: string;
  type: string;
  amount: number;
  balance_before: number;
  balance_after: number;
  reference?: string;
  description?: string;
  created_at: string;
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

// Re-export role helpers so existing imports from @/lib/api/types keep working.
export {
  ADMIN_ROLES,
  MERCHANT_ROLES,
  isAdminRole,
  isMerchantRole,
  isSuperAdminRole,
  postLoginPath,
  defaultHomeForRoles,
  hasAnyRole,
  permissionsForRoles,
} from "@/features/authentication/roles";
