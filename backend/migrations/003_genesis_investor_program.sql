-- ─── Genesis Investor Program Schema ──────────────────
-- Internal ledger for CDT investments before TGE.
-- All balances will migrate 1:1 to blockchain CDT later.

-- ─── Investment Plans ─────────────────────────────────
CREATE TABLE IF NOT EXISTS investment_plans (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    description TEXT DEFAULT '',
    minimum_amount DECIMAL(40, 2) NOT NULL,
    maximum_amount DECIMAL(40, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'NGN',
    roi_percent DECIMAL(10, 4) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_investment_plans_enabled ON investment_plans(enabled);

-- ─── Investments ──────────────────────────────────────
CREATE TABLE IF NOT EXISTS investments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES identity_users(id) ON DELETE CASCADE,
    plan_id UUID NOT NULL REFERENCES investment_plans(id),
    payment_provider VARCHAR(20) NOT NULL,
    payment_reference VARCHAR(255) NOT NULL,
    payment_status VARCHAR(20) NOT NULL DEFAULT 'pending',
    amount_paid DECIMAL(40, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'NGN',
    exchange_rate DECIMAL(20, 8) NOT NULL DEFAULT 1,
    cdt_price DECIMAL(20, 8) NOT NULL,
    allocated_cdt DECIMAL(40, 8) NOT NULL,
    roi_percent DECIMAL(10, 4) NOT NULL,
    roi_cdt DECIMAL(40, 8) NOT NULL DEFAULT 0,
    lock_period_days INT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    started_at TIMESTAMPTZ,
    matures_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_investments_user_id ON investments(user_id);
CREATE INDEX idx_investments_plan_id ON investments(plan_id);
CREATE INDEX idx_investments_status ON investments(status);
CREATE INDEX idx_investments_payment_reference ON investments(payment_reference);
CREATE INDEX idx_investments_matures_at ON investments(matures_at) WHERE status = 'active';
CREATE UNIQUE INDEX idx_investments_payment_ref_provider ON investments(payment_provider, payment_reference);

-- ─── Payment Transactions ─────────────────────────────
CREATE TABLE IF NOT EXISTS payment_transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES identity_users(id) ON DELETE CASCADE,
    provider VARCHAR(20) NOT NULL,
    reference VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    amount DECIMAL(40, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'NGN',
    response JSONB DEFAULT '{}',
    paid_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payment_transactions_user_id ON payment_transactions(user_id);
CREATE INDEX idx_payment_transactions_reference ON payment_transactions(reference);
CREATE INDEX idx_payment_transactions_provider_ref ON payment_transactions(provider, reference);
CREATE INDEX idx_payment_transactions_status ON payment_transactions(status);

-- ─── Wallets (Internal CDT Ledger) ────────────────────
CREATE TABLE IF NOT EXISTS wallets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES identity_users(id) ON DELETE CASCADE,
    available_balance DECIMAL(40, 8) NOT NULL DEFAULT 0,
    locked_balance DECIMAL(40, 8) NOT NULL DEFAULT 0,
    staking_balance DECIMAL(40, 8) NOT NULL DEFAULT 0,
    total_balance DECIMAL(40, 8) NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id)
);

CREATE INDEX idx_wallets_user_id ON wallets(user_id);

-- ─── Wallet Transactions ──────────────────────────────
CREATE TABLE IF NOT EXISTS wallet_transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL,
    amount DECIMAL(40, 8) NOT NULL,
    balance_before DECIMAL(40, 8) NOT NULL,
    balance_after DECIMAL(40, 8) NOT NULL,
    reference VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_wallet_transactions_wallet_id ON wallet_transactions(wallet_id);
CREATE INDEX idx_wallet_transactions_type ON wallet_transactions(type);
CREATE INDEX idx_wallet_transactions_created_at ON wallet_transactions(created_at);

-- ─── CDT Pricing ──────────────────────────────────────
CREATE TABLE IF NOT EXISTS cdt_pricing (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    price_ngn DECIMAL(20, 8) NOT NULL,
    set_by UUID REFERENCES identity_users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Insert default pricing (₦10 per CDT)
INSERT INTO cdt_pricing (id, price_ngn)
VALUES (uuid_generate_v4(), 10.00);

-- ─── Webhook Processing Log ───────────────────────────
CREATE TABLE IF NOT EXISTS webhook_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    provider VARCHAR(20) NOT NULL,
    event_id VARCHAR(255) NOT NULL,
    reference VARCHAR(255),
    status VARCHAR(20) NOT NULL DEFAULT 'received',
    payload JSONB DEFAULT '{}',
    processed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(provider, event_id)
);

CREATE INDEX idx_webhook_events_provider ON webhook_events(provider);
CREATE INDEX idx_webhook_events_reference ON webhook_events(reference);

-- ─── Seed Default Investment Plans ────────────────────
INSERT INTO investment_plans (id, name, description, minimum_amount, maximum_amount, currency, roi_percent, enabled) VALUES
    (uuid_generate_v4(), 'Starter', 'Perfect for beginners. Start your CDT investment journey with a small amount.', 1000, 50000, 'NGN', 5.0000, true),
    (uuid_generate_v4(), 'Builder', 'Build your CDT portfolio with medium-term investment.', 50000, 500000, 'NGN', 10.0000, true),
    (uuid_generate_v4(), 'Pro', 'Professional grade investment for serious investors.', 500000, 5000000, 'NGN', 15.0000, true),
    (uuid_generate_v4(), 'Whale', 'Maximum returns for maximum investment. Exclusive high-tier plan.', 5000000, 50000000, 'NGN', 25.0000, true);

-- ─── Triggers ─────────────────────────────────────────
DO $$
DECLARE
    t text;
BEGIN
    FOR t IN
        SELECT table_name FROM information_schema.columns
        WHERE column_name = 'updated_at' AND table_schema = 'public'
        AND table_name IN ('investment_plans', 'investments', 'cdt_pricing')
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