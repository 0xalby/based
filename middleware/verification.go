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
			// Checking for email verification
			if !account.Verified {
				utils.Response(w, http.StatusForbidden,
					map[string]interface{}{"message": "email not verified", "status": http.StatusForbidden},
				)
				return
			}
			ctx := context.WithValue(r.Context(), userKey, account)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
