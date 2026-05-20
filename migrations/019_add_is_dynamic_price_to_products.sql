-- +goose Up
-- +goose StatementBegin
ALTER TABLE `products`
    ADD COLUMN `is_dynamic_price` TINYINT(1) NOT NULL DEFAULT 0 AFTER `stock`;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `products`
    DROP COLUMN `is_dynamic_price`;
-- +goose StatementEnd
