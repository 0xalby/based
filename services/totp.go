package services

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/charmbracelet/log"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
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
	_, err = service.DB.Exec("UPDATE accounts SET secret = ? WHERE id = ?", key.Secret(), id)
	if err != nil {
		log.Error("failed to store TOTP secret", "err", err)
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
	w := standard.NewWithWriter(nopCloser{&buf}, standard.WithQRWidth(64)) // QR Code size
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
		if err := saveQRCode(buf.Bytes(), "qrcode.png"); err != nil {
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
	// Validate the TOTP code
	valid, err := totp.ValidateCustom(code, secret, time.Now(), totp.ValidateOpts{Skew: 1, Digits: 6})
	if !valid {
		return false, err
	}
	return true, nil
}

func saveQRCode(data []byte, filePath string) error {
	// Write the data to a file
	err := os.WriteFile(filePath, data, 0644)
	if err != nil {
		return err
	}
	return nil
}
