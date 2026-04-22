-- ============================================================================
-- ArtShop — Payments table (Chapa integration)
-- ============================================================================
-- One Order can have MANY Payment attempts (failed card → retry Telebirr, etc.).
-- We never mutate a finished attempt; a new attempt inserts a new row. This
-- gives us a complete audit trail, which matters for dispute resolution.
-- ============================================================================

CREATE TABLE IF NOT EXISTS payments (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id        UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,

    -- Our unique reference sent to Chapa as tx_ref. We send the payment's UUID
    -- as tx_ref so collisions are impossible and we can look up a payment
    -- directly from any webhook/verify response.
    tx_ref          VARCHAR(64) UNIQUE NOT NULL,

    -- Chapa's own transaction reference — not known until payment completes.
    provider_ref    VARCHAR(128),

    provider        VARCHAR(30) NOT NULL DEFAULT 'chapa',
    amount          DECIMAL(12,2) NOT NULL,
    currency        VARCHAR(10) NOT NULL DEFAULT 'ETB',

    -- State machine: pending → success | failed | cancelled.
    -- We never go back from a terminal state (success/failed/cancelled).
    status          VARCHAR(20) NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending', 'success', 'failed', 'cancelled')),

    -- URL returned by Chapa for the hosted checkout page. We send the user here.
    checkout_url    TEXT,

    -- Raw verify payload stored as JSONB for debugging / audit. Kept out of
    -- application logic — treat as opaque unless debugging.
    raw_response    JSONB,

    failure_reason  TEXT,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at    TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_payments_order_id ON payments(order_id);
CREATE INDEX IF NOT EXISTS idx_payments_user_id  ON payments(user_id);
CREATE INDEX IF NOT EXISTS idx_payments_status   ON payments(status);
CREATE INDEX IF NOT EXISTS idx_payments_created  ON payments(created_at DESC);