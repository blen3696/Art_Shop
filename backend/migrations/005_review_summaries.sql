-- ============================================================================
-- ArtShop — AI-generated review summaries
-- ============================================================================
-- Cached LLM summaries of a product's reviews ("Buyers praise the vivid colors,
-- a few mention the frame runs small"). Regenerated when the review count
-- changes or the summary is older than a week.
-- ============================================================================

CREATE TABLE IF NOT EXISTS product_review_summaries (
  product_id    UUID PRIMARY KEY REFERENCES products(id) ON DELETE CASCADE,
  summary       TEXT         NOT NULL,
  review_count  INTEGER      NOT NULL,
  generated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_review_summaries_generated_at
  ON product_review_summaries (generated_at);
