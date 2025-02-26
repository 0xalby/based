-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS codes (
  `id` INTEGER NOT NULL PRIMARY KEY,
  `code` VARCHAR(6) NOT NULL DEFAULT "",
  `recovery` VARCHAR(255) NOT NULL DEFAULT "",
  `expiration` TIMESTAMP NOT NULL,
  `account` INTEGER NOT NULL, 
  FOREIGN KEY (account) REFERENCES accounts(id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE codes;
-- +goose StatementEnd