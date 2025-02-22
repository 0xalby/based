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
		utils.Response(w, http.StatusInternalServerError, "internal server error")
		return
	}
	// Updating account email
	if err := handler.AS.UpdateAccountEmail(id, payload.New, payload.Old); err != nil {
		if err.Error() == "no rows affected" {
			utils.Response(w, http.StatusBadRequest, "wrong email address")
			return
		}
		utils.Response(w, http.StatusInternalServerError, "internal server error")
		return
	}
	// Optionally send email notification
	if os.Getenv("SMTP_ADDRESS") != "" {
	}
	// Sending a response
	utils.Response(w, http.StatusOK, "updated")
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
	// Claiming the account id from request context
	id, err := utils.ContextClaimID(r)
	if err != nil {
		utils.Response(w, http.StatusInternalServerError, "internal server error")
		return
	}
	// Getting the account
	account, err := handler.AS.GetAccountByID(id)
	if err != nil {
		utils.Response(w, http.StatusInternalServerError, "internal server error")
		return
	}
	// Comparing passwords
	if !utils.CompareHashedAndPlain(account.Password, payload.Old) {
		utils.Response(w, http.StatusBadRequest, "wrong password")
		return
	}
	// Updating account password
	if err := handler.AS.UpdateAccountPassword(id, payload.New, payload.Old); err != nil {
		if err.Error() == "no rows affected" {
			utils.Response(w, http.StatusBadRequest, "wrong password")
			return
		}
		utils.Response(w, http.StatusInternalServerError, "internal server error")
		return
	}
	// Optionally send email notification
	if os.Getenv("SMTP_ADDRESS") != "" {
	}
	// Sending a response
	utils.Response(w, http.StatusOK, "updated")
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
		utils.Response(w, http.StatusInternalServerError, "internal server error")
		return
	}
	// Getting the account
	account, err := handler.AS.GetAccountByID(id)
	if err != nil {
		utils.Response(w, http.StatusInternalServerError, "internal server error")
		return
	}
	// Comparing passwords
	if !utils.CompareHashedAndPlain(account.Password, payload.Password) {
		utils.Response(w, http.StatusBadRequest, "wrong password")
		return
	}
	// Deleting the user from the database
	if err := handler.AS.DeleteAccount(id); err != nil {
		if err.Error() == "no rows affected" {
			utils.Response(w, http.StatusBadRequest, "bad request")
			return
		}
		utils.Response(w, http.StatusInternalServerError, "internal server error")
		return
	}
	utils.Response(w, http.StatusOK, "deleted")
}
