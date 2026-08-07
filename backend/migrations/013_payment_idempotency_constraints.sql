-- Payment idempotency constraints.
--
-- Webhooks and browser verification can arrive concurrently. These database
-- constraints make the provider reference the final uniqueness boundary,
-- rather than relying only on application-level existence checks.
--
-- If an existing database contains duplicates, index creation intentionally
-- fails so the records can be reconciled instead of silently deleting money.

CREATE UNIQUE INDEX IF NOT EXISTS idx_payment_transactions_provider_reference_unique
    ON payment_transactions(provider, reference);

CREATE UNIQUE INDEX IF NOT EXISTS idx_earnings_payment_transactions_provider_reference_unique
    ON earnings_payment_transactions(provider, reference);

-- Wallet references are the idempotency key used by wallet credit helpers.
CREATE UNIQUE INDEX IF NOT EXISTS idx_wallet_transactions_reference_unique
    ON wallet_transactions(reference);
