package handlers

import (
	"net/http"

	"github.com/0xalby/base/services"
)

type AccountsHandler struct {
	AS *services.AccountsService
}

func (handler *AccountsHandler) Register(w http.ResponseWriter, r *http.Request) {}
