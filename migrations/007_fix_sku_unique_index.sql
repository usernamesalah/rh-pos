-- Migration 007: Replace global SKU unique index with per-tenant composite index
-- Pre-check before running:
-- SELECT sku, COUNT(DISTINCT tenant_id) FROM products GROUP BY sku HAVING COUNT(DISTINCT tenant_id) > 1;
-- If this returns rows, resolve SKU conflicts first.

-- Drop the old global unique index (GORM default name)
ALTER TABLE products DROP INDEX `sku`;

-- Add composite unique index: one SKU per tenant
ALTER TABLE products ADD UNIQUE INDEX `idx_tenant_sku` (`tenant_id`, `sku`);
