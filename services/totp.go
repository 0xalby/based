package services

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/0xalby/based/utils"
	"github.com/charmbracelet/log"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
	"golang.org/x/crypto/bcrypt"
)

type TotpService struct {
	DB *sql.DB
}

// Generates and a saves a totp secret
func (service *TotpService) GenerateTOTPSecret(email string, id int) (*otp.Key, error) {
	// Generate a new TOTP key
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Based",
		AccountName: email,
		Digits:      6,
	})
	if err != nil {
		log.Error("failed to generate totp key", "err", err)
		return nil, err
	}
	// Store the secret in the database
	rows, err := service.DB.Exec("UPDATE accounts SET secret = ? WHERE id = ?", key.Secret(), id)
	if err != nil {
		// Checking for affected rows
		affected, err := rows.RowsAffected()
		if err != nil {
			log.Error("failed to get affacted rows", "err", err)
			return nil, err
		}
		if affected == 0 {
			log.Error("failed to add verification code")
			return nil, fmt.Errorf("no rows affected")
		}
		log.Error("failed to store totp secret", "err", err)
		return nil, err
	}
	return key, nil
}

// Wrapping io.Writer
type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error { return nil }

// Generates a qrcode
func (service *TotpService) GenerateQRCode(key *otp.Key) ([]byte, error) {
	// Creating a qrcode
	qrc, err := qrcode.New(key.URL())
	if err != nil {
		log.Error("failed to create QR code", "err", err)
		return nil, err
	}
	// Encode the QR code as a PNG
	var buf bytes.Buffer
	w := standard.NewWithWriter(nopCloser{&buf}, standard.WithQRWidth(16)) // QR Code size
	if err := qrc.Save(w); err != nil {
		log.Error("failed to encode qrcode", "err", err)
		return nil, err
	}
	// Optionally saving the qrcode image for debug purposes
	if os.Getenv("QR_CODE_DEBUG") != "" {
		if err := qrc.Save(w); err != nil {
			log.Error("failed to encode qrcode as png", "err", err)
			return nil, err
		}
		// Ensuring the log directory exists
		if err := os.MkdirAll("qrcode", 0755); err != nil {
			log.Fatal("failed to create log directory", "err", err)
		}
		// Saving the image
		if err := saveQRCode(buf.Bytes(), "qrcode/qrcode.png"); err != nil {
			log.Error("failed to save qrcode", "err", err)
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

// Validates a totp code
func (service *TotpService) ValidateTOTP(id int, code string) (bool, error) {
	// Retrieving the code from the database
	var secret string
	err := service.DB.QueryRow("SELECT secret FROM accounts WHERE id = ?", id).Scan(&secret)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, fmt.Errorf("account not found")
		}
		log.Error("failed to retrieve totp secret", "err", err)
		return false, err
	}
	// Validate the totp code
	valid, err := totp.ValidateCustom(code, secret, time.Now(), totp.ValidateOpts{Skew: 1, Digits: 6})
	if !valid {
		return false, err
	}
	return true, nil
}

// Generates backup codes
func (service *TotpService) GenerateBackupCodes(count int, length int) ([]string, error) {
	codes := make([]string, count)
	for i := 0; i < count; i++ {
		bytes := make([]byte, length)
		if _, err := rand.Read(bytes); err != nil {
			log.Error("failed to generate backup code", "err", err)
			return nil, fmt.Errorf("failed to generate backup code")
		}
		codes[i] = hex.EncodeToString(bytes)[:length] // Truncate to desired length
	}
	return codes, nil
}

// Stores backup codes in the database
func (service *TotpService) AddBackupCodes(codes []string, account int) error {
	// Looping over the codes
	var rows sql.Result
	for _, code := range codes {
		// Hashing the code
		hash, err := utils.Hash(code)
		if err != nil {
			return err
		}
		// Adding the code to the database
		rows, err = service.DB.Exec("INSERT INTO backup (hash, account) VALUES (?, ?)", hash, account)
		if err != nil {
			log.Error("failed to add backup code", "err", err)
			return fmt.Errorf("failed to add backup code")
		}
		// Checking for affected rows
		affected, err := rows.RowsAffected()
		if err != nil {
			log.Error("failed to get affacted rows", "err", err)
			return err
		}
		if affected == 0 {
			log.Error("failed to add backup codes", "err", err)
			return fmt.Errorf("no rows affected")
		}
	}
	return nil
}

// Validates backup codes
func (service *TotpService) ValidateBackupCode(account int, code string) error {
	// Fetch all unused backup codes for the account
	rows, err := service.DB.Query("SELECT id, hash FROM backup WHERE account = ?", account)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Error("no backup codes found for the account", "err", err)
			return fmt.Errorf("code not found")
		}
		log.Error("failed to retrieve backup code", "err", err)
		return err
	}
	defer rows.Close()
	// Looping over the backup codes
	for rows.Next() {
		var (
			id     int
			hashed string
		)
		if err := rows.Scan(&id, &hashed); err != nil {
			log.Error("failed to iterate over rows", "err", err)
			return fmt.Errorf("failed to scan backup code")
		}
		// Compare the provided code with the hashed code
		if err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(code)); err == nil {
			return nil
		}
	}
	return fmt.Errorf("invalid backup code")
}

// Deletes backup codes
func (service *TotpService) DeleteBackupCodes(account int) error {
	result, err := service.DB.Exec("DELETE FROM backup WHERE account = ?", account)
	if err != nil {
		log.Error("failed to delete backup codes", "err", err)
		return fmt.Errorf("failed to delete backup codes")
	}
	// Getting affected rows
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Error("failed to get affected rows", "err", err)
		return fmt.Errorf("failed to check affected rows")
	}
	if rowsAffected == 0 {
		log.Error("no affected rows", "err", err)
		return fmt.Errorf("no affected rows")
	}
	return nil
}

// Saves the qrcode to a png
func saveQRCode(data []byte, filePath string) error {
	// Write the data to a file
	err := os.WriteFile(filePath, data, 0644)
	if err != nil {
		return err
	}
	return nil
}
