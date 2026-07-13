-- ─── Coindistro Earn Service Schema ───────────────────
-- Product catalog, participations, rewards, campaigns.

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ─── Earn Products ────────────────────────────────────
CREATE TABLE IF NOT EXISTS earn_products (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    category VARCHAR(50) NOT NULL, -- flexible, fixed, stablecoin, ai_smart, signal_vault, launchpool, learn_earn, referral
    supported_assets TEXT[] NOT NULL DEFAULT '{}',
    duration_days INT, -- null for flexible
    capacity_total DECIMAL(40, 8),
    capacity_used DECIMAL(40, 8) NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'draft', -- draft, active, paused, closed, archived
    risk_level VARCHAR(20) NOT NULL DEFAULT 'medium', -- low, medium, high
    min_allocation DECIMAL(40, 8) NOT NULL DEFAULT 0,
    max_allocation DECIMAL(40, 8),
    reward_model VARCHAR(50) NOT NULL DEFAULT 'flexible', -- flexible, fixed, promotional, educational, referral
    reward_apr DECIMAL(10, 4) DEFAULT 0, -- display / estimated APR (not financial execution)
    eligibility JSONB NOT NULL DEFAULT '{}',
    rules JSONB NOT NULL DEFAULT '{}',
    strategy_profiles TEXT[] DEFAULT '{}', -- for AI Smart Earn
    featured BOOLEAN NOT NULL DEFAULT false,
    metadata JSONB NOT NULL DEFAULT '{}',
    starts_at TIMESTAMPTZ,
    ends_at TIMESTAMPTZ,
    created_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_earn_products_category ON earn_products(category);
CREATE INDEX idx_earn_products_status ON earn_products(status);
CREATE INDEX idx_earn_products_featured ON earn_products(featured) WHERE featured = true;
CREATE INDEX idx_earn_products_slug ON earn_products(slug);

-- ─── Participations ───────────────────────────────────
CREATE TABLE IF NOT EXISTS earn_participations (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    product_id UUID NOT NULL REFERENCES earn_products(id),
    asset VARCHAR(20) NOT NULL,
    allocated_amount DECIMAL(40, 8) NOT NULL DEFAULT 0,
    current_balance DECIMAL(40, 8) NOT NULL DEFAULT 0,
    estimated_rewards DECIMAL(40, 8) NOT NULL DEFAULT 0,
    accrued_rewards DECIMAL(40, 8) NOT NULL DEFAULT 0,
    lifetime_rewards DECIMAL(40, 8) NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- active, locked, completed, exited, cancelled
    strategy_profile VARCHAR(30), -- conservative, balanced, growth, aggressive
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    lock_start_at TIMESTAMPTZ,
    lock_end_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    exited_at TIMESTAMPTZ,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_earn_part_user ON earn_participations(user_id);
CREATE INDEX idx_earn_part_product ON earn_participations(product_id);
CREATE INDEX idx_earn_part_status ON earn_participations(status);
CREATE INDEX idx_earn_part_user_status ON earn_participations(user_id, status);

-- ─── Reward ledger ────────────────────────────────────
CREATE TABLE IF NOT EXISTS earn_rewards (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    product_id UUID NOT NULL REFERENCES earn_products(id),
    participation_id UUID REFERENCES earn_participations(id),
    asset VARCHAR(20) NOT NULL,
    amount DECIMAL(40, 8) NOT NULL,
    reward_type VARCHAR(50) NOT NULL, -- daily, fixed_maturity, promotional, educational, referral, milestone
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, calculated, granted, failed
    description TEXT,
    period_start TIMESTAMPTZ,
    period_end TIMESTAMPTZ,
    granted_at TIMESTAMPTZ,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_earn_rewards_user ON earn_rewards(user_id);
CREATE INDEX idx_earn_rewards_product ON earn_rewards(product_id);
CREATE INDEX idx_earn_rewards_participation ON earn_rewards(participation_id);
CREATE INDEX idx_earn_rewards_status ON earn_rewards(status);
CREATE INDEX idx_earn_rewards_created ON earn_rewards(created_at);

-- ─── Transaction history ──────────────────────────────
CREATE TABLE IF NOT EXISTS earn_transactions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    product_id UUID REFERENCES earn_products(id),
    participation_id UUID REFERENCES earn_participations(id),
    type VARCHAR(40) NOT NULL, -- join, add_funds, withdraw, exit, reward, lock, unlock
    asset VARCHAR(20) NOT NULL,
    amount DECIMAL(40, 8) NOT NULL,
    balance_after DECIMAL(40, 8),
    status VARCHAR(20) NOT NULL DEFAULT 'completed',
    reference VARCHAR(100),
    description TEXT,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_earn_tx_user ON earn_transactions(user_id);
CREATE INDEX idx_earn_tx_product ON earn_transactions(product_id);
CREATE INDEX idx_earn_tx_type ON earn_transactions(type);
CREATE INDEX idx_earn_tx_created ON earn_transactions(created_at);

-- ─── Launchpool campaigns ─────────────────────────────
CREATE TABLE IF NOT EXISTS earn_launchpool_campaigns (
    id UUID PRIMARY KEY,
    product_id UUID REFERENCES earn_products(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    supported_assets TEXT[] NOT NULL DEFAULT '{}',
    window_start TIMESTAMPTZ NOT NULL,
    window_end TIMESTAMPTZ NOT NULL,
    allocation_rules JSONB NOT NULL DEFAULT '{}',
    reward_distribution JSONB NOT NULL DEFAULT '{}',
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_earn_launchpool_status ON earn_launchpool_campaigns(status);

-- ─── Learn & Earn campaigns ───────────────────────────
CREATE TABLE IF NOT EXISTS earn_learn_campaigns (
    id UUID PRIMARY KEY,
    product_id UUID REFERENCES earn_products(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    academy_course_id UUID, -- future Academy integration
    reward_asset VARCHAR(20) NOT NULL DEFAULT 'USDT',
    reward_amount DECIMAL(40, 8) NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    starts_at TIMESTAMPTZ,
    ends_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS earn_learn_completions (
    id UUID PRIMARY KEY,
    campaign_id UUID NOT NULL REFERENCES earn_learn_campaigns(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    completed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reward_eligible BOOLEAN NOT NULL DEFAULT true,
    reward_granted BOOLEAN NOT NULL DEFAULT false,
    reward_id UUID,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(campaign_id, user_id)
);

CREATE INDEX idx_earn_learn_comp_user ON earn_learn_completions(user_id);

-- ─── Referral reward milestones ───────────────────────
CREATE TABLE IF NOT EXISTS earn_referral_milestones (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    required_referrals INT NOT NULL DEFAULT 1,
    reward_asset VARCHAR(20) NOT NULL DEFAULT 'CDT',
    reward_amount DECIMAL(40, 8) NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS earn_referral_reward_claims (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    milestone_id UUID REFERENCES earn_referral_milestones(id),
    amount DECIMAL(40, 8) NOT NULL DEFAULT 0,
    asset VARCHAR(20) NOT NULL DEFAULT 'CDT',
    status VARCHAR(20) NOT NULL DEFAULT 'eligible', -- eligible, granted, rejected
    granted_at TIMESTAMPTZ,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_earn_ref_claims_user ON earn_referral_reward_claims(user_id);

-- ─── Performance snapshots (scheduler) ────────────────
CREATE TABLE IF NOT EXISTS earn_performance_snapshots (
    id UUID PRIMARY KEY,
    product_id UUID REFERENCES earn_products(id),
    snapshot_date DATE NOT NULL,
    participants INT NOT NULL DEFAULT 0,
    total_allocated DECIMAL(40, 8) NOT NULL DEFAULT 0,
    total_rewards DECIMAL(40, 8) NOT NULL DEFAULT 0,
    capacity_used_pct DECIMAL(10, 4) DEFAULT 0,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(product_id, snapshot_date)
);

-- ─── updated_at triggers ──────────────────────────────
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$
DECLARE
    t text;
BEGIN
    FOR t IN
        SELECT unnest(ARRAY[
            'earn_products', 'earn_participations', 'earn_rewards',
            'earn_launchpool_campaigns', 'earn_learn_campaigns', 'earn_referral_milestones'
        ])
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
