-- ─── Investor Earnings Dashboard & Investment Flow ──────────
-- Naira-based daily earnings investment system.
-- All monetary values use DECIMAL(40,2) for precision.
-- All monetary calculations use decimal precision.

-- ─── Exchange Rates ─────────────────────────────────────────
CREATE TABLE IF NOT EXISTS exchange_rates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    usd_to_ngn DECIMAL(20, 2) NOT NULL DEFAULT 1400.00,
    set_by UUID REFERENCES identity_users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_exchange_rates_created_at ON exchange_rates(created_at DESC);

-- Insert default exchange rate (idempotent)
INSERT INTO exchange_rates (id, usd_to_ngn)
VALUES ('22222222-2222-2222-2222-222222222222'::uuid, 1400.00)
ON CONFLICT (id) DO NOTHING;

-- ─── Investment Settings (Admin Configurable) ──────────────
CREATE TABLE IF NOT EXISTS investment_settings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    minimum_investment_usd DECIMAL(20, 2) NOT NULL DEFAULT 30.00,
    daily_reward_ngn DECIMAL(20, 2) NOT NULL DEFAULT 650.00,
    max_business_days INT NOT NULL DEFAULT 20,
    roi_percent DECIMAL(10, 4) NOT NULL DEFAULT 30.0000,
    referral_percent DECIMAL(10, 4) NOT NULL DEFAULT 10.0000,
    min_referrals_for_payout INT NOT NULL DEFAULT 5,
    early_withdrawal_penalty_percent DECIMAL(10, 4) NOT NULL DEFAULT 15.0000,
    early_withdrawal_fee_percent DECIMAL(10, 4) NOT NULL DEFAULT 5.0000,
    withdrawal_processing_hours INT NOT NULL DEFAULT 24,
    enabled BOOLEAN NOT NULL DEFAULT true,
    updated_by UUID REFERENCES identity_users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Insert default settings (idempotent)
INSERT INTO investment_settings (id, minimum_investment_usd, daily_reward_ngn, max_business_days, roi_percent, referral_percent, min_referrals_for_payout, early_withdrawal_penalty_percent, early_withdrawal_fee_percent)
VALUES ('33333333-3333-3333-3333-333333333333'::uuid, 30.00, 650.00, 20, 30.0000, 10.0000, 5, 15.0000, 5.0000)
ON CONFLICT (id) DO NOTHING;

-- ─── Withdrawal Fee Tiers (Admin Configurable) ─────────────
CREATE TABLE IF NOT EXISTS withdrawal_fee_tiers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    min_amount DECIMAL(20, 2) NOT NULL,
    max_amount DECIMAL(20, 2) NOT NULL,
    fee_percent DECIMAL(10, 4) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Insert default fee tiers (idempotent)
INSERT INTO withdrawal_fee_tiers (id, min_amount, max_amount, fee_percent) VALUES
    ('aaaaaaaa-0001-0001-0001-aaaaaaaa0001'::uuid, 0, 50000, 3.0000),
    ('aaaaaaaa-0001-0001-0001-aaaaaaaa0002'::uuid, 50000, 200000, 2.0000),
    ('aaaaaaaa-0001-0001-0001-aaaaaaaa0003'::uuid, 200000, 999999999, 1.0000)
ON CONFLICT (id) DO NOTHING;

-- ─── Earnings Investments ──────────────────────────────────
-- Investment statuses: pending_payment, active, completed, paused, cancelled, early_withdrawal
CREATE TABLE IF NOT EXISTS earnings_investments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES identity_users(id) ON DELETE CASCADE,
    amount_usd DECIMAL(20, 2) NOT NULL,
    amount_ngn DECIMAL(20, 2) NOT NULL,
    exchange_rate DECIMAL(20, 2) NOT NULL DEFAULT 1400.00,
    payment_provider VARCHAR(20) NOT NULL,
    payment_reference VARCHAR(255) NOT NULL,
    payment_status VARCHAR(20) NOT NULL DEFAULT 'pending',
    daily_reward_ngn DECIMAL(20, 2) NOT NULL DEFAULT 650.00,
    max_business_days INT NOT NULL DEFAULT 20,
    paid_business_days INT NOT NULL DEFAULT 0,
    total_earned_ngn DECIMAL(20, 2) NOT NULL DEFAULT 0,
    total_pending_ngn DECIMAL(20, 2) NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'pending_payment',
    maturity_date DATE,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    paused_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    early_withdrawal_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_earnings_investments_user_id ON earnings_investments(user_id);
CREATE INDEX IF NOT EXISTS idx_earnings_investments_status ON earnings_investments(status);
CREATE INDEX IF NOT EXISTS idx_earnings_investments_payment_reference ON earnings_investments(payment_reference);
CREATE UNIQUE INDEX IF NOT EXISTS idx_earnings_investments_payment_ref_provider ON earnings_investments(payment_provider, payment_reference);
CREATE INDEX IF NOT EXISTS idx_earnings_investments_maturity_date ON earnings_investments(maturity_date) WHERE status = 'active';

-- ─── Daily Rewards ─────────────────────────────────────────
-- Record of every daily reward credited to an investor
CREATE TABLE IF NOT EXISTS investment_rewards (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    investment_id UUID NOT NULL REFERENCES earnings_investments(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES identity_users(id) ON DELETE CASCADE,
    amount_ngn DECIMAL(20, 2) NOT NULL,
    reward_date DATE NOT NULL,
    business_day_number INT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'credited',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_investment_rewards_investment_id ON investment_rewards(investment_id);
CREATE INDEX IF NOT EXISTS idx_investment_rewards_user_id ON investment_rewards(user_id);
CREATE INDEX IF NOT EXISTS idx_investment_rewards_reward_date ON investment_rewards(reward_date);
CREATE UNIQUE INDEX IF NOT EXISTS idx_investment_rewards_unique_date ON investment_rewards(investment_id, reward_date);

-- ─── Withdrawals ──────────────────────────────────────────
CREATE TABLE IF NOT EXISTS withdrawals (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES identity_users(id) ON DELETE CASCADE,
    investment_id UUID REFERENCES earnings_investments(id) ON DELETE SET NULL,
    -- Null for earnings withdrawal, set for principal+profits withdrawal
    amount_ngn DECIMAL(20, 2) NOT NULL,
    fee_ngn DECIMAL(20, 2) NOT NULL DEFAULT 0,
    penalty_ngn DECIMAL(20, 2) NOT NULL DEFAULT 0,
    net_amount_ngn DECIMAL(20, 2) NOT NULL,
    withdrawal_type VARCHAR(20) NOT NULL, -- 'earnings', 'early', 'normal'
    status VARCHAR(20) NOT NULL DEFAULT 'pending_review',
    -- Statuses: pending_review, approved, processing, completed, rejected
    -- For early withdrawal: automatically deducts penalty + fee
    -- For normal withdrawal: capital + profits - fee
    reviewed_by UUID REFERENCES identity_users(id),
    reviewed_at TIMESTAMPTZ,
    processed_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    rejection_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_withdrawals_user_id ON withdrawals(user_id);
CREATE INDEX IF NOT EXISTS idx_withdrawals_status ON withdrawals(status);
CREATE INDEX IF NOT EXISTS idx_withdrawals_investment_id ON withdrawals(investment_id);

-- ─── Payment Transactions (Earnings) ──────────────────────
CREATE TABLE IF NOT EXISTS earnings_payment_transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES identity_users(id) ON DELETE CASCADE,
    investment_id UUID REFERENCES earnings_investments(id) ON DELETE SET NULL,
    provider VARCHAR(20) NOT NULL, -- 'paystack', 'flutterwave'
    reference VARCHAR(255) NOT NULL,
    type VARCHAR(20) NOT NULL DEFAULT 'investment', -- 'investment', 'withdrawal'
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    amount_ngn DECIMAL(20, 2) NOT NULL,
    amount_usd DECIMAL(20, 2) NOT NULL DEFAULT 0,
    exchange_rate DECIMAL(20, 2) NOT NULL DEFAULT 1400.00,
    response JSONB DEFAULT '{}',
    paid_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_earnings_payment_transactions_user_id ON earnings_payment_transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_earnings_payment_transactions_reference ON earnings_payment_transactions(reference);
CREATE INDEX IF NOT EXISTS idx_earnings_payment_transactions_status ON earnings_payment_transactions(status);

-- ─── Referral Commissions ─────────────────────────────────
CREATE TABLE IF NOT EXISTS referral_commissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    referrer_id UUID NOT NULL REFERENCES identity_users(id) ON DELETE CASCADE,
    referred_id UUID NOT NULL REFERENCES identity_users(id) ON DELETE CASCADE,
    investment_id UUID NOT NULL REFERENCES earnings_investments(id) ON DELETE CASCADE,
    amount_usd DECIMAL(20, 2) NOT NULL,
    amount_ngn DECIMAL(20, 2) NOT NULL,
    percent DECIMAL(10, 4) NOT NULL DEFAULT 10.0000,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- 'pending', 'paid', 'cancelled'
    paid_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(investment_id, referrer_id)
);

CREATE INDEX IF NOT EXISTS idx_referral_commissions_referrer_id ON referral_commissions(referrer_id);
CREATE INDEX IF NOT EXISTS idx_referral_commissions_referred_id ON referral_commissions(referred_id);
CREATE INDEX IF NOT EXISTS idx_referral_commissions_status ON referral_commissions(status);

-- ─── User Notifications ───────────────────────────────────
CREATE TABLE IF NOT EXISTS investment_notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES identity_users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    -- Types: payment_confirmed, investment_activated, daily_reward_credited,
    -- investment_matured, withdrawal_approved, withdrawal_rejected,
    -- referral_commission_received, early_withdrawal_warning
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    data JSONB DEFAULT '{}',
    is_read BOOLEAN NOT NULL DEFAULT false,
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_investment_notifications_user_id ON investment_notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_investment_notifications_is_read ON investment_notifications(is_read);
CREATE INDEX IF NOT EXISTS idx_investment_notifications_type ON investment_notifications(type);

-- ─── Triggers for updated_at ──────────────────────────────
DO $$
DECLARE
    t text;
BEGIN
    FOR t IN
        SELECT table_name FROM information_schema.columns
        WHERE column_name = 'updated_at' AND table_schema = 'public'
        AND table_name IN ('exchange_rates', 'investment_settings', 'withdrawal_fee_tiers',
                          'earnings_investments', 'withdrawals')
    LOOP
        EXECUTE format('
            DROP TRIGGER IF EXISTS trigger_%s_updated_at ON %I;
            CREATE TRIGGER trigger_%s_updated_at
            BEFORE UPDATE ON %I
            FOR EACH ROW
            EXECUTE FUNCTION update_updated_at_column();
        ', t, t, t, t);
    END LOOP;
END;
$$ LANGUAGE plpgsql;