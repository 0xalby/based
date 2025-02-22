package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/charmbracelet/log"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-playground/validator/v10"
)

// Unmarshals json into a type struct
func Unmarshal(w http.ResponseWriter, r *http.Request, payload any) error {
	err := json.NewDecoder(r.Body).Decode(payload)
	if err != nil {
		Response(w, http.StatusBadRequest, "invalid request body")
		return err
	}
	return nil
}

// Initializes the validator
var Validator = validator.New(validator.WithRequiredStructEnabled())

// Validates an application/json body
func Validate(w http.ResponseWriter, r *http.Request, payload any) error {
	// Checks for an empty payload
	if r.Body == nil {
		Response(w, http.StatusBadRequest, "empty request body")
		return fmt.Errorf("empty request")
	}
	// Checks for validation errors in the payload
	if err := Validator.Struct(payload); err != nil {
		if verrs := err.(validator.ValidationErrors); verrs != nil {
			Response(w, http.StatusBadRequest, "failed to validate request body")
			return errors.New(verrs.Error())
		}
	}
	return nil
}

// Sends a response
func Response(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
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
		log.Error("failed to marshal")
		return nil, err
	}
	// Createing an http request
	req, err := http.NewRequest(method, endpoint, bytes.NewBuffer(marshal))
	if err != nil {
		log.Error("failed to create a request")
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
		log.Error("failed to successfully send a request")
		return resp, err
	}
	return resp, nil
}

// Claims the account's ID from the request
func ContextClaimID(r *http.Request) (int, error) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		log.Error("failed to get claims")
		return 0, err
	}
	id, ok := claims["account_id"].(float64)
	if !ok {
		log.Error("account_id not found in claims or not a float64")
		return 0, err
	}
	return int(id), nil
}
