-- +goose Up
-- +goose StatementBegin
CREATE TABLE `tenants` (
    `id` int unsigned NOT NULL AUTO_INCREMENT,
    `name` varchar(255) NOT NULL,
    `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE `users`
ADD COLUMN `tenant_id` int unsigned NULL,
ADD CONSTRAINT `fk_users_tenant` FOREIGN KEY (`tenant_id`) REFERENCES `tenants` (`id`);
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE `products`
ADD COLUMN `tenant_id` int unsigned NULL,
ADD CONSTRAINT `fk_products_tenant` FOREIGN KEY (`tenant_id`) REFERENCES `tenants` (`id`);
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE `transactions`
ADD COLUMN `tenant_id` int unsigned NULL,
ADD CONSTRAINT `fk_transactions_tenant` FOREIGN KEY (`tenant_id`) REFERENCES `tenants` (`id`);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX `idx_users_tenant_id` ON `users` (`tenant_id`);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX `idx_products_tenant_id` ON `products` (`tenant_id`);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX `idx_transactions_tenant_id` ON `transactions` (`tenant_id`);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX `idx_transactions_tenant_id` ON `transactions`;
-- +goose StatementEnd

-- +goose StatementBegin
DROP INDEX `idx_products_tenant_id` ON `products`;
-- +goose StatementEnd

-- +goose StatementBegin
DROP INDEX `idx_users_tenant_id` ON `users`;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE `transactions` DROP FOREIGN KEY `fk_transactions_tenant`;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE `products` DROP FOREIGN KEY `fk_products_tenant`;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE `users` DROP FOREIGN KEY `fk_users_tenant`;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE `transactions` DROP COLUMN `tenant_id`;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE `products` DROP COLUMN `tenant_id`;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE `users` DROP COLUMN `tenant_id`;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE `tenants`;
-- +goose StatementEnd 