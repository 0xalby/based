package types

import "time"

// Data
type Account struct {
	ID          int
	Email       string
	Password    string
	Verified    bool
	TotpEnabled bool
	Updated     time.Time
	Created     time.Time
}

// Payloads
type PayloadRegister struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=12,max=128,containsany=!@#$%^&*"`
}
type PayloadLogin struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}
type PayloadVerification struct {
	Code string `json:"code" validate:"required,len=6,ascii"`
}
type PayloadTotpGenerate struct{}
type PayloadTotpCode struct {
	Code string `json:"code" validate:"required,len=6,ascii"`
}
type PayloadTotpBackup struct{}
type PayloadAccountUpdateEmail struct {
	Old string `json:"old" validate:"required,email"`
	New string `json:"new" validate:"required,email"`
}
type PayloadAccountUpdatePassword struct {
	Old string `json:"old" validate:"required,min=12,max=128,containsany=!@#$%^&*"`
	New string `json:"new" validate:"required,min=12,max=128,containsany=!@#$%^&*"`
}
type PayloadAccountRecovery struct{}
type PayloadAccountDelete struct {
	Password string `json:"password" validate:"required,min=12,max=128,containsany=!@#$%^&*"`
}
