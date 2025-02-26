package handlers

import (
	"net/http"
	"os"

	"github.com/0xalby/base/services"
	"github.com/0xalby/base/types"
	"github.com/0xalby/base/utils"
)

type AccountsHandler struct {
	AS *services.AccountsService
	ES *services.EmailService
	TS *services.TotpService
}

func (handler *AccountsHandler) SendConfirmationEmail(w http.ResponseWriter, r *http.Request) {
	// Creating a payload
	var payload types.PayloadAccountSendConfirmationEmail
	// Unmarshaling payload
	if err := utils.Unmarshal(w, r, &payload); err != nil {
		return
	}
	// Validating payload
	if err := utils.Validate(w, r, &payload); err != nil {
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
	// Ensuring emails are different
	if account.Email == payload.Email {
		utils.Response(w, http.StatusBadRequest,
			map[string]interface{}{"message": "the new email has to be different from the old one", "status": http.StatusBadRequest},
		)
		return
	}
	// Adds the code to the database
	if err := handler.ES.AddVerificationCode(code, id); err != nil {
		utils.Response(w, http.StatusInternalServerError,
			map[string]interface{}{"message": "internal server error", "status": http.StatusInternalServerError},
		)
	}
	// Saving pending email
	if err := handler.AS.SavePending(payload.Email, id); err != nil {
		utils.Response(w, http.StatusInternalServerError,
			map[string]interface{}{"message": "internal server error", "status": http.StatusInternalServerError},
		)
		return
	}
	// Sending confirmation email
	if err := handler.ES.SendVerificationEmail(payload.Email, code); err != nil {
		utils.Response(w, http.StatusInternalServerError,
			map[string]interface{}{"message": "internal server error", "status": http.StatusInternalServerError},
		)
		return
	}
	utils.Response(w, http.StatusOK, "confirmation email sent")
}

func (handler *AccountsHandler) UpdateEmail(w http.ResponseWriter, r *http.Request) {
	// Creating a payload
	var payload types.PayloadAccountUpdateEmail
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
		if err.Error() == "account not found in claims or not a float64" {
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
	// Comparing confirmation codes
	if err := handler.ES.CompareCodes(payload.Code, id); err != nil {
		if err.Error() == "invalid verification or confirmation code" || err.Error() == "verification or confirmation code has expired" {
			utils.Response(w, http.StatusUnauthorized,
				map[string]interface{}{"message": "invalid confirmation code", "status": http.StatusUnauthorized},
			)
			return
		}
		utils.Response(w, http.StatusInternalServerError,
			map[string]interface{}{"message": "internal server error", "status": http.StatusInternalServerError},
		)
		return
	}
	// Updating account email
	if err := handler.AS.UpdateAccountEmail(account.Pending, id); err != nil {
		utils.Response(w, http.StatusInternalServerError,
			map[string]interface{}{"message": "internal server error", "status": http.StatusInternalServerError},
		)
		return
	}
	// Clean pending email
	if err := handler.AS.CleanPendingEmail(id); err != nil {
		utils.Response(w, http.StatusInternalServerError,
			map[string]interface{}{"message": "internal server error", "status": http.StatusInternalServerError},
		)
		return
	}
	/* Optionally send email notification */
	if os.Getenv("SMTP_ADDRESS") != "" {
		// Sending a notification email
		if err := handler.ES.SendNotificationEmail(account.Pending, "Updated email address", "Your email address has been updated"); err != nil {
			utils.Response(w, http.StatusInternalServerError,
				map[string]interface{}{"message": "internal server error", "status": http.StatusInternalServerError},
			)
			return
		}
	}
	// Sending a response
	utils.Response(w, http.StatusOK,
		map[string]interface{}{"message": "updated", "status": http.StatusOK},
	)
}

func (handler *AccountsHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	// Creating a payload
	var payload types.PayloadAccountUpdatePassword
	// Unmarshaling payload
	if err := utils.Unmarshal(w, r, &payload); err != nil {
		return
	}
	// Validating payload
	if err := utils.Validate(w, r, &payload); err != nil {
		return
	}
	// Ensuring the passwords are different
	if payload.Old == payload.New {
		utils.Response(w, http.StatusBadRequest,
			map[string]interface{}{"message": "the new password has to be different from the old one", "status": http.StatusBadRequest})
		return
	}
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
	// Comparing passwords
	if !utils.CompareHashedAndPlain(account.Password, payload.Old) {
		utils.Response(w, http.StatusUnauthorized,
			map[string]interface{}{"message": "wrong email or password", "status": http.StatusUnauthorized},
		)
		return
	}
	// Hashes the new password
	hashed, err := utils.Hash(payload.New)
	if err != nil {
		utils.Response(w, http.StatusInternalServerError,
			map[string]interface{}{"message": "internal server error", "status": http.StatusInternalServerError},
		)
		return
	}
	// Updating account password
	if err := handler.AS.UpdateAccountPassword(hashed, id); err != nil {
		utils.Response(w, http.StatusInternalServerError,
			map[string]interface{}{"message": "internal server error", "status": http.StatusInternalServerError},
		)
		return
	}
	// Optionally send email notification
	if os.Getenv("SMTP_ADDRESS") != "" {
		// Getting the account
		account, err := handler.AS.GetAccountByID(id)
		if err != nil {
			utils.Response(w, http.StatusInternalServerError,
				map[string]interface{}{"message": "internal server error", "status": http.StatusInternalServerError},
			)
			return
		}
		// Sending a notification email
		if err := handler.ES.SendNotificationEmail(account.Email, "Updated password", "Your password has been updated"); err != nil {
			utils.Response(w, http.StatusInternalServerError,
				map[string]interface{}{"message": "internal server error", "status": http.StatusInternalServerError},
			)
			return
		}
	}
	// Sending a response
	utils.Response(w, http.StatusOK,
		map[string]interface{}{"message": "updated", "status": http.StatusOK},
	)
}

func (handler *AccountsHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	// Creating a payload
	var payload types.PayloadAccountDelete
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
		if err.Error() == "account doesn't exist" {
			utils.Response(w, http.StatusUnauthorized,
				map[string]interface{}{"message": "account doesn't exist", "status": http.StatusUnauthorized},
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
			map[string]interface{}{"message": "wrong password", "status": http.StatusUnauthorized},
		)
		return
	}
	// Deleting the account from the database
	if err := handler.AS.DeleteAccount(id); err != nil {
		utils.Response(w, http.StatusInternalServerError,
			map[string]interface{}{"message": "internal server error", "status": http.StatusInternalServerError},
		)
		return
	}
	// Deleting left over account codes
	if err := handler.ES.DeleteCodes(id); err != nil {
		if err.Error() != "no rows affected" {
			utils.Response(w, http.StatusInternalServerError,
				map[string]interface{}{"message": "internal server error", "status": http.StatusInternalServerError},
			)
		}
	}
	/* Optionally send email notification */
	if os.Getenv("SMTP_ADDRESS") != "" {
		// Sending a notification email
		if err := handler.ES.SendNotificationEmail(account.Pending, "Deleted account", "Your account has been deleted, goodbye"); err != nil {
			utils.Response(w, http.StatusInternalServerError,
				map[string]interface{}{"message": "internal server error", "status": http.StatusInternalServerError},
			)
			return
		}
	}
	utils.Response(w, http.StatusOK,
		map[string]interface{}{"message": "deleted", "status": http.StatusOK},
	)
}

func (handler *AccountsHandler) Recovery(w http.ResponseWriter, r *http.Request) {
	// Creating a payload
	var payload types.PayloadAccountRecovery
	// Unmarshaling payload
	if err := utils.Unmarshal(w, r, &payload); err != nil {
		return
	}
	// Validating payload
	if err := utils.Validate(w, r, &payload); err != nil {
		return
	}
	// Getting account by email
	account, err := handler.AS.GetAccountByEmail(payload.Email)
	if err != nil {
		if err.Error() == "account doesn't exists" {
			utils.Response(w, http.StatusBadRequest,
				map[string]interface{}{"message": "account not existing", "status": http.StatusBadRequest})
			return
		}
		utils.Response(w, http.StatusInternalServerError,
			map[string]interface{}{"message": "internal server error", "status": http.StatusInternalServerError},
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
	// Adding the recovery code to the database
	if err := handler.ES.AddRecoveryCode(code, account.ID); err != nil {
		utils.Response(w, http.StatusInternalServerError,
			map[string]interface{}{"message": "internal server error", "status": http.StatusInternalServerError},
		)
		return
	}
	// Sending a recovery email with the code
	if err := handler.ES.SendRecoveryEmail(account.Email, code); err != nil {
		utils.Response(w, http.StatusInternalServerError,
			map[string]interface{}{"message": "internal server error", "status": http.StatusInternalServerError},
		)
	}
	utils.Response(w, http.StatusOK,
		map[string]interface{}{"message": "recovery email sent", "status": http.StatusOK},
	)
}

func (handler *AccountsHandler) Reset(w http.ResponseWriter, r *http.Request) {
	// Creating a payload
	var payload types.PayloadAccountReset
	// Unmarshaling payload
	if err := utils.Unmarshal(w, r, &payload); err != nil {
		return
	}
	// Validating payload
	if err := utils.Validate(w, r, &payload); err != nil {
		return
	}
	// Getting account by code ownership
	id, err := handler.ES.GetAccountIDByCodeOwnership(payload.Code)
	if err != nil {
		if err.Error() == "invalid or expired code" {
			utils.Response(w, http.StatusBadRequest,
				map[string]interface{}{"message": "invalid or expired code", "status": http.StatusBadRequest})
			return
		}
		utils.Response(w, http.StatusInternalServerError,
			map[string]interface{}{"message": "internal server error", "status": http.StatusInternalServerError},
		)
		return
	}
	// Comparing recovery codes
	if err := handler.ES.CompareRecoveryCodes(payload.Code, id); err != nil {
		if err.Error() == "invalid recovery code" || err.Error() == "recovery code expired" {
			utils.Response(w, http.StatusUnauthorized,
				map[string]interface{}{"message": "invalid recovery code", "status": http.StatusUnauthorized},
			)
			return
		}
		utils.Response(w, http.StatusInternalServerError,
			map[string]interface{}{"message": "internal server error", "status": http.StatusInternalServerError},
		)
		return
	}
	// Hashes the new password
	hashed, err := utils.Hash(payload.Password)
	if err != nil {
		utils.Response(w, http.StatusInternalServerError,
			map[string]interface{}{"message": "internal server error", "status": http.StatusInternalServerError},
		)
		return
	}
	// Resetting the password
	if err := handler.AS.UpdateAccountPassword(hashed, id); err != nil {
		utils.Response(w, http.StatusInternalServerError,
			map[string]interface{}{"message": "internal server error", "status": http.StatusInternalServerError},
		)
		return
	}
	utils.Response(w, http.StatusOK,
		map[string]interface{}{"message": "recovered", "status": http.StatusOK},
	)
}

func (handler *AccountsHandler) AccountEnableTOTP(w http.ResponseWriter, r *http.Request) {
	// Claiming the account id from request context
	id, err := utils.ContextClaimID(r)
	if err != nil {
		if err.Error() == "account not found in claims or not a float64" {
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
	// Generating a totp secret
	key, err := handler.TS.GenerateTOTPSecret(account.Email, id)
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
	// Enabling 2fa totp for the account
	if err := handler.AS.EnableTOTP(id); err != nil {
		utils.Response(w, http.StatusInternalServerError,
			map[string]interface{}{"message": "internal server error", "status": http.StatusInternalServerError},
		)
		return
	}
	/* Base64 encoded png image */
	utils.Response(w, http.StatusOK,
		map[string]interface{}{"message": "enabled", "secret": key.Secret(), "qr_code": qrCode, "status": http.StatusOK},
	)
}

func (handler *AccountsHandler) AccountDisableTOTP(w http.ResponseWriter, r *http.Request) {
	// Claiming the account id from request context
	id, err := utils.ContextClaimID(r)
	if err != nil {
		if err.Error() == "account not found in claims or not a float64" {
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
	// Disabling 2fa totp for the account
	if err := handler.AS.DisableTOTP(id); err != nil {
		utils.Response(w, http.StatusInternalServerError,
			map[string]interface{}{"message": "internal server error", "status": http.StatusInternalServerError},
		)
		return
	}
	utils.Response(w, http.StatusOK,
		map[string]interface{}{"message": "disabled", "status": http.StatusOK},
	)
}
