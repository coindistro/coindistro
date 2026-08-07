-- Payment lifecycle hardening (idempotent and backward compatible).
-- Existing paid_at is retained for API compatibility; completed_at is now the
-- canonical lifecycle timestamp. Historical rows are deliberately not rewritten.

ALTER TABLE payment_transactions
    ADD COLUMN IF NOT EXISTS initialized_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS processing_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS verified_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS completed_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS failed_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS cancelled_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS expired_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS refunded_at TIMESTAMPTZ;

ALTER TABLE earnings_payment_transactions
    ADD COLUMN IF NOT EXISTS initialized_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS processing_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS verified_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS completed_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS failed_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS cancelled_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS expired_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS refunded_at TIMESTAMPTZ;

CREATE TABLE IF NOT EXISTS payment_audit_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    payment_scope VARCHAR(20) NOT NULL,
    payment_id UUID NOT NULL,
    provider VARCHAR(20) NOT NULL,
    reference VARCHAR(255) NOT NULL,
    user_id UUID NOT NULL REFERENCES identity_users(id) ON DELETE RESTRICT,
    investment_id UUID,
    event VARCHAR(64) NOT NULL,
    wallet_before DECIMAL(40, 8),
    wallet_after DECIMAL(40, 8),
    amount DECIMAL(40, 8) NOT NULL,
    currency VARCHAR(10) NOT NULL,
    request_id VARCHAR(255),
    ip INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_payment_audit_events_payment ON payment_audit_events(payment_scope, payment_id, created_at);
CREATE INDEX IF NOT EXISTS idx_payment_audit_events_reference ON payment_audit_events(provider, reference);

CREATE TABLE IF NOT EXISTS payment_webhook_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    payment_scope VARCHAR(20) NOT NULL,
    provider VARCHAR(20) NOT NULL,
    event_id VARCHAR(255) NOT NULL,
    reference VARCHAR(255),
    signature TEXT NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}',
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ,
    UNIQUE(payment_scope, provider, event_id)
);
CREATE INDEX IF NOT EXISTS idx_payment_webhook_events_reference ON payment_webhook_events(provider, reference);

CREATE TABLE IF NOT EXISTS payment_retry_queue (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    payment_scope VARCHAR(20) NOT NULL,
    payment_id UUID NOT NULL,
    provider VARCHAR(20) NOT NULL,
    reference VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}',
    failure_reason TEXT NOT NULL,
    attempts INTEGER NOT NULL DEFAULT 0,
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '1 minute',
    resolved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(payment_scope, payment_id)
);
CREATE INDEX IF NOT EXISTS idx_payment_retry_queue_ready ON payment_retry_queue(next_attempt_at) WHERE resolved_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_payment_transactions_created_at ON payment_transactions(created_at);
CREATE INDEX IF NOT EXISTS idx_earnings_payment_transactions_provider_reference ON earnings_payment_transactions(provider, reference);
CREATE INDEX IF NOT EXISTS idx_earnings_payment_transactions_created_at ON earnings_payment_transactions(created_at);
