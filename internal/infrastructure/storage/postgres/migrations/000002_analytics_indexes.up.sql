-- Migration: Analytics optimization indexes

-- Index for user + date range queries (main analytics filter)
CREATE INDEX IF NOT EXISTS idx_purchases_user_date 
ON purchases(user_id, purchase_date);

-- Index for expense lookups by purchase (for JOINs)
CREATE INDEX IF NOT EXISTS idx_expenses_purchase 
ON expenses(purchase_id);

-- Index for category-based aggregation queries
CREATE INDEX IF NOT EXISTS idx_expenses_category 
ON expenses(category_id);

-- Index for organization-based filtering
CREATE INDEX IF NOT EXISTS idx_purchases_org 
ON purchases(user_id, organisation_name);

-- Composite index optimized for analytics queries
-- Covers: purchase filtering + expense aggregation by category
CREATE INDEX IF NOT EXISTS idx_expenses_analytics 
ON expenses(purchase_id, category_id, total_price_minor);
