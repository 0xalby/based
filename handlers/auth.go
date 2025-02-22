package handlers

import (
	"net/http"
	"os"
	"time"

	"github.com/0xalby/base/config"
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

func (handler *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
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
	// Getting the account
	account, err := handler.AS.GetAccountByEmail(payload.Email)
	if err != nil {
		utils.Response(w, http.StatusBadRequest, "wrong email")
		return
	}
	// Comparing passwords
	if !utils.CompareHashedAndPlain(account.Password, payload.Password) {
		utils.Response(w, http.StatusBadRequest, "wrong password")
		return
	}
	// Generating a new jwt token providing access to protected routes for some time
	expiration := time.Now().Add(time.Hour * 24 * 7)
	_, token, err := config.TokenAuth.Encode(map[string]interface{}{
		"account": account.ID,
		"exp":     expiration.Unix(),
	})
	if err != nil {
		utils.Response(w, http.StatusInternalServerError, "failes to generate jwt")
		return
	}
	// Setting the jwt token as a secure httponly cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    token,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		MaxAge:   int(time.Until(expiration).Seconds()),
		SameSite: http.SameSiteLaxMode,
		Expires:  expiration})
	/* Here we could have an http redirect to the dashboard page, but can also be handled frontend side */
	response := map[string]string{"token": token, "redirect": "/dashboard"}
	utils.Response(w, http.StatusOK, response)
}

func (handlers *AuthHandler) Verification(w http.ResponseWriter, r *http.Request) {

}

// Hashes a string
func Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
