-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS blacklist (
    `token` VARCHAR(36) NOT NULL PRIMARY KEY, -- Unique identifier for the JWT token
	`expiration` TIMESTAMP NOT NULL,
	`account` INTEGER NOT NULL, 
	 FOREIGN KEY (account) REFERENCES accounts(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE blacklist;
-- +goose StatementEnd