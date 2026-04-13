-- +goose Up
-- +goose StatementBegin
ALTER TABLE `products` MODIFY COLUMN `sku` VARCHAR(191) NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `products` MODIFY COLUMN `sku` VARCHAR(191) NOT NULL DEFAULT '';
-- +goose StatementEnd
