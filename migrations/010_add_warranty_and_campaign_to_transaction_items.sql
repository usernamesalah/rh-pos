-- +goose Up
-- +goose StatementBegin
ALTER TABLE `transaction_items`
ADD COLUMN `warranty_days` int NOT NULL DEFAULT 0,
ADD COLUMN `discount_percentage` decimal(5,2) NOT NULL DEFAULT 0.00,
ADD COLUMN `campaign_id` int unsigned DEFAULT NULL,
ADD CONSTRAINT `fk_transaction_items_campaign` FOREIGN KEY (`campaign_id`) REFERENCES `discount_campaigns` (`id`) ON DELETE SET NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `transaction_items`
DROP FOREIGN KEY `fk_transaction_items_campaign`,
DROP COLUMN `campaign_id`,
DROP COLUMN `discount_percentage`,
DROP COLUMN `warranty_days`;
-- +goose StatementEnd
