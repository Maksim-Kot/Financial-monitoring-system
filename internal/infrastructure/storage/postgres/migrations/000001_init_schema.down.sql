DROP INDEX IF EXISTS idx_expenses_purchase_id;
DROP INDEX IF EXISTS idx_organisations_user_updated_at;
DROP INDEX IF EXISTS idx_purchases_user_date;

DROP TRIGGER IF EXISTS set_expenses_updated_at ON expenses;
DROP TRIGGER IF EXISTS set_purchases_updated_at ON purchases;
DROP TRIGGER IF EXISTS set_organisations_updated_at ON organisations;

DROP TABLE IF EXISTS expenses;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS purchases;
DROP TABLE IF EXISTS organisations;

DROP FUNCTION IF EXISTS set_updated_at_column_to_now();