package handlers

import (
	"net/http"
	"os"
	"time"

	"github.com/0xalby/base/config"
	"github.com/0xalby/base/services"
	"github.com/0xalby/base/types"
	"github.com/0xalby/base/utils"
)

type AuthHandler struct {
	AS *services.AccountsService
	ES *services.EmailService
	// TS *services.TotpService
}

func (handler *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
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
	hashed, err := utils.Hash(payload.Password)
	if err != nil {
		utils.Response(w, http.StatusInternalServerError, "failed to hash password")
		return
	}
	// Creating an account
	account := &types.Account{
		Email:    payload.Email,
		Password: hashed,
	}
	if err := handler.AS.CreateAccount(account); err != nil {
		// switch err.Error()
		if err.Error() == "email already used" {
			utils.Response(w, http.StatusConflict, "email already used")
			return
		}
		utils.Response(w, http.StatusInternalServerError, "internal server error")
	}
	// Optionally sending a verification email
	if os.Getenv("SMTP_ADDRESS") != "" {
		// Generates a random code
		code, err := utils.GenerateRandomCode(6)
		if err != nil {
			utils.Response(w, http.StatusInternalServerError, "internal server error")
			return
		}
		// Sends a verification email
		if err := handler.ES.SendVerificationEmail(account.Email, code); err != nil {
			utils.Response(w, http.StatusInternalServerError, "internal server error")
			return
		}
		// Getting account by email
		account, err = handler.AS.GetAccountByEmail(account.Email)
		if err != nil {
			utils.Response(w, http.StatusInternalServerError, "internal server error")
			return
		}
		// Adds the code to the database
		if err := handler.ES.AddVerificationCode(code, account.ID); err != nil {
			utils.Response(w, http.StatusInternalServerError, "internal server error")
		}
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
		utils.Response(w, http.StatusUnauthorized, "wrong email or password")
		return
	}
	// Comparing passwords
	if !utils.CompareHashedAndPlain(account.Password, payload.Password) {
		utils.Response(w, http.StatusUnauthorized, "wrong email or password")
		return
	}
	// Generating a new jwt token providing access to protected routes for some time
	expiration := time.Now().Add(time.Hour * 24 * 7)
	_, token, err := config.TokenAuth.Encode(map[string]interface{}{
		"account": account.ID,
		"exp":     expiration.Unix(),
	})
	if err != nil {
		utils.Response(w, http.StatusInternalServerError, "failed to generate jwt")
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

func (handler *AuthHandler) Verification(w http.ResponseWriter, r *http.Request) {
	// Creating a payload
	var payload types.PayloadVerification
	// Unmarshaling payload
	if err := utils.Unmarshal(w, r, &payload); err != nil {
		return
	}
	// Validating payload
	if err := utils.Validate(w, r, &payload); err != nil {
		return
	}
	// Claiming the account id from request context
	id, err := utils.ContextClaimID(r)
	if err != nil {
		if err.Error() == "failed to get claims" || err.Error() == "account not found in claims or not a float64" {
			utils.Response(w, http.StatusUnauthorized, "invalid token")
			return
		}
		utils.Response(w, http.StatusInternalServerError, "internal server error")
		return
	}
	// Getting the account
	account, err := handler.AS.GetAccountByID(id)
	if err != nil {
		utils.Response(w, http.StatusInternalServerError, "internal server error")
		return
	}
	// Ensuring the account isn't already verified
	if account.Verified {
		utils.Response(w, http.StatusForbidden, "account already verified")
		return
	}
	// comparing verification codes
	if err := handler.ES.CompareCodes(payload.Code, id); err != nil {
		if err.Error() == "invalid verification code" || err.Error() == "verification code expired" {
			utils.Response(w, http.StatusUnauthorized, "wrong verification code")
			return
		}
		utils.Response(w, http.StatusInternalServerError, "internal server error")
		return
	}
	// Marking account as verified
	if err := handler.ES.MarkAccountAsVerified(id); err != nil {
		utils.Response(w, http.StatusInternalServerError, "internal server error")
		return
	}
	utils.Response(w, http.StatusOK, "verified")
}
