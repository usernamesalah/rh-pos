-- +goose Up
-- +goose StatementBegin
CREATE TABLE `categories` (
    `id` int unsigned NOT NULL AUTO_INCREMENT,
    `name` varchar(255) NOT NULL,
    `tenant_id` int unsigned NULL,
    `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    CONSTRAINT `fk_categories_tenant` FOREIGN KEY (`tenant_id`) REFERENCES `tenants` (`id`) ON DELETE CASCADE,
    INDEX `idx_categories_tenant_id` (`tenant_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE `products`
    ADD COLUMN `category_id` int unsigned NULL AFTER `tenant_id`,
    ADD CONSTRAINT `fk_products_category` FOREIGN KEY (`category_id`) REFERENCES `categories` (`id`) ON DELETE SET NULL,
    ADD INDEX `idx_products_category_id` (`category_id`);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `products`
    DROP FOREIGN KEY `fk_products_category`,
    DROP INDEX `idx_products_category_id`,
    DROP COLUMN `category_id`;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS `categories`;
-- +goose StatementEnd
