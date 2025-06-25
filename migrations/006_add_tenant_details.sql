-- +goose Up
-- +goose StatementBegin
ALTER TABLE `tenants`
ADD COLUMN `about` TEXT NULL,
ADD COLUMN `address` TEXT NULL,
ADD COLUMN `phone_number` VARCHAR(20) NULL,
ADD COLUMN `logo` VARCHAR(500) NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `tenants`
DROP COLUMN `about`,
DROP COLUMN `address`,
DROP COLUMN `phone_number`,
DROP COLUMN `logo`;
-- +goose StatementEnd 