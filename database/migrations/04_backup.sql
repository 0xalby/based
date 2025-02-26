-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS backup (
    `id` INTEGER PRIMARY KEY,
    `hash` VARCHAR(255) NOT NULL, -- Hashed TOTP backup code
    -- `used` BOOLEAN NOT NULL DEFAULT 0, -- Indicates if the code has been used
	`account` INTEGER NOT NULL, 
	FOREIGN KEY (account) REFERENCES accounts(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE backup;
-- +goose StatementEnd