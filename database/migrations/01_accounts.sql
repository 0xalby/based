-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE IF NOT EXISTS accounts (
  `id` INTEGER NOT NULL PRIMARY KEY,
  -- `customer` VARCHAR(255) NOT NULL UNIQUE, -- Might be a Stripe customer id
  `email` VARCHAR(255) NOT NULL UNIQUE,
  `pending` VARCHAR(255) NOT NULL DEFAULT "",
  `password` VARCHAR(255) NOT NULL,
  `verified` BOOLEAN NOT NULL DEFAULT 0, -- Verified true/false
  `2fa` BOOLEAN NOT NULL DEFAULT 0, -- 2FA disabled/enabled
  `updated` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `created` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE accounts;
-- +goose StatementEnd
