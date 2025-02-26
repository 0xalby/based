package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/charmbracelet/log"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

// Unmarshals json into a type struct
func Unmarshal(w http.ResponseWriter, r *http.Request, payload any) error {
	// Checking for an empty payload
	if r.Body == nil {
		Response(w, http.StatusBadRequest, map[string]interface{}{"message": "empty request body", "status": http.StatusBadRequest})
		return fmt.Errorf("empty request")
	}
	// Decoding the payload
	err := json.NewDecoder(r.Body).Decode(payload)
	if err != nil {
		Response(w, http.StatusBadRequest, map[string]interface{}{"message": "invalid request body", "status": http.StatusBadRequest})
		return err
	}
	return nil
}

// Global instance of the validator
var Validator = validator.New(validator.WithRequiredStructEnabled())

// Validates a struct
func Validate(w http.ResponseWriter, r *http.Request, payload any) error {
	if err := Validator.Struct(payload); err != nil {
		if verrs := err.(validator.ValidationErrors); verrs != nil {
			Response(w, http.StatusBadRequest, map[string]interface{}{"message": "failed to validate one or more request body fields", "error": verrs.Error(), "status": http.StatusBadRequest})
			return errors.New(verrs.Error())
		}
	}
	return nil
}

// Sends a response
func Response(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	// Setting security headers
	w.Header().Set("Content-Security-Policy", "default-src 'self'")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	// Adding the http status as an header
	w.WriteHeader(status)
	// Encoding the payload
	return json.NewEncoder(w).Encode(v)
}

// Sends a request
func Request(method string, headers map[string]string, endpoint string, payload any) (*http.Response, error) {
	// Marshaling the payload
	marshal, err := json.Marshal(payload)
	if err != nil {
		log.Error("failed to marshal", "err", err)
		return nil, err
	}
	// Createing an http request
	req, err := http.NewRequest(method, endpoint, bytes.NewBuffer(marshal))
	if err != nil {
		log.Error("failed to create a request", "err", err)
		return nil, err
	}
	// Attaching headers
	for header, i := range headers {
		req.Header.Set(header, i)
	}
	// Creating an http client
	client := &http.Client{Timeout: time.Minute * 10}
	// Sending the request
	resp, err := client.Do(req)
	if err != nil {
		log.Error("failed to successfully send a request", "err", err)
		return resp, err
	}
	return resp, nil
}

// Claims the account's id from the request
func ContextClaimID(r *http.Request) (int, error) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		log.Error("failed to get claims", "err", err)
		return 0, err
	}
	id, ok := claims["account"].(float64)
	if !ok {
		log.Error("account not found in claims or not a float64")
		return 0, fmt.Errorf("account not found in claims or not a float64")
	}
	return int(id), nil
}

// Claims the jwt expiration from the request
func ContextClaimExpiration(r *http.Request) (time.Time, error) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		log.Error("failed to get claims", "err", err)
		return time.Time{}, err
	}
	exp, ok := claims["exp"].(time.Time)
	if !ok {
		log.Error("expiration not found in claims or not a float64")
		return time.Time{}, fmt.Errorf("expiration not found in claims or not a float64")
	}
	return exp, nil
}

// Hashes a string
func Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// Compares hashed strings with plaintext ones
func CompareHashedAndPlain(hashed, plain string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain))
	return err == nil
}

// Generating a random alphanumeric code
func GenerateRandomCode(lenght int) (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	code := make([]byte, lenght)
	for i := range code {
		code[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(code), nil
}
