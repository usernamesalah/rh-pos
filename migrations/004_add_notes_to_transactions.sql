-- +goose Up
-- +goose StatementBegin
ALTER TABLE `transactions` ADD COLUMN `notes` TEXT NULL AFTER `total_price`;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `transactions` DROP COLUMN `notes`;
-- +goose StatementEnd 