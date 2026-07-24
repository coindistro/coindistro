// ─── Coindistro Frontend API Types ──────────────────────
// Generated from backend models - keep in sync with Go models

/** Standard API envelope */
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

/** Auth user (subset of full user) */
export interface AuthUser {
  id: string;
  email: string;
  username?: string | null;
  display_name?: string | null;
  avatar_url?: string | null;
  roles?: string[];
  referral_code?: string;
  is_genesis?: boolean;
  is_founder?: boolean;
  status?: string;
}

/** Full user profile */
export interface UserProfile {
  id: string;
  username?: string | null;
  email: string;
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

/** Auth tokens response */
export interface AuthTokens {
  access_token: string;
  refresh_token: string;
  token_type?: string;
  expires_in?: number;
}

export interface AuthPayload extends AuthTokens {
  user: AuthUser;
}

/** Login/Register inputs */
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

export interface ForgotPasswordInput {
  email: string;
}

export interface ResetPasswordInput {
  token: string;
  password: string;
}

export interface VerifyEmailInput {
  token: string;
}

export interface ChangePasswordInput {
  current_password: string;
  new_password: string;
}

export interface UpdateProfileInput {
  display_name?: string;
  username?: string;
  country?: string;
  timezone?: string;
  avatar_url?: string;
}

/** Session */
export interface Session {
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

/** Device */
export interface Device {
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

/** Activity log */
export interface ActivityLog {
  id: string;
  action: string;
  ip_address?: string | null;
  device_id?: string | null;
  details?: Record<string, unknown> | null;
  created_at: string;
}

/** Referral dashboard */
export interface ReferralDashboard {
  referral_code: string;
  referral_link: string;
  invitation_credits: number;
  total_invites: number;
  successful_invites: number;
  pending_invites: number;
  conversion_rate: number;
  referral_tree?: ReferralNode[];
  leaderboard_rank: number;
  rewards_earned: number;
}

export interface ReferralNode {
  id: string;
  username?: string | null;
  level: number;
  date: string;
  children?: ReferralNode[];
}

/** Invitation */
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

/** Platform stats (admin) */
export interface PlatformStats {
  total_users: number;
  active_users: number;
  verified_users: number;
  genesis_members: number;
  total_referrals: number;
  total_invitations: number;
  genesis_config?: GenesisConfig;
  recent_registrations?: AdminUserSummary[];
  recent_logins?: AdminUserSummary[];
  recent_activity?: ActivityLog[];
}

export interface GenesisConfig {
  id: string;
  max_genesis_members: number;
  current_genesis_count: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
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

/** System status (admin) */
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
  workers: WorkerStatus;
  scheduler: SchedulerStatus;
  feature_flags: Record<string, boolean>;
}

export interface WorkerStatus {
  enabled: boolean;
  status: string;
  workers?: number;
  queue_size?: number;
}

export interface SchedulerStatus {
  enabled: boolean;
  status: string;
  task_count: number;
  tasks?: Array<{
    id: string;
    name: string;
    interval: string;
    last_run?: string;
    next_run?: string;
    status: string;
  }>;
}

/** Feature flags */
export interface FeatureFlags {
  flags: Record<string, boolean>;
}

/** Earn types */
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
  strategy_profiles?: string[];
  featured: boolean;
  metadata?: Record<string, unknown>;
  starts_at?: string | null;
  ends_at?: string | null;
  created_at: string;
  updated_at: string;
}

export interface EarnProductListFilter {
  category?: string;
  status?: string;
  featured?: boolean;
  asset?: string;
  page?: number;
  per_page?: number;
}

export interface PortfolioOverview {
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
  metadata?: Record<string, unknown>;
  created_at: string;
  updated_at: string;
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
  description?: string | null;
  period_start?: string | null;
  period_end?: string | null;
  granted_at?: string | null;
  metadata?: Record<string, unknown>;
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
  description?: string | null;
  metadata?: Record<string, unknown>;
  created_at: string;
}

export interface LaunchpoolCampaign {
  id: string;
  product_id?: string | null;
  name: string;
  description: string;
  supported_assets: string[];
  window_start: string;
  window_end: string;
  allocation_rules: Record<string, unknown>;
  reward_distribution: Record<string, unknown>;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface LearnCampaign {
  id: string;
  product_id?: string | null;
  name: string;
  description: string;
  academy_course_id?: string | null;
  reward_asset: string;
  reward_amount: number;
  status: string;
  starts_at?: string | null;
  ends_at?: string | null;
  created_at: string;
  updated_at: string;
}

export interface LearnCompletion {
  id: string;
  campaign_id: string;
  user_id: string;
  completed_at: string;
  reward_eligible: boolean;
  reward_granted: boolean;
  reward_id?: string | null;
  metadata?: Record<string, unknown>;
  created_at: string;
}

export interface ReferralMilestone {
  id: string;
  name: string;
  description: string;
  required_referrals: number;
  reward_asset: string;
  reward_amount: number;
  status: string;
  metadata?: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

export interface ReferralRewardClaim {
  id: string;
  user_id: string;
  milestone_id?: string | null;
  amount: number;
  asset: string;
  status: string;
  granted_at?: string | null;
  metadata?: Record<string, unknown>;
  created_at: string;
}

export interface ReferralRewardSummary {
  total_eligible: number;
  total_granted: number;
  claims: ReferralRewardClaim[];
  milestones: ReferralMilestone[];
}

export interface ProductAnalytics {
  product_id: string;
  participants: number;
  total_allocated: number;
  capacity_used: number;
  capacity_total?: number | null;
  capacity_used_pct: number;
  total_rewards: number;
  avg_allocation: number;
}

export interface PaginatedResponse<T> {
  data: T[];
  meta: {
    page: number;
    per_page: number;
    total: number;
    total_pages: number;
  };
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

export interface CompleteLearnInput {
  metadata?: Record<string, unknown>;
}