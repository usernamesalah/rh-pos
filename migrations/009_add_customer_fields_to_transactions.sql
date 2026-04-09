-- +goose Up
-- +goose StatementBegin
ALTER TABLE `transactions`
ADD COLUMN `customer_name` varchar(255) DEFAULT NULL,
ADD COLUMN `customer_email` varchar(255) DEFAULT NULL,
ADD COLUMN `customer_phone` varchar(50) DEFAULT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `transactions`
DROP COLUMN `customer_phone`,
DROP COLUMN `customer_email`,
DROP COLUMN `customer_name`;
-- +goose StatementEnd
