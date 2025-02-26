package main

import (
	"database/sql"
	"embed"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/0xalby/based/config"
	"github.com/0xalby/based/database"
	"github.com/0xalby/based/database/drivers"
	"github.com/0xalby/based/handlers"
	"github.com/0xalby/based/middleware"
	"github.com/0xalby/based/services"
	"github.com/charmbracelet/log"
	"github.com/go-chi/chi/v5"
	chiddlware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/go-chi/jwtauth/v5"
	"github.com/joho/godotenv"
)

// An API instance
type API struct {
	addr string
	db   *sql.DB
}

// Creates a new API instance
func NewAPI(addr string, db *sql.DB) *API {
	return &API{
		addr: addr,
		db:   db,
	}
}

// Init is ran before the entry point
func init() {
	// Loading enviroment variables from .env
	if err := godotenv.Load(); err != nil {
		log.Fatal("failed to load .env enviroment variables")
	}
}

// Entry point
func main() {
	// Initializing JWT
	config.InitJWT(os.Getenv("API_JWT_SECRET"))
	// Creating a database connection
	var driver database.Driver
	switch os.Getenv("DATABASE_DRIVER") {
	case "sqlite3":
		driver = &drivers.DriverSqlite3{}
	case "postgres":
		driver = &drivers.DriverPostgres{}
	default:
		log.Fatal("database driver unsupported or not set")
	}
	connection, err := driver.MustConnect(
		os.Getenv("DATABASE_ADDRESS"),
		os.Getenv("DATABASE_USER"),
		os.Getenv("DATABASE_PASSWORD"),
	)
	if err != nil {
		log.Errorf("failed to connect to the %s database %s", os.Getenv("DATABASE_DRIVER"), err)
		return
	}
	defer func() {
		if err := driver.Close(); err != nil {
			log.Errorf("failed to close the database connection %s", err)
			return
		}
	}()
	maxOpenConns, _ := strconv.Atoi(os.Getenv("POSTGRES_MAX_OPEN_CONNS"))
	maxIdleConns, _ := strconv.Atoi(os.Getenv("POSTGRES_MAX_IDLE_CONNS"))
	maxConnsLifetimeMinutes, _ := strconv.Atoi(os.Getenv("POSTGRES_MAX_CONNS_LIFETIME"))
	connection.SetMaxOpenConns(maxOpenConns)
	connection.SetMaxIdleConns(maxIdleConns)
	connection.SetConnMaxLifetime(time.Duration(maxConnsLifetimeMinutes) * time.Minute)
	// Creating an API instance
	api := NewAPI(os.Getenv("API_ADDRESS"), connection)
	// Running the new instance
	api.Run()
}

//go:embed templates/*
var templateFS embed.FS

// Running
func (server *API) Run() error {
	// Creating a router
	router := chi.NewRouter()
	// Rate limiting everything reasonably
	router.Use(httprate.LimitByIP(50, time.Hour/2))
	// Ensuring the log directory exists
	if err := os.MkdirAll("log", 0755); err != nil {
		log.Fatal("failed to create log directory", "err", err)
	}
	// Opening a file for logging
	file, err := os.OpenFile("log/api.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("failed to open log file", "err", err)
	}
	defer file.Close()
	// Creating a logger
	multiWriter := io.MultiWriter(os.Stdout, file)
	logger := log.New(multiWriter)
	logger.SetReportTimestamp(false)
	logger.SetReportCaller(false)
	logger.SetLevel(log.InfoLevel)
	// Creating a subrouter
	subrouter := chi.NewRouter()
	// Mounting the subrouter with versioning
	router.Mount("/api/v"+os.Getenv("API_VERSION"), subrouter)
	// Creating services
	accountService := &services.AccountsService{DB: server.db}
	emailService := &services.EmailService{FS: templateFS, DB: server.db}
	totpService := &services.TotpService{DB: server.db}
	blacklistService := &services.BlacklistService{DB: server.db}
	// Creating handlers
	accountHandler := &handlers.AccountsHandler{AS: accountService, ES: emailService, TS: totpService}
	authHandler := &handlers.AuthHandler{AS: accountService, ES: emailService, TS: totpService, BS: blacklistService}
	// Using the real ip middleware
	subrouter.Use(chiddlware.RealIP)
	// Using the logger middleware
	subrouter.Use(middleware.Logger(*logger))
	// Registering the routes
	subrouter.Route("/auth", func(r chi.Router) {
		r.With(httprate.LimitByIP(20, time.Hour)).
			Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
		r.With(jwtauth.Verifier(config.TokenAuth)).
			With(jwtauth.Authenticator(config.TokenAuth)).
			With(middleware.Revocation(authHandler)).
			With(middleware.Verified(authHandler)).
			Post("/logout", authHandler.Logout)
		if os.Getenv("SMTP_ADDRESS") != "" {
			r.With(httprate.LimitByIP(5, time.Hour*24)).
				Post("/verification", authHandler.Verification)
			r.With(httprate.LimitByIP(5, time.Hour*24)).
				Post("/resend", authHandler.ResendVerification)
		}
		r.With(httprate.LimitByIP(5, time.Hour*24)).
			Post("/backup", authHandler.LoginWithBackupCode)
	})
	subrouter.Route("/account", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(jwtauth.Verifier(config.TokenAuth))
			r.Use(jwtauth.Authenticator(config.TokenAuth))
			r.Use(middleware.Revocation(authHandler))
			r.Use(middleware.Verified(authHandler))
			if os.Getenv("SMTP_ADDRESS") != "" {
				r.With(httprate.LimitByIP(5, 24*time.Hour)).
					Get("/confirmation", accountHandler.SendConfirmationEmail)
			}
			r.Put("/update/email", accountHandler.UpdateEmail)
			r.Put("/update/password", accountHandler.UpdatePassword)
			r.Put("/totp/enable", accountHandler.AccountEnableTOTP)
			r.Put("/totp/disable", accountHandler.AccountDisableTOTP)
			r.With(httprate.LimitByIP(5, 24*time.Hour)).
				Delete("/delete", accountHandler.DeleteAccount)
		})
		if os.Getenv("SMTP_ADDRESS") != "" {
			r.With(httprate.LimitByIP(5, 24*time.Hour)).
				Get("/recovery", accountHandler.Recovery)
			r.Post("/reset", accountHandler.Reset)
		}
	})
	// Listening
	logger.Printf("running on %s", server.addr)
	return http.ListenAndServe(server.addr, router)
}
