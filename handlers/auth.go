package handlers

import "net/http"

type AuthHandler struct {
	// AS *services.AccountsService
	// ES *services.EmailService
	// TS *services.TotpService
}

// Registers a new user
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	// Creating a payload
	// var payload types.PayloadSignUp
	// Validating and unmarshaling
	// Hashing password
	// Adding the user to the database
	// Optionally send verification email
	// if os.Getenv("SMTP_ADDRESS") != "" {}
	// Sending a response
	/* Here we could have an http redirect to the email verification page, but can also be handled frontend side */
}
