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

// Creates an account in the database
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
		log.Error("failed to get affacted rows", "err", err)
		return err
	}
	if affected == 0 {
		log.Error("failed to create account")
		return fmt.Errorf("no rows affected")
	}
	return nil
}

// Updates account email in the database
func (service *AccountsService) UpdateAccountEmail(email string, id int) error {
	rows, err := service.DB.Exec("UPDATE accounts SET email = ? WHERE id = ?", email, id)
	if err != nil {
		log.Error("failed to update the database", "err", err)
		return err
	}
	affected, err := rows.RowsAffected()
	if err != nil {
		log.Error("failed to get affacted rows", "err", err)
		return err
	}
	if affected == 0 {
		log.Error("failed to update account email")
		return fmt.Errorf("no rows affected")
	}
	return nil
}

// Updates account password in the database
func (service *AccountsService) UpdateAccountPassword(password string, id int) error {
	rows, err := service.DB.Exec("UPDATE accounts SET password = ? WHERE id = ?", password, id)
	if err != nil {
		log.Error("failed to update the database", "err", err)
		return err
	}
	affected, err := rows.RowsAffected()
	if err != nil {
		log.Error("failed to get affacted rows", "err", err)
		return err
	}
	if affected == 0 {
		log.Error("failed to update account password")
		return fmt.Errorf("no rows affected")
	}
	return nil
}

// Deletes an account in the database
func (service *AccountsService) DeleteAccount(id int) error {
	rows, err := service.DB.Exec("DELETE FROM accounts WHERE id = ?", id)
	if err != nil {
		log.Error("failed to delete account", "err", err)
		return err
	}
	affected, err := rows.RowsAffected()
	if err != nil {
		log.Error("failed to get affacted rows", "err", err)
		return err
	}
	if affected == 0 {
		log.Error("failed to delete account")
		return fmt.Errorf("no rows affected")
	}
	return nil
}

// Gets an account by id
func (service *AccountsService) GetAccountByID(id int) (*types.Account, error) {
	// Querying the database
	rows, err := service.DB.Query("SELECT * FROM accounts WHERE id = ?", id)
	if err != nil {
		log.Error("failed to database query", "err", err)
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
	if account == nil || account.ID == 0 {
		return nil, fmt.Errorf("account doesn't exist")
	}
	if err = rows.Err(); err != nil {
		log.Error("failed iterating rows", "err", err)
		return nil, err
	}
	return account, nil
}

// Gets an account by email
func (service *AccountsService) GetAccountByEmail(email string) (*types.Account, error) {
	// Querying the database
	rows, err := service.DB.Query("SELECT * FROM accounts WHERE email = ?", email)
	if err != nil {
		log.Error("failed to database query", "err", err)
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
		log.Error("failed interating rows", "err", err)
		return nil, err
	}
	if account == nil || account.ID == 0 {
		log.Error("account doesn't exists")
		return nil, fmt.Errorf("account doesn't exists")
	}
	return account, nil
}

// Marks the account as verified
func (service *AccountsService) MarkAccountAsVerified(id int) error {
	rows, err := service.DB.Exec("UPDATE accounts SET verified = 1 WHERE id = ?", id)
	if err != nil {
		// Checking for affected rows
		affected, err := rows.RowsAffected()
		if err != nil {
			log.Error("failed to get affacted rows", "err", err)
			return err
		}
		if affected == 0 {
			log.Error("failed to add verification code")
			return fmt.Errorf("no rows affected")
		}
		log.Error("failed to database update", "err", err)
		return err
	}
	return nil
}

// Saves pending email before confirmation
func (service *AccountsService) SavePending(email string, account int) error {
	rows, err := service.DB.Exec("UPDATE accounts SET pending = ? WHERE id = ?", email, account)
	if err != nil {
		// Checking for affected rows
		affected, err := rows.RowsAffected()
		if err != nil {
			log.Error("failed to get affacted rows", "err", err)
			return err
		}
		if affected == 0 {
			log.Error("failed to add pending email")
			return fmt.Errorf("no rows affected")
		}
		log.Error("failed to database update", "err", err)
		return err
	}
	return nil
}

func (service *AccountsService) CleanPendingEmail(id int) error {
	rows, err := service.DB.Exec("UPDATE accounts SET pending = ? WHERE id = ?", "", id)
	if err != nil {
		// Checking for affected rows
		affected, err := rows.RowsAffected()
		if err != nil {
			log.Error("failed to get affacted rows", "err", err)
			return err
		}
		if affected == 0 {
			log.Error("failed to clean pending email")
			return fmt.Errorf("no rows affected")
		}
		log.Error("failed to database update", "err", err)
		return err
	}
	return nil
}

// Scans accounts's table rows
func scanAccounts(row *sql.Rows) (*types.Account, error) {
	var account types.Account
	// This has to be ordered
	err := row.Scan(
		&account.ID,
		&account.Email,
		&account.Pending,
		&account.Password,
		&account.Verified,
		&account.TotpEnabled,
		&account.Updated,
		&account.Created,
	)
	if err != nil {
		log.Error("failed to database scan", "err", err)
		return nil, err
	}
	return &account, err
}
