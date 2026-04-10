-- +goose Up
-- +goose StatementBegin
CREATE TABLE `audit_logs` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT,
    `tenant_id` bigint unsigned NOT NULL,
    `user_id` bigint unsigned NOT NULL,
    `entity_type` varchar(50) NOT NULL,
    `entity_id` bigint unsigned NOT NULL,
    `action` varchar(20) NOT NULL,
    `before_state` json NULL,
    `after_state` json NULL,
    `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    INDEX `idx_audit_logs_entity` (`entity_type`, `entity_id`),
    INDEX `idx_audit_logs_tenant` (`tenant_id`),
    INDEX `idx_audit_logs_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS `audit_logs`;
-- +goose StatementEnd
