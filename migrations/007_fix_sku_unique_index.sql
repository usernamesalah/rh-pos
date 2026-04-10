-- +goose Up
-- +goose StatementBegin
ALTER TABLE products DROP INDEX `idx_products_sku`;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE products ADD UNIQUE INDEX `idx_tenant_sku` (`tenant_id`, `sku`);
-- +goose StatementEnd
