-- Fix: the verification-token repository queries `verification_tokens`, but
-- 001_init.sql created the table as `email_verifications`. Rename and align
-- columns so email-verification flows actually work. No data exists in this
-- table in any deployed environment (the flow was broken before this fix).
ALTER TABLE IF EXISTS email_verifications RENAME TO verification_tokens;

ALTER TABLE verification_tokens ADD COLUMN IF NOT EXISTS email VARCHAR(320) NOT NULL DEFAULT '';
ALTER TABLE verification_tokens ADD COLUMN IF NOT EXISTS consumed_at TIMESTAMPTZ;
ALTER TABLE verification_tokens DROP COLUMN IF EXISTS verified_at;
ALTER TABLE verification_tokens ALTER COLUMN email DROP DEFAULT;
