export interface InvestmentSettings {
  minimum_investment_usd: number;
  daily_reward_ngn: number;
  max_business_days: number;
  roi_percent: number;
  referral_percent: number;
  min_referrals_for_payout: number;
  early_withdrawal_penalty_percent: number;
  early_withdrawal_fee_percent: number;
  withdrawal_processing_hours: number;
  withdrawal_interval_days: number;
  enabled: boolean;
}

export interface ExchangeRate {
  usd_to_ngn: number;
  updated_at: string;
}

export interface WithdrawalFeeTier {
  id: string;
  min_amount: number;
  max_amount: number;
  fee_percent: number;
}

export type InvestmentStatus = 'pending_payment' | 'active' | 'completed' | 'paused' | 'cancelled' | 'early_withdrawal';

export interface EarningsInvestment {
  id: string;
  amount_usd: number;
  amount_ngn: number;
  exchange_rate: number;
  payment_provider: string;
  payment_reference: string;
  daily_reward_ngn: number;
  paid_business_days: number;
  max_business_days: number;
  total_earned_ngn: number;
  total_pending_ngn: number;
  status: InvestmentStatus;
  maturity_date?: string | null;
  started_at?: string | null;
  created_at: string;
}

export interface EarningsSummary {
  id: string;
  amount_usd: number;
  amount_ngn: number;
  exchange_rate: number;
  daily_reward_ngn: number;
  paid_business_days: number;
  max_business_days: number;
  remaining_days: number;
  total_earned_ngn: number;
  total_earned_usd?: number;
  total_pending_ngn: number;
  portfolio_value_usd?: number;
  portfolio_value_ngn?: number;
  roi_percentage?: number;
  status: InvestmentStatus;
  progress_pct: number;
  maturity_date?: string | null;
  started_at?: string | null;
  created_at: string;
}

export interface EarningsDashboard {
  total_invested_usd: number;
  total_invested_ngn: number;
  portfolio_value_usd?: number;
  portfolio_value_ngn?: number;
  total_profit_usd?: number;
  total_profit_ngn?: number;
  profit_earned_usd?: number;
  profit_earned_ngn?: number;
  capital_invested_usd?: number;
  capital_invested_ngn?: number;
  locked_balance_usd?: number;
  locked_balance_ngn?: number;
  withdrawable_balance_ngn?: number;
  roi_percentage?: number;
  today_earnings_ngn: number;
  monthly_earnings_ngn: number;
  available_balance_ngn: number;
  pending_withdrawal_ngn: number;
  referral_earnings_ngn: number;
  active_investments: number;
  completed_investments: number;
  exchange_rate: number;
  /** Withdrawals unlock when successful referrals >= min_referrals_required. */
  withdrawals_unlocked?: boolean;
  withdrawal_lock_message?: string;
  active_referrals?: number;
  min_referrals_required?: number;
  remaining_referrals?: number;
  /** Timestamp of the user's most recent withdrawal request (weekly withdrawal lock). */
  last_withdrawal_at?: string | null;
  investments: EarningsSummary[];
  referral_info?: ReferralInfo;
}

export interface ReferralInfo {
  referral_code: string;
  referral_link: string;
  total_referrals: number;
  active_referrals: number;
  referral_earnings_ngn: number;
  withdrawable_balance_ngn: number;
  minimum_target: number;
}

export interface RewardHistoryItem {
  id: string;
  amount_ngn: number;
  reward_date: string;
  business_day_number: number;
  status: string;
  created_at: string;
}

export interface PaymentHistoryItem {
  id: string;
  amount_ngn: number;
  amount_usd: number;
  provider: string;
  reference: string;
  status: string;
  paid_at?: string | null;
  created_at: string;
}

export interface WithdrawalHistoryItem {
  id: string;
  amount_ngn: number;
  fee_ngn: number;
  penalty_ngn: number;
  net_amount_ngn: number;
  withdrawal_type: 'earnings' | 'early' | 'normal';
  status: 'pending_review' | 'approved' | 'processing' | 'completed' | 'rejected';
  rejection_reason?: string | null;
  created_at: string;
  completed_at?: string | null;
}

export interface InvestmentNotification {
  id: string;
  type: string;
  title: string;
  message: string;
  data?: Record<string, unknown>;
  is_read: boolean;
  read_at?: string | null;
  created_at: string;
}

export interface InitPaymentResponse {
  authorization_url: string;
  reference: string;
  access_code?: string;
}

export interface AdminEarningsDashboard {
  total_invested_usd: number;
  total_invested_ngn: number;
  total_paid_out_ngn: number;
  active_investors: number;
  total_investors: number;
  pending_withdrawals: number;
  pending_payments: number;
  today_payout_ngn: number;
  total_referral_paid_ngn: number;
}
