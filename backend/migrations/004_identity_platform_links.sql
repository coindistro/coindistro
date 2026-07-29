-- ─── Link platform tables to identity_users ───────────
-- Auth lives on identity_users; wallets/merchant/kyc originally referenced
-- legacy users(id). Drop those FKs so seed + future services can use identity IDs.

ALTER TABLE wallets DROP CONSTRAINT IF EXISTS wallets_user_id_fkey;
ALTER TABLE merchant_accounts DROP CONSTRAINT IF EXISTS merchant_accounts_user_id_fkey;
ALTER TABLE kyc_submissions DROP CONSTRAINT IF EXISTS kyc_submissions_user_id_fkey;
ALTER TABLE transactions DROP CONSTRAINT IF EXISTS transactions_user_id_fkey;

-- Helpful uniqueness for merchant (one primary account per identity user)
CREATE UNIQUE INDEX IF NOT EXISTS idx_merchant_user_id_unique ON merchant_accounts(user_id);

-- Wallet uniqueness already exists: UNIQUE(user_id, currency)
