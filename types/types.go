package types

// Data
type Account struct{}

// Payloads
type PayloadRegister struct{}
type PayloadLogin struct{}
type PayloadVerification struct{}
type PayloadTotpGenerate struct{}
type PayloadTotpCode struct{}
type PayloadTotpRecover struct{}
type PayloadAccountUpdateEmail struct{}
type PayloadAccountUpdatePassword struct{}
type PayloadAccountRecover struct{}
type PayloadAccountDelete struct{}
