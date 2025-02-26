package handlers

import (
	"net/http"
	"os"
	"time"

	"github.com/0xalby/base/config"
	"github.com/0xalby/base/services"
	"github.com/0xalby/base/types"
	"github.com/0xalby/base/utils"
	"github.com/go-chi/jwtauth/v5"
	"github.com/google/uuid"
)

type AuthHandler struct {
	AS *services.AccountsService
	ES *services.EmailService
	TS *services.TotpService
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
	// Hashing the password
	hashed, err := utils.Hash(payload.Password)
	if err != nil {
		utils.Response(w, http.StatusInternalServerError,
			map[string]interface{}{"message": "internal server error", "status": http.StatusInternalServerError},
		)
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
			utils.Response(w, http.StatusConflict,
				map[string]interface{}{"message": "email already used", "status": http.StatusConflict},
			)
			return
		}
		utils.Response(w, http.StatusInternalServerError,
			map[string]interface{}{"message": "internal server error", "status": http.StatusInternalServerError},
		)
		return
	}
	// Optionally sending a verification email
	if os.Getenv("SMTP_ADDRESS") != "" {
		// Generating a random code
		code, err := utils.GenerateRandomCode(6)
		if err != nil {
			utils.Response(w, http.StatusInternalServerError, "internal server error")
			return
		}
		// Sending a verification email
		if err := handler.ES.SendVerificationEmail(account.Email, code); err != nil {
			utils.Response(w, http.StatusInternalServerError,
				map[string]interface{}{"message": "internal server error", "status": http.StatusInternalServerError},
			)
			return
		}
		// Getting account by email
		account, err = handler.AS.GetAccountByEmail(account.Email)
		if err != nil {
			if err.Error() == "account doesn't exists" {
				utils.Response(w, http.StatusBadRequest,
					map[string]interface{}{"message": "account not existing", "status": http.StatusBadRequest},
				)
				return
			}
			utils.Response(w, http.StatusInternalServerError,
				map[string]interface{}{"message": "internal server error", "status": http.StatusInternalServerError},
			)
			return
		}
		// Adding the verification code to the database
		if err := handler.ES.AddVerificationCode(code, account.ID); err != nil {
			utils.Response(w, http.StatusInternalServerError,
				map[string]interface{}{"message": "internal server error", "status": http.StatusInternalServerError},
			)
		}
	}
	// Sending a response
	/* Here we could have an http redirect to the email verification page, but can also be handled frontend side */
	utils.Response(w, http.StatusCreated,
		map[string]interface{}{"message": "created", "status": http.StatusCreated},
	)
}

func (handler *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// Creating a payload
	var payload types.PayloadLogin
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
		if err.Error() == "account doesn't exists" {
			utils.Response(w, http.StatusBadRequest,
				map[string]interface{}{"message": "account not existing", "status": http.StatusBadRequest},
			)
			return
		}
		utils.Response(w, http.StatusInternalServerError,
			map[string]interface{}{"message": "internal server error", "status": http.StatusInternalServerError},
		)
		return
	}
	// Comparing passwords
	if !utils.CompareHashedAndPlain(account.Password, payload.Password) {
		utils.Response(w, http.StatusUnauthorized,
			map[string]interface{}{"message": "wrong email or password", "status": http.StatusUnauthorized},
		)
		return
	}
	/* TOTP requires two pages or one page including TOTP input for before sending the request */
	// Asking for totp validation if the account has it enabled
	if account.TotpEnabled {
		valid, err := handler.TS.ValidateTOTP(account.ID, payload.TOTP)
		if !valid {
			utils.Response(w, http.StatusUnauthorized,
				map[string]interface{}{"message": "wrong totp code", "status": http.StatusUnauthorized},
			)
			return
		}
		if err != nil {
			utils.Response(w, http.StatusInternalServerError,
				map[string]interface{}{"message": "internal server error", "status": http.StatusInternalServerError},
			)
			return
		}
	}
	// Generating a new jwt token providing access to protected routes for some time
	expiration := time.Now().Add(time.Hour * 24 * 7)
	_, token, err := config.TokenAuth.Encode(map[string]interface{}{
		"account": account.ID,
		"exp":     expiration.Unix(),
		"jti":     uuid.New().String(),
	})
	if err != nil {
		utils.Response(w, http.StatusInternalServerError,
			map[string]interface{}{"message": "internal server error", "status": http.StatusInternalServerError},
		)
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
	utils.Response(w, http.StatusOK,
		map[string]interface{}{"message": "token generated", "token": token, "redirect": "/", "status": http.StatusOK},
	)
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
	// Getting account id by code ownership
	id, err := handler.ES.GetAccountIDByCodeOwnership(payload.Code)
	if err != nil {
		if err.Error() == "invalid or expired code" {
			utils.Response(w, http.StatusBadRequest,
				map[string]interface{}{"message": "invalid or expired code", "status": http.StatusBadRequest},
			)
			return
		}
		utils.Response(w, http.StatusInternalServerError,
			map[string]interface{}{"message": "internal server error", "status": http.StatusInternalServerError},
		)
		return
	}
	// Getting the account
	account, err := handler.AS.GetAccountByID(id)
	if err != nil {
		utils.Response(w, http.StatusInternalServerError,
			map[string]interface{}{"message": "internal server error", "status": http.StatusInternalServerError},
		)
		return
	}
	// Ensuring the account isn't already verified
	if account.Verified {
		utils.Response(w, http.StatusForbidden,
			map[string]interface{}{"message": "account already verified", "status": http.StatusForbidden},
		)
		return
	}
	// Comparing verification codes
	if err := handler.ES.CompareCodes(payload.Code, account.ID); err != nil {
		if err.Error() == "invalid verification or confirmation code" || err.Error() == "verification or confirmation code has expired" {
			utils.Response(w, http.StatusUnauthorized,
				map[string]interface{}{"message": "invalid or expired code", "status": http.StatusUnauthorized},
			)
			return
		}
		utils.Response(w, http.StatusInternalServerError,
			map[string]interface{}{"message": "internal server error", "status": http.StatusInternalServerError},
		)
		return
	}
	// Marking account as verified
	if err := handler.AS.MarkAccountAsVerified(account.ID); err != nil {
		utils.Response(w, http.StatusInternalServerError,
			map[string]interface{}{"message": "internal server error", "status": http.StatusInternalServerError},
		)
		return
	}
	utils.Response(w, http.StatusOK,
		map[string]interface{}{"message": "verified", "status": http.StatusOK},
	)
}

/* There will likely be a button for him/her to get a new one and send another email. */
// This if for a user that tries to log in and cannot do anything because the email isn't verified but the first code send after registration expired.
func (handler *AuthHandler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	// Claiming the account id from request context
	id, err := utils.ContextClaimID(r)
	if err != nil {
		if err.Error() == "failed to get claims" || err.Error() == "account not found in claims or not a float64" {
			utils.Response(w, http.StatusUnauthorized,
				map[string]interface{}{"message": "invalid token", "status": http.StatusUnauthorized},
			)
			return
		}
		utils.Response(w, http.StatusInternalServerError,
			map[string]interface{}{"message": "internal server error", "status": http.StatusInternalServerError},
		)
		return
	}
	// Getting the account
	account, err := handler.AS.GetAccountByID(id)
	if err != nil {
		utils.Response(w, http.StatusInternalServerError,
			map[string]interface{}{"message": "internal server error", "status": http.StatusInternalServerError},
		)
		return
	}
	// Ensuring the account isn't already verified
	if account.Verified {
		utils.Response(w, http.StatusForbidden,
			map[string]interface{}{"message": "account already verified", "status": http.StatusForbidden},
		)
		return
	}
	// Generating a random code
	code, err := utils.GenerateRandomCode(6)
	if err != nil {
		utils.Response(w, http.StatusInternalServerError,
			map[string]interface{}{"message": "internal server error", "status": http.StatusInternalServerError},
		)
		return
	}
	// Sending the verification email
	if err := handler.ES.SendVerificationEmail(account.Email, code); err != nil {
		utils.Response(w, http.StatusInternalServerError,
			map[string]interface{}{"message": "internal server error", "status": http.StatusInternalServerError},
		)
		return
	}
	// Adding the verification code to the database
	if err := handler.ES.AddVerificationCode(code, account.ID); err != nil {
		utils.Response(w, http.StatusInternalServerError,
			map[string]interface{}{"message": "internal server error", "status": http.StatusInternalServerError},
		)
	}
	utils.Response(w, http.StatusOK,
		map[string]interface{}{"message": "verification email resent", "status": http.StatusOK},
	)
}

func (handler *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Claiming the account id from request context
	id, err := utils.ContextClaimID(r)
	if err != nil {
		if err.Error() == "failed to get claims" || err.Error() == "account not found in claims or not a float64" {
			utils.Response(w, http.StatusUnauthorized,
				map[string]interface{}{"message": "invalid token", "status": http.StatusUnauthorized},
			)
			return
		}
		utils.Response(w, http.StatusInternalServerError,
			map[string]interface{}{"message": "internal server error", "status": http.StatusInternalServerError},
		)
		return
	}
	// Extracting the jwt token from the request
	token, _, err := jwtauth.FromContext(r.Context())
	if err != nil {
		utils.Response(w, http.StatusUnauthorized,
			map[string]interface{}{"message": "invalid token", "status": http.StatusUnauthorized},
		)
		return
	}
	// Getting the token id
	tokenID := token.JwtID()
	if tokenID == "" {
		utils.Response(w, http.StatusUnauthorized,
			map[string]interface{}{"message": "missing token id", "status": http.StatusUnauthorized},
		)
		return
	}
	// Claiming the jwt token expiration from the request
	exp, err := utils.ContextClaimExpiration(r)
	if err != nil {
		if err.Error() == "failed to get claims" || err.Error() == "expiration not found in claims or not a float64" {
			utils.Response(w, http.StatusUnauthorized,
				map[string]interface{}{"message": "invalid token", "status": http.StatusUnauthorized},
			)
			return
		}
		utils.Response(w, http.StatusInternalServerError,
			map[string]interface{}{"message": "internal server error", "status": http.StatusInternalServerError},
		)
		return
	}
	// Ensure the jwt token is not already expired
	if exp.Before(time.Now()) {
		utils.Response(w, http.StatusUnauthorized,
			map[string]interface{}{"message": "token has already expired", "status": http.StatusUnauthorized},
		)
		return
	}
	// Revoking the jwt token
	if err := handler.AS.RevokeToken(tokenID, id, exp); err != nil {
		utils.Response(w, http.StatusInternalServerError,
			map[string]interface{}{"message": "internal server error", "status": http.StatusInternalServerError},
		)
		return
	}
	// Clear the jwt cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    "",
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		MaxAge:   -1, // Expire the cookie immediately
		SameSite: http.SameSiteLaxMode,
	})
	utils.Response(w, http.StatusOK,
		map[string]interface{}{"message": "logged out", "status": http.StatusOK},
	)
}

func (handler *AuthHandler) LoginWithBackupCode(w http.ResponseWriter, r *http.Request) {
	// Creating a payload
	var payload types.PayloadLoginWithBackupCode
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
		if err.Error() == "account doesn't exists" {
			utils.Response(w, http.StatusBadRequest,
				map[string]interface{}{"message": "account not existing", "status": http.StatusBadRequest},
			)
			return
		}
		utils.Response(w, http.StatusInternalServerError,
			map[string]interface{}{"message": "internal server error", "status": http.StatusInternalServerError},
		)
		return
	}
	// Ensuring the email is verified
	if !account.Verified {
		utils.Response(w, http.StatusUnauthorized,
			map[string]interface{}{"message": "account not verified", "status": http.StatusUnauthorized},
		)
		return
	}
	// Validate the backup code
	if err := handler.TS.ValidateBackupCode(account.ID, payload.BackupCode); err != nil {
		if err.Error() == "code not found" {
			utils.Response(w, http.StatusUnauthorized,
				map[string]interface{}{"message": "code not found", "status": http.StatusUnauthorized},
			)
			return
		}
		if err.Error() == "invalid backup code" {
			utils.Response(w, http.StatusUnauthorized,
				map[string]interface{}{"message": "invalid backup code", "status": http.StatusUnauthorized},
			)
			return
		}
		utils.Response(w, http.StatusInternalServerError,
			map[string]interface{}{"message": "internal server error", "status": http.StatusInternalServerError},
		)
		return
	}
	// Deleting backup codes for the account
	if err := handler.TS.DeleteBackupCodes(account.ID); err != nil {
		utils.Response(w, http.StatusInternalServerError,
			map[string]interface{}{"message": "internal server error", "status": http.StatusInternalServerError},
		)
		return
	}
	// Generating a totp secret
	key, err := handler.TS.GenerateTOTPSecret(account.Email, account.ID)
	if err != nil {
		utils.Response(w, http.StatusInternalServerError,
			map[string]interface{}{"message": "failed to generate totp secret", "status": http.StatusInternalServerError},
		)
		return
	}
	// Generating a qrcoode
	qrCode, err := handler.TS.GenerateQRCode(key)
	if err != nil {
		utils.Response(w, http.StatusInternalServerError, map[string]interface{}{
			"message": "failed to generate qrcode", "status": http.StatusInternalServerError},
		)
		return
	}
	// Generating backup codes
	codes, err := handler.TS.GenerateBackupCodes(12, 8)
	if err != nil {
		utils.Response(w, http.StatusInternalServerError,
			map[string]interface{}{"message": "internal server error", "status": http.StatusInternalServerError},
		)
		return
	}
	// Adding backup codes
	if err := handler.TS.AddBackupCodes(codes, account.ID); err != nil {
		utils.Response(w, http.StatusInternalServerError,
			map[string]interface{}{"message": "internal server error", "status": http.StatusInternalServerError},
		)
		return
	}
	/* Base64 encoded png image */
	utils.Response(w, http.StatusOK,
		/* This would redirect to the 2fa totp setup page */
		map[string]interface{}{"message": "enabled", "secret": key.Secret(), "qr_code": qrCode, "backup": codes, "redirect": "/2fa", "status": http.StatusOK},
	)
}
