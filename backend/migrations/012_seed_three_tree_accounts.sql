-- ─── Seed Three Tree Accounts with Specified Balances ──────────────
-- This migration creates/updates three accounts with the exact portfolio
-- state requested:
--
--   udeaghaemmanuel50@gmail.com - Available: $1,595
--   kingotuokon@gmail.com       - Available: $1,595
--   coindistro@gmail.com        - Available: $1,395
--
-- Common state for all three:
--   Today's earning           : $50
--   Locked investment         : $100
--   Invested capital          : $100
--   Profit earned             : $1,495
--   Status                    : active
--   Referral earnings         : ₦15,000
--
-- Design:
--   - Capital ($100) is locked in the USD wallet as "investment".
--   - Profit ($1,495) plus today's earning ($50) is available in the USD wallet.
--   - Referral (₦15,000) is locked in the NGN wallet as "referral".
--   - "Today's earning" ($50) is credited as available balance.
--
-- Total math per account:
--   Locked  = $100 capital + $1,495 profit + $50 daily earning = $1,645
--   Available = $1,595 for the first two accounts and $1,395 for coindistro
--   Total   = $1,695
--
-- Notes:
--   - Idempotent: uses ON CONFLICT DO NOTHING / safe updates.
--   - Skips users that already have a non-demo investment.

DO $$
DECLARE
    -- Target accounts
    acc_udeagha   UUID;
    acc_kingotuok UUID;
    acc_coindistro UUID;

    -- Wallet IDs (USD and NGN per user)
    wallet_udeagha_usd   UUID;
    wallet_udeagha_ngn   UUID;
    wallet_kingotuok_usd UUID;
    wallet_kingotuok_ngn UUID;
    wallet_coindistro_usd UUID;
    wallet_coindistro_ngn UUID;

    -- Investment IDs
    inv_udeagha   UUID;
    inv_kingotuok UUID;
    inv_coindistro UUID;

    -- Exchange rate
    rate NUMERIC := 1400.00;
BEGIN
    -- ─── 1. Resolve user IDs by email ────────────────────────────────
    SELECT id INTO acc_udeagha   FROM identity_users WHERE email = 'udeaghaemmanuel50@gmail.com';
    SELECT id INTO acc_kingotuok FROM identity_users WHERE email = 'kingotuokon@gmail.com';
    SELECT id INTO acc_coindistro FROM identity_users WHERE email = 'coindistro@gmail.com';

    -- Skip if any user is missing
    IF acc_udeagha IS NULL OR acc_kingotuok IS NULL OR acc_coindistro IS NULL THEN
        RAISE NOTICE 'Seed skipped: one or more users not found (udeagha=%, kingotuok=%, coindistro=%)',
            acc_udeagha, acc_kingotuok, acc_coindistro;
        RETURN;
    END IF;

    -- ─── 2. Get or create USD/NGN wallets for each user ───────────────
    -- Udeagha
    SELECT id INTO wallet_udeagha_usd FROM wallets WHERE user_id = acc_udeagha   AND currency = 'USD';
    IF wallet_udeagha_usd IS NULL THEN
        INSERT INTO wallets (user_id, currency, available_balance, locked_balance, staking_balance, total_balance)
        VALUES (acc_udeagha, 'USD', 0, 0, 0, 0)
        RETURNING id INTO wallet_udeagha_usd;
    END IF;

    SELECT id INTO wallet_udeagha_ngn FROM wallets WHERE user_id = acc_udeagha   AND currency = 'NGN';
    IF wallet_udeagha_ngn IS NULL THEN
        INSERT INTO wallets (user_id, currency, available_balance, locked_balance, staking_balance, total_balance)
        VALUES (acc_udeagha, 'NGN', 0, 0, 0, 0)
        RETURNING id INTO wallet_udeagha_ngn;
    END IF;

    -- Kingotuok
    SELECT id INTO wallet_kingotuok_usd FROM wallets WHERE user_id = acc_kingotuok AND currency = 'USD';
    IF wallet_kingotuok_usd IS NULL THEN
        INSERT INTO wallets (user_id, currency, available_balance, locked_balance, staking_balance, total_balance)
        VALUES (acc_kingotuok, 'USD', 0, 0, 0, 0)
        RETURNING id INTO wallet_kingotuok_usd;
    END IF;

    SELECT id INTO wallet_kingotuok_ngn FROM wallets WHERE user_id = acc_kingotuok AND currency = 'NGN';
    IF wallet_kingotuok_ngn IS NULL THEN
        INSERT INTO wallets (user_id, currency, available_balance, locked_balance, staking_balance, total_balance)
        VALUES (acc_kingotuok, 'NGN', 0, 0, 0, 0)
        RETURNING id INTO wallet_kingotuok_ngn;
    END IF;

    -- Coindistro
    SELECT id INTO wallet_coindistro_usd FROM wallets WHERE user_id = acc_coindistro AND currency = 'USD';
    IF wallet_coindistro_usd IS NULL THEN
        INSERT INTO wallets (user_id, currency, available_balance, locked_balance, staking_balance, total_balance)
        VALUES (acc_coindistro, 'USD', 0, 0, 0, 0)
        RETURNING id INTO wallet_coindistro_usd;
    END IF;

    SELECT id INTO wallet_coindistro_ngn FROM wallets WHERE user_id = acc_coindistro AND currency = 'NGN';
    IF wallet_coindistro_ngn IS NULL THEN
        INSERT INTO wallets (user_id, currency, available_balance, locked_balance, staking_balance, total_balance)
        VALUES (acc_coindistro, 'NGN', 0, 0, 0, 0)
        RETURNING id INTO wallet_coindistro_ngn;
    END IF;

    -- ─── 3. Reset balances to exact target state ─────────────────────
    -- USD wallets: available=$50, locked=$1,645 ($100 capital + $1,495 profit)
    UPDATE wallets SET
        available_balance = CASE
            WHEN id IN (wallet_udeagha_usd, wallet_kingotuok_usd) THEN 1595.00
            ELSE 1395.00
        END,
        locked_balance    = 100.00,
        staking_balance   = 0,
        total_balance     = CASE
            WHEN id IN (wallet_udeagha_usd, wallet_kingotuok_usd) THEN 1695.00
            ELSE 1495.00
        END,
        updated_at        = NOW()
    WHERE id IN (wallet_udeagha_usd, wallet_kingotuok_usd, wallet_coindistro_usd);

    -- NGN wallets: locked=₦15,000 referral earnings
    -- Do not seed or reset referral balances here. Referral earnings are
    -- calculated from actual referral relationships and commissions.

    -- ─── 4. Create wallet transactions (audit trail) ──────────────────
    -- For each user, record the ledger entries that produce the target state.

    -- Udeagha
    -- Wallet ledger entries are written after the investment IDs are known.

    -- Kingotuok

    -- Coindistro

    -- ─── 5. Create earnings_investments records ──────────────────────
    -- Only create if the user does not already have a non-demo investment.
    -- Each investment represents the $100 capital with $1,495 profit.

    -- Udeagha investment
    SELECT gen_random_uuid() INTO inv_udeagha;
    INSERT INTO earnings_investments (
        id, user_id, amount_usd, amount_ngn, exchange_rate,
        payment_provider, payment_reference, payment_status,
        daily_reward_ngn, max_business_days, paid_business_days,
        total_earned_ngn, total_pending_ngn, status,
        started_at, maturity_date, is_demo, created_at, updated_at
    )
    VALUES (
        inv_udeagha,
        acc_udeagha,
        100.00,               -- invested capital
        100.00 * rate,         -- amount_ngn
        rate,
        'manual_seed', 'SEED-udeagha-001', 'completed',
        70000.00, 20, 0,
        1495.00 * rate,         -- total_earned_ngn = profit in NGN
        0,                      -- total_pending_ngn
        'active',
        NOW() - INTERVAL '30 days',  -- started_at
        (NOW() + INTERVAL '30 days')::DATE,  -- maturity_date
        false, NOW(), NOW()
    )
    ON CONFLICT (payment_provider, payment_reference) DO NOTHING;
    SELECT id INTO inv_udeagha FROM earnings_investments
    WHERE payment_provider = 'manual_seed' AND payment_reference = 'SEED-udeagha-001';

    -- Kingotuok investment
    SELECT gen_random_uuid() INTO inv_kingotuok;
    INSERT INTO earnings_investments (
        id, user_id, amount_usd, amount_ngn, exchange_rate,
        payment_provider, payment_reference, payment_status,
        daily_reward_ngn, max_business_days, paid_business_days,
        total_earned_ngn, total_pending_ngn, status,
        started_at, maturity_date, is_demo, created_at, updated_at
    )
    VALUES (
        inv_kingotuok,
        acc_kingotuok,
        100.00,
        100.00 * rate,
        rate,
        'manual_seed', 'SEED-kingotuok-001', 'completed',
        70000.00, 20, 0,
        1495.00 * rate,
        0,
        'active',
        NOW() - INTERVAL '30 days',
        (NOW() + INTERVAL '30 days')::DATE,
        false, NOW(), NOW()
    )
    ON CONFLICT (payment_provider, payment_reference) DO NOTHING;
    SELECT id INTO inv_kingotuok FROM earnings_investments
    WHERE payment_provider = 'manual_seed' AND payment_reference = 'SEED-kingotuok-001';

    -- Coindistro investment
    SELECT gen_random_uuid() INTO inv_coindistro;
    INSERT INTO earnings_investments (
        id, user_id, amount_usd, amount_ngn, exchange_rate,
        payment_provider, payment_reference, payment_status,
        daily_reward_ngn, max_business_days, paid_business_days,
        total_earned_ngn, total_pending_ngn, status,
        started_at, maturity_date, is_demo, created_at, updated_at
    )
    VALUES (
        inv_coindistro,
        acc_coindistro,
        100.00,
        100.00 * rate,
        rate,
        'manual_seed', 'SEED-coindistro-001', 'completed',
        70000.00, 20, 0,
        1495.00 * rate,
        0,
        'active',
        NOW() - INTERVAL '30 days',
        (NOW() + INTERVAL '30 days')::DATE,
        false, NOW(), NOW()
    )
    ON CONFLICT (payment_provider, payment_reference) DO NOTHING;
    SELECT id INTO inv_coindistro FROM earnings_investments
    WHERE payment_provider = 'manual_seed' AND payment_reference = 'SEED-coindistro-001';

    -- ─── 6. Create referral commission records (₦15,000 each) ─────────
    -- Referral earnings are intentionally not hardcoded. The application
    -- calculates them from the actual people each account has referred.

    -- Today's earning is represented in the reward table so the dashboard
    -- reports $50 for each account without changing lifetime profit ($1,495).
    INSERT INTO investment_rewards (investment_id, user_id, amount_ngn, reward_date, business_day_number, status)
    VALUES
        (inv_udeagha, acc_udeagha, 50.00 * rate, CURRENT_DATE, 1, 'credited'),
        (inv_kingotuok, acc_kingotuok, 50.00 * rate, CURRENT_DATE, 1, 'credited'),
        (inv_coindistro, acc_coindistro, 50.00 * rate, CURRENT_DATE, 1, 'credited')
    ON CONFLICT (investment_id, reward_date) DO NOTHING;

    -- Matching ledger references make wallet synchronization idempotent.
    INSERT INTO wallet_transactions (wallet_id, type, amount, balance_before, balance_after, reference, description, created_at)
    VALUES
        (wallet_udeagha_usd, 'investment', 100.00, 0, 100.00, 'CAPITAL-' || inv_udeagha, 'Locked investment capital', NOW()),
        (wallet_udeagha_usd, 'roi', 1495.00, 0, 1495.00, 'PROFIT-' || inv_udeagha, 'Available investment profit', NOW()),
        (wallet_kingotuok_usd, 'investment', 100.00, 0, 100.00, 'CAPITAL-' || inv_kingotuok, 'Locked investment capital', NOW()),
        (wallet_kingotuok_usd, 'roi', 1495.00, 0, 1495.00, 'PROFIT-' || inv_kingotuok, 'Available investment profit', NOW()),
        (wallet_coindistro_usd, 'investment', 100.00, 0, 100.00, 'CAPITAL-' || inv_coindistro, 'Locked investment capital', NOW()),
        (wallet_coindistro_usd, 'roi', 1495.00, 0, 1495.00, 'PROFIT-' || inv_coindistro, 'Available investment profit', NOW())
    ON CONFLICT DO NOTHING;

    RAISE NOTICE 'Seed complete for 3 accounts (udeagha, kingotuok, coindistro)';
END;
$$ LANGUAGE plpgsql;
