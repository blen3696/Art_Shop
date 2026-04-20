-- ============================================================================
-- ArtShop — Product embeddings for semantic search & "similar products"
-- ============================================================================
-- Stores a 768-dim vector per product (Gemini text-embedding-004). Enables:
--   * Semantic search ("calm artwork" matches peaceful pieces without keyword overlap)
--   * "You might also like" (nearest neighbours of the current product)
--
-- Requires the pgvector extension (pre-installed on Supabase).
-- ============================================================================

CREATE EXTENSION IF NOT EXISTS vector;

ALTER TABLE products
  ADD COLUMN IF NOT EXISTS embedding     vector(768),
  ADD COLUMN IF NOT EXISTS embedded_at   TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS embedding_src TEXT;  -- the exact text we embedded, for change detection

-- HNSW index for fast cosine-similarity search. m/ef_construction use sensible defaults.
CREATE INDEX IF NOT EXISTS idx_products_embedding_hnsw
  ON products USING hnsw (embedding vector_cosine_ops);
