package types

import "time"

// Represents an account in the system
type Account struct {
	ID          int       `json:"id"`           // Unique identifier for the account
	Email       string    `json:"email"`        // Email address of the account
	Pending     string    `json:"pending"`      // Pending email address(used during email updates)
	Password    string    `json:"-"`            // Hashed password
	Verified    bool      `json:"verified"`     // Whether the account is verified
	TotpEnabled bool      `json:"totp_enabled"` // Whether TOTP(2FA) is enabled
	TotpSecret  string    `json:"-"`            // TOTP secret
	Updated     time.Time `json:"updated"`      // Timestamp of the last update
	Created     time.Time `json:"created"`      // Timestamp of account creation
}

// Payloads
type (
	// The payload for registering a new account
	PayloadRegister struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=12,max=128,containsany=!@#$%^&*"`
	}
	// The payload for logging into an account
	PayloadLogin struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=12,max=128,containsany=!@#$%^&*"`
		TOTP     string `json:"totp" validate:"omitempty"` // TOTP code(optional)
	}
	// The payload for verifying an account
	PayloadVerification struct {
		Code string `json:"code" validate:"required,len=6,ascii"`
	}
	// The payload for logging in with a backup code
	PayloadLoginWithBackupCode struct {
		Email      string `json:"email" validate:"required,email"`
		BackupCode string `json:"code" validate:"required,len=8,ascii"`
	}
	// The payload for sending a confirmation email
	PayloadAccountSendConfirmationEmail struct {
		Email string `json:"new" validate:"required,email"`
	}
	// The payload for updating an account's email
	PayloadAccountUpdateEmail struct {
		Code string `json:"code" validate:"required,len=6,ascii"`
	}
	// The payload for updating an account's password
	PayloadAccountUpdatePassword struct {
		Old string `json:"old" validate:"required,min=12,max=128,containsany=!@#$%^&*"`
		New string `json:"new" validate:"required,min=12,max=128,containsany=!@#$%^&*"`
	}
	// The payload for initiating account recovery
	PayloadAccountRecovery struct {
		Email string `json:"email" validate:"required,email"`
	}
	// The payload for resetting an account's password
	PayloadAccountReset struct {
		Code     string `json:"code" validate:"required,len=6,ascii"`
		Password string `json:"password" validate:"required,min=12,max=128,containsany=!@#$%^&*"`
	}
	// The payload for deleting an account.
	PayloadAccountDelete struct {
		Password string `json:"password" validate:"required,min=12,max=128,containsany=!@#$%^&*"` // Account password
	}
)
