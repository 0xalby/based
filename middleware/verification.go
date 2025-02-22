package middleware

import (
	"context"
	"net/http"

	"github.com/0xalby/base/handlers"
	"github.com/0xalby/base/utils"
)

type key int

const userKey key = 0

// Middleware function assuring accounts are verified
func Verified(handler *handlers.AuthHandler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
			// Checking for email verification
			if !account.Verified {
				utils.Response(w, http.StatusForbidden, "email not verified")
				return
			}
			ctx := context.WithValue(r.Context(), userKey, account)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
