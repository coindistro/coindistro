-- 005_sessions_token_columns.sql
-- Fix: login/register session inserts fail with
--   "value too long for type character varying(255)"
-- Root cause: refresh tokens were stored as hex(raw JWT) which exceeds 255 chars.
-- Application now stores SHA-256 hex (64 chars); widen columns to TEXT for safety.

ALTER TABLE sessions
    ALTER COLUMN refresh_token_hash TYPE TEXT;

ALTER TABLE sessions
    ALTER COLUMN access_token_jti TYPE TEXT;

-- Devices fingerprint can also grow with complex client fingerprints.
ALTER TABLE devices
    ALTER COLUMN fingerprint TYPE TEXT;
