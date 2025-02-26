package services

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/charmbracelet/log"
)

type BlacklistService struct {
	DB *sql.DB
}

// Revokes jwt tokens
func (service *BlacklistService) RevokeToken(tokenID string, id int, expiration time.Time) error {
	rows, err := service.DB.Exec("INSERT INTO blacklist (token, account, expiration) VALUES (?, ?, ?)", tokenID, id, expiration)
	if err != nil {
		log.Error("failed to database insert", "err", err)
		return err
	}
	// Checking for affected rows
	affected, err := rows.RowsAffected()
	if err != nil {
		log.Error("failed to get affacted rows", "err", err)
		return err
	}
	if affected == 0 {
		log.Error("failed to revoke jwt token")
		return fmt.Errorf("no rows affected")
	}
	return err
}
