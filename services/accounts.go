package services

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/0xalby/base/types"
	"github.com/charmbracelet/log"
)

type AccountsService struct {
	DB *sql.DB
}

func (service *AccountsService) CreateAccount(account *types.Account) error {
	// Executing on the database
	rows, err := service.DB.Exec("INSERT INTO accounts (email, password) VALUES (?,?)", account.Email, account.Password)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			if strings.Contains(err.Error(), "email") {
				return fmt.Errorf("email already used")
			}
		}
		log.Error("failed to database insert", "err", err)
		return err
	}
	// Checking for affected rows
	affected, err := rows.RowsAffected()
	if err != nil {
		log.Error("failed to get affacted rows")
		return err
	}
	if affected == 0 {
		log.Error("failed to create account")
		return fmt.Errorf("no rows affected")
	}
	return nil
}

func (service *AccountsService) GetAccountByID(id int) (*types.Account, error) {
	// Querying the database
	rows, err := service.DB.Query("SELECT * FROM accounts WHERE id = ?", id)
	if err != nil {
		log.Error("failed to database query")
		return nil, err
	}
	defer rows.Close()
	// Scanning the rows
	var account *types.Account
	for rows.Next() {
		account, err = scanAccounts(rows)
		if err != nil {
			return nil, err
		}
	}
	if err = rows.Err(); err != nil {
		log.Error("failed interating rows")
		return nil, err
	}
	if account == nil || account.ID == 0 {
		log.Error("account doesn't exists")
		return nil, fmt.Errorf("account doesn't exists")
	}
	return account, nil
}

// Scans accounts's table rows
func scanAccounts(row *sql.Rows) (*types.Account, error) {
	var account types.Account
	// This has to be ordered
	err := row.Scan(
		&account.ID,
		&account.Email,
		&account.Password,
		&account.Verified,
		&account.TotpEnabled,
		&account.Updated,
		&account.Created,
	)
	if err != nil {
		log.Error("failed to database scan")
		return nil, err
	}
	return &account, err
}
