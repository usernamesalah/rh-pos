-- +goose Up
-- +goose StatementBegin
-- Migrate existing users with legacy role 'user' to 'cashier'
UPDATE `users` SET `role` = 'cashier' WHERE `role` = 'user';
-- +goose StatementEnd

-- +goose StatementBegin
-- Update column default to 'cashier'
ALTER TABLE `users` MODIFY COLUMN `role` varchar(50) NOT NULL DEFAULT 'cashier';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `users` MODIFY COLUMN `role` varchar(50) NOT NULL DEFAULT 'user';
-- +goose StatementEnd
