DROP INDEX IF EXISTS idx_payments_created;
DROP INDEX IF EXISTS idx_payments_status;
DROP INDEX IF EXISTS idx_payments_user_id;
DROP INDEX IF EXISTS idx_payments_order_id;
DROP TABLE IF EXISTS payments CASCADE;