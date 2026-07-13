-- ─── Coindistro Identity Service Schema ──────────────
-- Extends the base schema with identity-specific tables.

-- ─── Users (enhancement of base users table) ─────────
-- Note: This assumes the base users table from 001_initial_schema exists.
-- If rebuilding, use the enhanced version below.

CREATE TABLE IF NOT EXISTS identity_users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(50) UNIQUE,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(20) UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    display_name VARCHAR(100),
    avatar_url TEXT,
    country VARCHAR(3),
    timezone VARCHAR(50) DEFAULT 'UTC',
    locale VARCHAR(10) DEFAULT 'en',
    metadata JSONB DEFAULT '{}',

    -- Referral
    referral_code VARCHAR(20) UNIQUE NOT NULL,
    referred_by UUID REFERENCES identity_users(id),
    referral_level INT DEFAULT 0,

    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, active, suspended, banned

    -- Verification
    email_verified_at TIMESTAMPTZ,
    phone_verified_at TIMESTAMPTZ,
    email_verification_token VARCHAR(255),
    email_verification_sent_at TIMESTAMPTZ,
    password_reset_token VARCHAR(255),
    password_reset_sent_at TIMESTAMPTZ,
    password_reset_at TIMESTAMPTZ,

    -- Genesis / Founder
    is_genesis BOOLEAN DEFAULT false,
    genesis_number INT,
    genesis_date TIMESTAMPTZ,
    is_founder BOOLEAN DEFAULT false,
    founder_badge BOOLEAN DEFAULT false,

    -- Security
    failed_login_attempts INT DEFAULT 0,
    locked_until TIMESTAMPTZ,
    last_login_at TIMESTAMPTZ,
    last_login_ip VARCHAR(45),
    last_login_user_agent TEXT,

    -- Roles (Postgres array, synced with RBAC)
    roles TEXT[] DEFAULT '{user}',

    -- Soft delete
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Indexes
CREATE INDEX idx_identity_users_email ON identity_users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_identity_users_username ON identity_users(username) WHERE deleted_at IS NULL;
CREATE INDEX idx_identity_users_phone ON identity_users(phone) WHERE deleted_at IS NULL;
CREATE INDEX idx_identity_users_referral_code ON identity_users(referral_code);
CREATE INDEX idx_identity_users_referred_by ON identity_users(referred_by);
CREATE INDEX idx_identity_users_status ON identity_users(status);
CREATE INDEX idx_identity_users_genesis ON identity_users(is_genesis) WHERE is_genesis = true;
CREATE INDEX idx_identity_users_founder ON identity_users(is_founder) WHERE is_founder = true;

-- ─── Sessions ────────────────────────────────────────
CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES identity_users(id) ON DELETE CASCADE,
    refresh_token_hash VARCHAR(255) NOT NULL,
    access_token_jti VARCHAR(255),
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- active, expired, revoked, terminated
    ip_address VARCHAR(45),
    user_agent TEXT,
    browser VARCHAR(100),
    operating_system VARCHAR(100),
    device_name VARCHAR(100),
    device_type VARCHAR(50), -- desktop, mobile, tablet, unknown
    country VARCHAR(3),
    city VARCHAR(100),
    is_current BOOLEAN DEFAULT false,
    login_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_activity_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    terminated_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_refresh_token ON sessions(refresh_token_hash);
CREATE INDEX idx_sessions_status ON sessions(status);
CREATE INDEX idx_sessions_active ON sessions(user_id, status) WHERE status = 'active';
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

-- ─── Devices ─────────────────────────────────────────
CREATE TABLE IF NOT EXISTS devices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES identity_users(id) ON DELETE CASCADE,
    fingerprint VARCHAR(255),
    name VARCHAR(255),
    browser VARCHAR(100),
    operating_system VARCHAR(100),
    device_type VARCHAR(50),
    is_trusted BOOLEAN DEFAULT false,
    is_current BOOLEAN DEFAULT false,
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    first_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_devices_user_id ON devices(user_id);
CREATE INDEX idx_devices_fingerprint ON devices(fingerprint);

-- ─── Invitation Credits ──────────────────────────────
CREATE TABLE IF NOT EXISTS invitation_credits (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES identity_users(id) ON DELETE CASCADE,
    total_credits INT NOT NULL DEFAULT 0,
    used_credits INT NOT NULL DEFAULT 0,
    pending_credits INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id)
);

CREATE INDEX idx_invitation_credits_user_id ON invitation_credits(user_id);

-- ─── Invitations ─────────────────────────────────────
CREATE TABLE IF NOT EXISTS invitations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    inviter_id UUID NOT NULL REFERENCES identity_users(id) ON DELETE CASCADE,
    invitee_email VARCHAR(255) NOT NULL,
    invitee_id UUID REFERENCES identity_users(id),
    code VARCHAR(20) UNIQUE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, accepted, expired, revoked
    message TEXT,
    role VARCHAR(20) DEFAULT 'user',
    consumed_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_invitations_inviter_id ON invitations(inviter_id);
CREATE INDEX idx_invitations_invitee_email ON invitations(invitee_email);
CREATE INDEX idx_invitations_code ON invitations(code);
CREATE INDEX idx_invitations_status ON invitations(status);

-- ─── Referrals ───────────────────────────────────────
CREATE TABLE IF NOT EXISTS referrals (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    referrer_id UUID NOT NULL REFERENCES identity_users(id) ON DELETE CASCADE,
    referred_id UUID NOT NULL REFERENCES identity_users(id) ON DELETE CASCADE,
    referral_code VARCHAR(20) NOT NULL,
    level INT NOT NULL DEFAULT 1,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, active, converted, expired
    reward_amount DECIMAL(40, 8) DEFAULT 0,
    reward_currency VARCHAR(10) DEFAULT 'CDT',
    converted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(referred_id)
);

CREATE INDEX idx_referrals_referrer_id ON referrals(referrer_id);
CREATE INDEX idx_referrals_referred_id ON referrals(referred_id);
CREATE INDEX idx_referrals_level ON referrals(level);
CREATE INDEX idx_referrals_status ON referrals(status);

-- ─── Activity Log ────────────────────────────────────
CREATE TABLE IF NOT EXISTS activity_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES identity_users(id) ON DELETE CASCADE,
    action VARCHAR(100) NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    device_id UUID REFERENCES devices(id),
    details JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_activity_user_id ON activity_log(user_id);
CREATE INDEX idx_activity_action ON activity_log(action);
CREATE INDEX idx_activity_created_at ON activity_log(created_at);

-- ─── Genesis Configuration ───────────────────────────
CREATE TABLE IF NOT EXISTS genesis_config (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    max_genesis_members INT NOT NULL DEFAULT 10000,
    current_genesis_count INT NOT NULL DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Insert default genesis config
INSERT INTO genesis_config (id, max_genesis_members, current_genesis_count, is_active)
VALUES (uuid_generate_v4(), 10000, 0, true);

-- ─── Updated_at trigger function (if not exists) ─────
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply triggers
DO $$
DECLARE
    t text;
BEGIN
    FOR t IN
        SELECT table_name FROM information_schema.columns
        WHERE column_name = 'updated_at' AND table_schema = 'public'
        AND table_name IN ('identity_users', 'sessions', 'devices', 'invitation_credits', 'invitations', 'referrals')
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