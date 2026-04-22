-- Remove the embedding column from products. Does NOT drop the pgvector
-- extension because other migrations or future tables may depend on it.
ALTER TABLE products DROP COLUMN IF EXISTS embedding;
DROP INDEX IF EXISTS idx_products_embedding;