package handlers

import (
	"net/http"
	"os"

	"github.com/0xalby/base/services"
	"github.com/0xalby/base/types"
	"github.com/0xalby/base/utils"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	AS *services.AccountsService
	// ES *services.EmailService
	// TS *services.TotpService
}

// Registers a new user
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	// Creating a payload
	var payload types.PayloadRegister
	// Unmarshaling payload
	if err := utils.Unmarshal(w, r, &payload); err != nil {
		return
	}
	// Validating payload
	if err := utils.Validate(w, r, &payload); err != nil {
		return
	}
	// Hashing password
	hashed, err := Hash(payload.Password)
	if err != nil {
		utils.Response(w, http.StatusInternalServerError, "failed to hash password")
		return
	}
	// Creating an account
	account := &types.Account{
		Email:    payload.Email,
		Password: hashed,
	}
	if err := h.AS.CreateAccount(account); err != nil {
		// switch err.Error()
		if err.Error() == "email already used" {
			utils.Response(w, http.StatusConflict, "email already used")
			return
		}
		utils.Response(w, http.StatusInternalServerError, "internal server error")
	}
	// Optionally send verification email
	if os.Getenv("SMTP_ADDRESS") != "" {
	}
	// Sending a response
	/* Here we could have an http redirect to the email verification page, but can also be handled frontend side */
	utils.Response(w, http.StatusCreated, "created")
}

// Hashes a string
func Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
