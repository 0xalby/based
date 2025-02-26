package services

import (
	"bytes"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"os"
	"strconv"
	"time"

	"github.com/charmbracelet/log"
	"gopkg.in/gomail.v2"
)

// ATTENTION in this file for slightly better structuring I declared relevant structs below the functions

type EmailService struct {
	FS embed.FS
	DB *sql.DB
}

// Sends emails based on template and data
func (service *EmailService) SendEmail(email, subject, path string, data interface{}) error {
	// Creating an smtp server
	port, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		log.Error("failed to convert SMTP_PORT", "err", err)
		return err
	}
	server := &smtpServer{
		Email:    os.Getenv("SMTP_EMAIL"),
		Address:  os.Getenv("SMTP_ADDRESS"),
		Port:     port,
		User:     os.Getenv("SMTP_USER"),
		Password: os.Getenv("SMTP_PASSWORD"),
	}
	// Parsing the template from embedded filesystem
	t, err := template.ParseFS(service.FS, path)
	if err != nil {
		log.Error("failed to parse email template", "err", err)
		return err
	}
	// Executing the template
	var body bytes.Buffer
	if err := t.Execute(&body, data); err != nil {
		log.Error("failed to execute template", "err", err)
		return err
	}
	// Sending the email
	message := gomail.NewMessage()
	message.SetHeader("From", server.Email)
	message.SetHeader("To", email)
	message.SetHeader("Subject", subject)
	message.SetBody("text/html", body.String())
	dialer := gomail.NewDialer(server.Address, server.Port, server.User, server.Password)
	if err := dialer.DialAndSend(message); err != nil {
		log.Error("failed to send email", "err", err)
		return err
	}
	return nil
}

type smtpServer struct {
	Email    string
	Address  string
	Port     int
	User     string
	Password string
}

// Sends a verification email
func (service *EmailService) SendVerificationEmail(email, code string) error {
	data := verification{
		Recipient: email,
		Code:      code,
	}
	return service.SendEmail(email, "Email verification or account changes", "templates/verification.html", data)
}

type verification struct {
	Recipient string
	Code      string
}

// Sends an account recovery email
func (service *EmailService) SendRecoveryEmail(email, code string) error {
	data := recovery{
		Recipient: email,
		Code:      code,
	}
	return service.SendEmail(email, "Account Recovery", "templates/recovery.html", data)
}

type recovery struct {
	Recipient string
	Code      string
}

// Sends a notification email
func (service *EmailService) SendNotificationEmail(email, subject, message string) error {
	data := notification{
		Recipient: email,
		Message:   message,
	}
	return service.SendEmail(email, subject, "templates/notification.html", data)
}

type notification struct {
	Recipient string
	Message   string
}

// Gets an account by code ownership
func (service *EmailService) GetAccountIDByCodeOwnership(code string) (int, error) {
	var account int
	err := service.DB.QueryRow("SELECT account FROM codes WHERE code = ? OR recovery = ?", code, code).
		Scan(&account)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, errors.New("invalid or expired code")
		}
		log.Error("failed to query database", "err", err)
		return 0, err
	}
	return account, nil
}

// Adds the verification code to the database
func (service *EmailService) AddVerificationCode(code string, account int) error {
	// Executing on the database
	expiration := time.Now().Add(15 * time.Minute) // expires in 15 minutes
	rows, err := service.DB.Exec("INSERT INTO codes (code, expiration, account) VALUES (?,?,?)", code, expiration, account)
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
		log.Error("failed to add verification code")
		return fmt.Errorf("no rows affected")
	}
	return nil
}

// Compares the stored and the inputted verification codes
func (service *EmailService) CompareCodes(code string, account int) error {
	var (
		storedCode string
		expiration time.Time
	)
	err := service.DB.QueryRow("SELECT code, expiration FROM codes WHERE code = ? AND account = ?", code, account).
		Scan(&storedCode, &expiration)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("invalid verification or confirmation code")
		}
		log.Error("failed to database select", "err", err)
		return err
	}
	if time.Now().After(expiration) {
		return fmt.Errorf("verification or confirmation code has expired")
	}
	_, err = service.DB.Exec("DELETE FROM codes WHERE account = ?", account)
	if err != nil {
		log.Error("failed to delete used codes", "err", err)
		return err
	}
	return nil
}

// Adds the recovery code to the database
func (service *EmailService) AddRecoveryCode(code string, account int) error {
	// Executing on the database
	expiration := time.Now().Add(15 * time.Minute) // expires in 15 minutes
	rows, err := service.DB.Exec("INSERT INTO codes (recovery, expiration, account) VALUES (?,?,?)", code, expiration, account)
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
		log.Error("failed to add recovery code")
		return fmt.Errorf("no rows affected")
	}
	return nil
}

// Compares the stored and the inputted recovery codes
func (service *EmailService) CompareRecoveryCodes(code string, account int) error {
	var (
		storedCode string
		expiration time.Time
	)
	err := service.DB.QueryRow("SELECT recovery, expiration FROM codes WHERE recovery = ? AND account = ?", code, account).
		Scan(&storedCode, &expiration)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("invalid recovery code")
		}
		log.Error("failed to database select", "err", err)
		return err
	}
	if time.Now().After(expiration) {
		return fmt.Errorf("recovery code has expired")
	}
	_, err = service.DB.Exec("DELETE FROM codes WHERE account = ?", account)
	if err != nil {
		log.Error("failed to delete used codes", "err", err)
		return err
	}
	return nil
}

// Deletes the account codes
func (service *EmailService) DeleteCodes(account int) error {
	rows, err := service.DB.Exec("DELETE FROM codes WHERE account = ?", account)
	if err != nil {
		log.Error("failed to delete codes", "err", err)
		return err
	}
	affected, err := rows.RowsAffected()
	if err != nil {
		log.Error("failed to get affacted rows", "err", err)
		return err
	}
	if affected == 0 {
		log.Error("failed to delete codes")
		return fmt.Errorf("no rows affected")
	}
	return nil
}
