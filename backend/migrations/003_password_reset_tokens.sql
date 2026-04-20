-- ============================================================================
-- ArtShop — Password Reset Tokens
-- ============================================================================
-- Single-use, time-limited tokens for the forgot-password flow.
-- The raw token is sent in the reset email; only its SHA-256 hash is stored,
-- so a database leak cannot be used to take over accounts.
-- ============================================================================

CREATE TABLE IF NOT EXISTS password_reset_tokens (
  id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash  TEXT         NOT NULL UNIQUE,
  expires_at  TIMESTAMPTZ  NOT NULL,
  used_at     TIMESTAMPTZ,
  created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_user_id    ON password_reset_tokens (user_id);
CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_expires_at ON password_reset_tokens (expires_at);
