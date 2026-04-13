-- +goose Up
-- +goose StatementBegin
ALTER TABLE `discount_campaigns`
    ADD COLUMN `campaign_type` varchar(50) NOT NULL DEFAULT 'product_percentage_discount' AFTER `name`,
    ADD COLUMN `buy_quantity` int unsigned NULL AFTER `discount_percentage`,
    ADD COLUMN `discount_amount` decimal(10,2) NULL AFTER `buy_quantity`,
    ADD COLUMN `reward_product_id` int unsigned NULL AFTER `discount_amount`,
    ADD COLUMN `reward_quantity` int unsigned NULL AFTER `reward_product_id`,
    ADD INDEX `idx_discount_campaigns_type` (`campaign_type`),
    ADD CONSTRAINT `fk_discount_campaigns_reward_product` FOREIGN KEY (`reward_product_id`) REFERENCES `products` (`id`) ON DELETE SET NULL;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE `transaction_items`
    ADD COLUMN `discount_amount` decimal(10,2) NOT NULL DEFAULT 0.00 AFTER `discount_percentage`,
    ADD COLUMN `campaign_type` varchar(50) NULL AFTER `discount_amount`,
    ADD COLUMN `is_free_item` tinyint(1) NOT NULL DEFAULT 0 AFTER `campaign_type`,
    ADD COLUMN `campaign_group_key` varchar(64) NULL AFTER `is_free_item`;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `transaction_items`
    DROP COLUMN `campaign_group_key`,
    DROP COLUMN `is_free_item`,
    DROP COLUMN `campaign_type`,
    DROP COLUMN `discount_amount`;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE `discount_campaigns`
    DROP FOREIGN KEY `fk_discount_campaigns_reward_product`,
    DROP INDEX `idx_discount_campaigns_type`,
    DROP COLUMN `reward_quantity`,
    DROP COLUMN `reward_product_id`,
    DROP COLUMN `discount_amount`,
    DROP COLUMN `buy_quantity`,
    DROP COLUMN `campaign_type`;
-- +goose StatementEnd
