-- +goose Up
-- +goose StatementBegin
ALTER TABLE `transactions` ADD COLUMN `user_id` bigint unsigned NULL;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE `transactions` ADD INDEX `idx_transactions_user_id` (`user_id`);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `transactions` DROP INDEX `idx_transactions_user_id`;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE `transactions` DROP COLUMN `user_id`;
-- +goose StatementEnd