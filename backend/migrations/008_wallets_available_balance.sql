-- ─── Wallets: add available_balance / staking_balance / total_balance ───
-- Root cause of SQLSTATE 42703 "column available_balance does not exist":
--   001_initial_schema.sql created `wallets` with `balance` (not `available_balance`).
--   004_genesis_investor_program.sql used `CREATE TABLE IF NOT EXISTS wallets`
--   with `available_balance`, but because the table already existed (from 001),
--   the `IF NOT EXISTS` clause silently skipped the whole CREATE TABLE, so the
--   new columns were never added. The investments/earnings stores query
--   `available_balance`, `staking_balance`, and `total_balance`, which do not
--   exist on the 001-shaped table.
--
-- This migration adds the missing columns idempotently (ADD COLUMN IF NOT EXISTS)
-- so both fresh databases (where 004 already created them) and existing databases
-- (where 001 created the table and 004 was a no-op) converge to the same schema.
-- `balance` from 001 is kept for backward compatibility and mirrored into
-- `available_balance` on first run so no funds are lost.

-- available_balance: spendable CDT balance (queried by investments/earnings stores)
ALTER TABLE wallets ADD COLUMN IF NOT EXISTS available_balance DECIMAL(40, 8) NOT NULL DEFAULT 0;

-- staking_balance: CDT locked in staking positions
ALTER TABLE wallets ADD COLUMN IF NOT EXISTS staking_balance DECIMAL(40, 8) NOT NULL DEFAULT 0;

-- total_balance: available + locked + staking (denormalized total)
ALTER TABLE wallets ADD COLUMN IF NOT EXISTS total_balance DECIMAL(40, 8) NOT NULL DEFAULT 0;

-- Backfill available_balance / total_balance from the legacy `balance` column
-- for any rows that still have 0 in the new columns (only applies to 001-shaped DBs).
-- `balance` is the legacy 001 column; `locked_balance` exists in both schemas.
UPDATE wallets
SET available_balance = COALESCE(balance, 0),
    total_balance = COALESCE(balance, 0) + COALESCE(locked_balance, 0) + COALESCE(staking_balance, 0)
WHERE available_balance = 0
  AND total_balance = 0
  AND COALESCE(balance, 0) > 0;

-- The investments store relies on UNIQUE(user_id) (from 004) rather than
-- UNIQUE(user_id, currency) (from 001). The 004 CREATE TABLE was skipped on
-- existing DBs, so its UNIQUE(user_id) constraint was never added. Add it here
-- so GetOrCreateWallet's "one wallet per user" assumption holds.
-- Drop the legacy per-currency uniqueness first if present, then add per-user.
ALTER TABLE wallets DROP CONSTRAINT IF EXISTS wallets_user_id_currency_key;
ALTER TABLE wallets ADD CONSTRAINT wallets_user_id_unique UNIQUE (user_id);