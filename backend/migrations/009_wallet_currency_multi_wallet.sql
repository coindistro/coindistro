-- ─── Wallets: restore multi-currency support ─────────────
-- CoinDistro is a MULTI-CURRENCY digital bank & exchange.
-- One user owns MANY wallets, one per currency (NGN, USD, USDT, BTC, ETH, CDT...).
--
-- Migration 008 incorrectly replaced UNIQUE(user_id, currency) with
-- UNIQUE(user_id), forcing a single wallet per user. This migration:
--   1. Adds a NOT NULL `currency` column (defaulting existing rows safely).
--   2. Drops the incorrect UNIQUE(user_id) constraint.
--   3. Restores UNIQUE(user_id, currency) so a user can hold many wallets
--      but never two wallets of the same currency.
--   4. Backfills NULL currencies without destroying any balances.
--
-- Backfill reasoning:
--   * The legacy 001 `wallets` table is a fiat wallet (has `currency` already).
--   * The 004/008 `wallets` table (CDT ledger) has no `currency` column.
--   * If a user has exactly one wallet, we label it 'CDT' (the internal ledger
--     currency used by investments/earnings). If a legacy fiat wallet already
--     has a currency, we keep it. We never destroy balances.

-- 1. Add currency column if missing (idempotent).
ALTER TABLE wallets ADD COLUMN IF NOT EXISTS currency VARCHAR(10);

-- 2. Backfill NULL currencies.
--    - Prefer an existing non-null currency (legacy fiat wallets).
--    - Otherwise default to 'CDT' (internal ledger currency).
UPDATE wallets
SET currency = COALESCE(NULLIF(TRIM(currency), ''), 'CDT')
WHERE currency IS NULL OR TRIM(currency) = '';

-- 3. Enforce NOT NULL now that all rows are backfilled.
ALTER TABLE wallets ALTER COLUMN currency SET NOT NULL;

-- 4. Drop the incorrect single-wallet constraint from migration 008.
ALTER TABLE wallets DROP CONSTRAINT IF EXISTS wallets_user_id_unique;

-- 5. Restore multi-currency uniqueness: one wallet per (user, currency).
--    Drop any pre-existing per-currency constraint first, then recreate.
ALTER TABLE wallets DROP CONSTRAINT IF EXISTS wallets_user_id_currency_key;
ALTER TABLE wallets ADD CONSTRAINT wallets_user_id_currency_key UNIQUE (user_id, currency);