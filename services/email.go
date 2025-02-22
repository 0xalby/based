package services

import (
	"bytes"
	"database/sql"
	"html/template"
	"io"
	"os"
	"strconv"

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
	return service.SendEmail(email, "Email Verification", "email/verification.html", data)
}

type verification struct {
	Recipient string
	Code      string
}

// Sends an account recovery email
func (service *EmailService) SendRecoveryEmail(email string) error {
	data := recovery{}
	return service.SendEmail(email, "Account Recovery", "email/recovery.html", data)
}

type recovery struct{}

// Sends a notification email
func (service *EmailService) SendNotificationEmail(email, subject, message string) error {
	data := notification{
		Recipient: email,
		Message:   message,
	}
	return service.SendEmail(email, subject, "email/notification.html", data)
}

type notification struct {
	Recipient string
	Message   string
}
