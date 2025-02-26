package middleware

import (
	"net/http"

	"github.com/0xalby/based/handlers"
	"github.com/0xalby/based/utils"
	"github.com/go-chi/jwtauth/v5"
)

// Middleware blocking blacklisted tokens
func Revocation(handler *handlers.AuthHandler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Getting the token from the request context
			token, _, err := jwtauth.FromContext(r.Context())
			if err != nil {
				utils.Response(w, http.StatusUnauthorized,
					map[string]interface{}{"message": "invalid token", "status": http.StatusUnauthorized},
				)
				return
			}
			tokenID := token.JwtID()
			if tokenID == "" {
				utils.Response(w, http.StatusUnauthorized,
					map[string]interface{}{"message": "missing token", "status": http.StatusUnauthorized},
				)
				return
			}
			// Querying the database for the token
			var exists bool
			err = handler.BS.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM blacklist WHERE token = ?)", tokenID).
				Scan(&exists)
			if err != nil {
				utils.Response(w, http.StatusInternalServerError,
					map[string]interface{}{"message": "internal server error", "status": http.StatusInternalServerError},
				)
				return
			}
			// Denying access if the token is blacklisted
			if exists {
				utils.Response(w, http.StatusUnauthorized,
					map[string]interface{}{"message": "token revoked", "status": http.StatusUnauthorized},
				)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
