-- +goose Up
-- +goose StatementBegin
CREATE TABLE `discount_campaigns` (
    `id` int unsigned NOT NULL AUTO_INCREMENT,
    `name` varchar(255) NOT NULL,
    `discount_percentage` decimal(5,2) NOT NULL,
    `start_date` datetime NOT NULL,
    `end_date` datetime NOT NULL,
    `tenant_id` int unsigned NULL,
    `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    CONSTRAINT `fk_discount_campaigns_tenant` FOREIGN KEY (`tenant_id`) REFERENCES `tenants` (`id`) ON DELETE CASCADE,
    INDEX `idx_discount_campaigns_tenant_id` (`tenant_id`),
    INDEX `idx_discount_campaigns_dates` (`tenant_id`, `start_date`, `end_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE `discount_campaign_products` (
    `id` int unsigned NOT NULL AUTO_INCREMENT,
    `campaign_id` int unsigned NOT NULL,
    `product_id` int unsigned NOT NULL,
    `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    CONSTRAINT `fk_dcp_campaign` FOREIGN KEY (`campaign_id`) REFERENCES `discount_campaigns` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_dcp_product` FOREIGN KEY (`product_id`) REFERENCES `products` (`id`) ON DELETE CASCADE,
    UNIQUE KEY `uq_campaign_product` (`campaign_id`, `product_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS `discount_campaign_products`;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS `discount_campaigns`;
-- +goose StatementEnd
