-- +goose Up
-- +goose StatementBegin
CREATE TABLE `users` (
    `id` int unsigned NOT NULL AUTO_INCREMENT,
    `username` varchar(255) NOT NULL,
    `password` varchar(255) NOT NULL,
    `role` varchar(50) NOT NULL DEFAULT 'user',
    `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_users_username` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE `users`;
-- +goose StatementEnd 