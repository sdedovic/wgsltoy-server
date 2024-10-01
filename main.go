package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//==== General HTTP Stuff ====\\

type ErrorResponse struct {
	Message string `json:"message"`
}

func MakeError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(ErrorResponse{message})
	if err != nil {
		log.Println("ERROR", err.Error())
		os.Exit(1)
	}
}

//==== User Register ====\\

type RegisterUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// UsernameRegex 5 to 15 chars, first one is letter, rest are alphanumeric, -, _, .
var UsernameRegex = regexp.MustCompile(`^[[:alpha:]][[:alnum:]-_.]{4,14}`)

// UsernameBlacklist these are banned usernames
var UsernameBlacklist = []string{
	"about", "access", "account", "accounts", "address", "admin", "administration", "advertising", "affiliate", "affiliates",
	"analytics", "anonymous", "archive", "authentication", "backup", "banner", "banners", "billing", "business", "careers",
	"contact", "contest", "dashboard", "delete", "deleteme", "deleted", "download", "downloads", "favorite", "feedback",
	"guest", "information", "mailer", "mailing", "manager", "marketing", "newsletter", "operator", "password", "postmaster",
	"project", "projects", "random", "register", "registration", "settings", "subscribe", "support", "supportsystem", "username",
	"website", "websites", "webmaster", "webmail", "yourname", "yourusername", "yoursite", "yourdomain",
}

func UserRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		MakeError(w, "Unsupported HTTP method!", 400)
		return
	}

	// parse JSON
	var registerUser RegisterUser
	err := json.NewDecoder(r.Body).Decode(&registerUser)
	if err != nil {
		log.Println("ERROR", err.Error())
		MakeError(w, "Unable to parse request!", 400)
		return
	}

	// validate JSON
	if len(registerUser.Username) == 0 {
		MakeError(w, "Field 'username' is required!", 400)
		return
	}

	if len(registerUser.Password) == 0 {
		MakeError(w, "Field 'password' is required!", 400)
		return
	}

	// validate username
	if !UsernameRegex.MatchString(registerUser.Username) {
		MakeError(w, "Supplied username is not valid!", 400)
		return
	}
	for _, element := range UsernameBlacklist {
		if strings.EqualFold(registerUser.Username, element) {
			log.Println("WARN", "Banned username attempted:", element)
			MakeError(w, "Supplied username is not permitted!", 400)
			return
		}
	}

	// validate password
	if len(registerUser.Password) < 10 {
		MakeError(w, "Supplied password is too short!", 400)
		return
	}

	// get connection from pool
	conn, err := pgPool.Acquire(context.Background())
	if err != nil {
		MakeError(w, "Oops!", 500)
		return
	}

	// pretend query db
	var text string
	err = conn.QueryRow(context.Background(), "SELECT 'Hello, World!'").Scan(&text)
	if err != nil {
		MakeError(w, "Oops!", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(201)
}

//==== Health Check ====\\

type HealthResponse struct {
	Status string `json:"status"`
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		MakeError(w, "Unsupported HTTP method!", 400)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	err := json.NewEncoder(w).Encode(HealthResponse{"ok"})
	if err != nil {
		log.Println("ERROR", err.Error())
		os.Exit(1)
	}
}

//==== Main ====\\

var pgPool *pgxpool.Pool

func main() {
	{
		var err error
		pgPool, err = pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
		if err != nil {
			log.Println("ERROR", "Unable to connect to database!", err.Error())
			os.Exit(1)
		}
	}
	defer pgPool.Close()

	http.HandleFunc("/health", HealthCheck)
	http.HandleFunc("/user/register", UserRegister)

	log.Println("INFO", "Starting server on 0.0.0.0:8080")

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Println("ERROR", err.Error())
		os.Exit(1)
	}
}
