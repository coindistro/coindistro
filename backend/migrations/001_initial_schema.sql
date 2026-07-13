-- ─── Coindistro Initial Schema ─────────────────────────
-- This migration creates the foundational tables for the Coindistro platform.
-- Future migrations will build upon this base.

-- ─── Extensions ───────────────────────────────────────
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ─── Users ────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(20) UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    display_name VARCHAR(100),
    avatar_url TEXT,
    roles TEXT[] DEFAULT '{user}',
    is_active BOOLEAN DEFAULT true,
    is_verified BOOLEAN DEFAULT false,
    email_verified_at TIMESTAMPTZ,
    phone_verified_at TIMESTAMPTZ,
    last_login_at TIMESTAMPTZ,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_phone ON users(phone) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_roles ON users USING GIN(roles);
CREATE INDEX idx_users_created_at ON users(created_at);

-- ─── Sessions / Refresh Tokens ────────────────────────
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    is_revoked BOOLEAN DEFAULT false,
    device_info TEXT,
    ip_address VARCHAR(45),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);

-- ─── KYC (Know Your Customer) ─────────────────────────
CREATE TABLE IF NOT EXISTS kyc_submissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, approved, rejected, needs_review
    id_type VARCHAR(50), -- passport, national_id, drivers_license
    id_number VARCHAR(100),
    id_front_url TEXT,
    id_back_url TEXT,
    selfie_url TEXT,
    proof_of_address_url TEXT,
    rejection_reason TEXT,
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reviewed_at TIMESTAMPTZ,
    reviewed_by UUID REFERENCES users(id),
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_kyc_user_id ON kyc_submissions(user_id);
CREATE INDEX idx_kyc_status ON kyc_submissions(status);

-- ─── Wallets ──────────────────────────────────────────
CREATE TABLE IF NOT EXISTS wallets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    currency VARCHAR(10) NOT NULL, -- BTC, ETH, USDT, NGN, etc.
    address VARCHAR(255),
    balance DECIMAL(40, 8) NOT NULL DEFAULT 0,
    locked_balance DECIMAL(40, 8) NOT NULL DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, currency)
);

CREATE INDEX idx_wallets_user_id ON wallets(user_id);
CREATE INDEX idx_wallets_currency ON wallets(currency);
CREATE INDEX idx_wallets_address ON wallets(address);

-- ─── Transactions ─────────────────────────────────────
CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    wallet_id UUID REFERENCES wallets(id),
    type VARCHAR(30) NOT NULL, -- deposit, withdrawal, transfer, trade, fee, reward
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, completed, failed, cancelled
    currency VARCHAR(10) NOT NULL,
    amount DECIMAL(40, 8) NOT NULL,
    fee DECIMAL(40, 8) DEFAULT 0,
    net_amount DECIMAL(40, 8) NOT NULL,
    balance_before DECIMAL(40, 8),
    balance_after DECIMAL(40, 8),
    reference VARCHAR(100) UNIQUE,
    description TEXT,
    metadata JSONB DEFAULT '{}',
    confirmed_at TIMESTAMPTZ,
    failed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_transactions_user_id ON transactions(user_id);
CREATE INDEX idx_transactions_wallet_id ON transactions(wallet_id);
CREATE INDEX idx_transactions_type ON transactions(type);
CREATE INDEX idx_transactions_status ON transactions(status);
CREATE INDEX idx_transactions_reference ON transactions(reference);
CREATE INDEX idx_transactions_created_at ON transactions(created_at);

-- ─── Merchant Accounts ────────────────────────────────
CREATE TABLE IF NOT EXISTS merchant_accounts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    business_name VARCHAR(255) NOT NULL,
    business_email VARCHAR(255),
    business_phone VARCHAR(20),
    business_address TEXT,
    business_type VARCHAR(50),
    registration_number VARCHAR(100),
    tax_id VARCHAR(100),
    website_url TEXT,
    webhook_url TEXT,
    callback_url TEXT,
    api_key VARCHAR(255) UNIQUE,
    api_secret VARCHAR(255),
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, active, suspended, closed
    fee_percentage DECIMAL(5, 2) DEFAULT 2.5,
    settlement_interval VARCHAR(20) DEFAULT 'daily', -- instant, daily, weekly, monthly
    is_verified BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_merchant_user_id ON merchant_accounts(user_id);
CREATE INDEX idx_merchant_api_key ON merchant_accounts(api_key);
CREATE INDEX idx_merchant_status ON merchant_accounts(status);

-- ─── Academy Courses ──────────────────────────────────
CREATE TABLE IF NOT EXISTS courses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    content TEXT,
    category VARCHAR(100),
    level VARCHAR(20) DEFAULT 'beginner', -- beginner, intermediate, advanced
    duration_minutes INT,
    price DECIMAL(20, 8) DEFAULT 0,
    currency VARCHAR(10) DEFAULT 'USD',
    thumbnail_url TEXT,
    video_url TEXT,
    is_published BOOLEAN DEFAULT false,
    author_id UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_courses_slug ON courses(slug);
CREATE INDEX idx_courses_category ON courses(category);
CREATE INDEX idx_courses_level ON courses(level);
CREATE INDEX idx_courses_published ON courses(is_published) WHERE is_published = true;

-- ─── Trading Signals ──────────────────────────────────
CREATE TABLE IF NOT EXISTS trading_signals (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    pair VARCHAR(20) NOT NULL,
    signal_type VARCHAR(10) NOT NULL, -- BUY, SELL
    entry_price DECIMAL(40, 8) NOT NULL,
    stop_loss DECIMAL(40, 8) NOT NULL,
    take_profit DECIMAL(40, 8) NOT NULL,
    risk_level VARCHAR(10) NOT NULL, -- low, medium, high
    confidence INT CHECK (confidence >= 0 AND confidence <= 100),
    analysis TEXT,
    is_active BOOLEAN DEFAULT true,
    is_profitable BOOLEAN,
    actual_exit_price DECIMAL(40, 8),
    actual_pnl DECIMAL(40, 8),
    created_by UUID REFERENCES users(id),
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_signals_pair ON trading_signals(pair);
CREATE INDEX idx_signals_active ON trading_signals(is_active) WHERE is_active = true;
CREATE INDEX idx_signals_created_at ON trading_signals(created_at);

-- ─── Trading Bots ─────────────────────────────────────
CREATE TABLE IF NOT EXISTS trading_bots (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    bot_type VARCHAR(50) NOT NULL, -- grid, dca, scalper, signal_executor, copy_trader
    status VARCHAR(20) NOT NULL DEFAULT 'stopped', -- running, stopped, paused, error
    config JSONB NOT NULL DEFAULT '{}',
    performance JSONB DEFAULT '{}',
    total_invested DECIMAL(40, 8) DEFAULT 0,
    current_value DECIMAL(40, 8) DEFAULT 0,
    total_pnl DECIMAL(40, 8) DEFAULT 0,
    total_trades INT DEFAULT 0,
    win_rate DECIMAL(5, 2) DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    started_at TIMESTAMPTZ,
    stopped_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_bots_user_id ON trading_bots(user_id);
CREATE INDEX idx_bots_type ON trading_bots(bot_type);
CREATE INDEX idx_bots_status ON trading_bots(status);

-- ─── Notifications ────────────────────────────────────
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL, -- trade, signal, system, security, promotion
    title VARCHAR(255) NOT NULL,
    body TEXT,
    data JSONB DEFAULT '{}',
    is_read BOOLEAN DEFAULT false,
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_unread ON notifications(user_id, is_read) WHERE is_read = false;
CREATE INDEX idx_notifications_created_at ON notifications(created_at);

-- ─── Audit Log ────────────────────────────────────────
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    resource VARCHAR(100),
    resource_id UUID,
    details JSONB DEFAULT '{}',
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);

-- ─── System Settings ──────────────────────────────────
CREATE TABLE IF NOT EXISTS system_settings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key VARCHAR(255) UNIQUE NOT NULL,
    value JSONB NOT NULL,
    description TEXT,
    is_public BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_settings_key ON system_settings(key);

-- ─── Functions & Triggers ─────────────────────────────
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply updated_at trigger to all tables with updated_at column
DO $$
DECLARE
    t text;
BEGIN
    FOR t IN
        SELECT table_name FROM information_schema.columns
        WHERE column_name = 'updated_at' AND table_schema = 'public'
    LOOP
        EXECUTE format('
            CREATE TRIGGER trigger_%s_updated_at
            BEFORE UPDATE ON %I
            FOR EACH ROW
            EXECUTE FUNCTION update_updated_at_column();
        ', t, t);
    END LOOP;
END;
$$ LANGUAGE plpgsql;