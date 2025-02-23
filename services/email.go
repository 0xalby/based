package services

import (
	"bytes"
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/charmbracelet/log"
	"gopkg.in/gomail.v2"
)

// ATTENTION in this file for slightly better structuring I declared relevant structs below the functions

type EmailService struct {
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
	// Opening the template file
	file, err := os.Open(path)
	if err != nil {
		log.Error("failed to open email template", "err", err)
		return err
	}
	defer file.Close()
	// Reading the template file
	templateData, err := io.ReadAll(file)
	if err != nil {
		log.Error("failed to read email template", "err", err)
		return err
	}
	// Creating and parsing the template
	t, err := template.New("emailTemplate").Parse(string(templateData))
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
	return service.SendEmail(email, "Email Verification", "templates/verification.html", data)
}

type verification struct {
	Recipient string
	Code      string
}

// Sends an account recovery email
func (service *EmailService) SendRecoveryEmail(email string) error {
	data := recovery{}
	return service.SendEmail(email, "Account Recovery", "templates/recovery.html", data)
}

type recovery struct{}

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

// Adds the verification code to the database
func (service *EmailService) AddVerificationCode(code string, account int) error {
	// Executing on the database
	expiration := time.Now().Add(15 * time.Minute) // expires in 15 minutes
	rows, err := service.DB.Exec("INSERT INTO codes (code, account, expiration) VALUES (?,?,?)", code, account, expiration)
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
	var storedCode string
	var expiration time.Time
	err := service.DB.QueryRow("SELECT code, expiration FROM codes WHERE account = ? AND code = ?", account, code).
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
		log.Error("failed to delete used verification codes", "err", err)
		return err
	}
	return nil
}

// Marks the account as verified
func (service *EmailService) MarkAccountAsVerified(id int) error {
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
func (service *EmailService) SavePending(email string, account int) error {
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

func (service *EmailService) CleanPendingEmail(id int) error {
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
