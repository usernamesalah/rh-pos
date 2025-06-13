-- +goose Up
-- +goose StatementBegin
CREATE TABLE `products` (
    `id` int unsigned NOT NULL AUTO_INCREMENT,
    `image` varchar(500) DEFAULT '',
    `name` varchar(255) NOT NULL,
    `sku` varchar(100) NOT NULL,
    `harga_modal` decimal(10,2) NOT NULL,
    `harga_jual` decimal(10,2) NOT NULL,
    `stock` int NOT NULL DEFAULT 0,
    `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_products_sku` (`sku`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO `products` (`name`, `sku`, `harga_modal`, `harga_jual`, `stock`, `tenant_id`) VALUES 
('Nasi Gudeg', 'NAS001', 8000.00, 12000.00, 50, 1),
('Ayam Bakar', 'AYM001', 15000.00, 20000.00, 30, 1),
('Es Teh Manis', 'MIN001', 2000.00, 5000.00, 100, 1),
('Sate Ayam', 'SAT001', 12000.00, 18000.00, 25, 1);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE `products`;
-- +goose StatementEnd 