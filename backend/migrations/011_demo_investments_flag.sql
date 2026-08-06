-- ─── Demo investment flag ────────────────────────────────
-- Marks temporary development/demo investments so they can be
-- purged without touching real paid investments.

ALTER TABLE earnings_investments
    ADD COLUMN IF NOT EXISTS is_demo BOOLEAN NOT NULL DEFAULT false;

CREATE INDEX IF NOT EXISTS idx_earnings_investments_is_demo
    ON earnings_investments (is_demo)
    WHERE is_demo = true;

CREATE INDEX IF NOT EXISTS idx_earnings_investments_user_demo
    ON earnings_investments (user_id, is_demo);
