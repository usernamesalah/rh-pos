-- +goose Up
-- +goose StatementBegin
ALTER TABLE `products`
  MODIFY COLUMN `harga_modal` DECIMAL(15,2) NULL,
  MODIFY COLUMN `harga_jual`  DECIMAL(15,2) NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `products`
  MODIFY COLUMN `harga_modal` DECIMAL(15,2) NOT NULL DEFAULT 0,
  MODIFY COLUMN `harga_jual`  DECIMAL(15,2) NOT NULL DEFAULT 0;
-- +goose StatementEnd
