-- +goose Up
-- +goose StatementBegin
CREATE TABLE `transactions` (
    `id` int unsigned NOT NULL AUTO_INCREMENT,
    `user` varchar(255) NOT NULL,
    `payment_method` varchar(50) NOT NULL,
    `discount` decimal(10,2) DEFAULT 0.00,
    `total_price` decimal(10,2) NOT NULL,
    `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE `transaction_items` (
    `id` int unsigned NOT NULL AUTO_INCREMENT,
    `transaction_id` int unsigned NOT NULL,
    `product_id` int unsigned NOT NULL,
    `quantity` int NOT NULL,
    `price` decimal(10,2) NOT NULL,
    `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_transaction_items_transaction_id` (`transaction_id`),
    KEY `idx_transaction_items_product_id` (`product_id`),
    CONSTRAINT `fk_transaction_items_transaction` FOREIGN KEY (`transaction_id`) REFERENCES `transactions` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_transaction_items_product` FOREIGN KEY (`product_id`) REFERENCES `products` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE `transaction_items`;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE `transactions`;
-- +goose StatementEnd 