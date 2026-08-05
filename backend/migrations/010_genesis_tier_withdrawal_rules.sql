-- ─── Genesis Tier & Weekly Withdrawal Rules ──────────────
-- Adds the smallest investment package (Genesis) and the
-- one-withdrawal-every-7-days rule. Existing users keep working.
-- Do NOT modify already-applied migrations; this file is additive only.

-- 1) Update investment_settings defaults for the Genesis tier.
ALTER TABLE investment_settings
    ALTER COLUMN minimum_investment_usd SET DEFAULT 10.00;
ALTER TABLE investment_settings
    ALTER COLUMN daily_reward_ngn SET DEFAULT 126.00;
ALTER TABLE investment_settings
    ALTER COLUMN roi_percent SET DEFAULT 18.0000;
ALTER TABLE investment_settings
    ALTER COLUMN withdrawal_processing_hours SET DEFAULT 24;

-- 2) Add weekly withdrawal interval column (default 7 days).
ALTER TABLE investment_settings
    ADD COLUMN IF NOT EXISTS withdrawal_interval_days INT NOT NULL DEFAULT 7;

-- 3) Update existing settings rows to Genesis defaults + weekly lock.
--    Safe to re-run: always converges on the new product policy.
UPDATE investment_settings
SET
    minimum_investment_usd = 10.00,
    daily_reward_ngn = 126.00,
    max_business_days = 20,
    roi_percent = 18.0000,
    withdrawal_processing_hours = 24,
    withdrawal_interval_days = COALESCE(NULLIF(withdrawal_interval_days, 0), 7),
    updated_at = NOW();

-- 4) Seed the legacy CDT investment_plans table with Genesis (idempotent).
--    Unique on name (idx_investment_plans_name) so re-runs are safe.
INSERT INTO investment_plans (id, name, description, minimum_amount, maximum_amount, currency, roi_percent, enabled) VALUES
    (gen_random_uuid(), 'Genesis', 'Start your investment journey with the smallest tier. 18% ROI over 20 business days.', 10, 999, 'USD', 18.0000, true)
ON CONFLICT (name) DO UPDATE SET
    description = EXCLUDED.description,
    minimum_amount = EXCLUDED.minimum_amount,
    maximum_amount = EXCLUDED.maximum_amount,
    currency = EXCLUDED.currency,
    roi_percent = EXCLUDED.roi_percent,
    enabled = EXCLUDED.enabled,
    updated_at = NOW();
