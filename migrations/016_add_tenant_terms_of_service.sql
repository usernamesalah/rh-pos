-- +goose Up
-- +goose StatementBegin
ALTER TABLE `tenants` ADD COLUMN `terms_of_service` TEXT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `tenants` DROP COLUMN `terms_of_service`;
-- +goose StatementEnd
