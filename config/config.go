package config

import (
	"github.com/charmbracelet/log"
	"github.com/go-chi/jwtauth/v5"
)

var TokenAuth *jwtauth.JWTAuth

// Initializes a new JWT
func InitJWT(secret string) {
	if secret == "" {
		log.Fatal("jwt secret not set")
	}
	TokenAuth = jwtauth.New("HS256", []byte(secret), nil)
	if TokenAuth == nil {
		log.Fatal("failed to initialize jwt")
		return
	}
}
