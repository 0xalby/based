-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS verification (
  `id` INTEGER NOT NULL PRIMARY KEY,
  `code` VARCHAR(255) NOT NULL,
  `account` INTEGER NOT NULL, 
  `expiration` TIMESTAMP NOT NULL,
  FOREIGN KEY (account) REFERENCES accounts(id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE verification;
-- +goose StatementEnd