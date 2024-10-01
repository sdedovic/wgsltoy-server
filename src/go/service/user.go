package service

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sdedovic/wgsltoy-server/src/go/infra"
	"log"
	"regexp"
	"strings"
)

// UsernameRegex validates input is 5 to 15 chars, first one is letter, rest are alphanumeric, -, _, .
var usernameRegex = regexp.MustCompile(`^[[:alpha:]][[:alnum:]-_.]{4,14}`)

// UsernameBlacklist is the list of potentially abusive usernames
var usernameBlacklist = []string{
	"about", "access", "account", "accounts", "address", "admin", "administration", "advertising", "affiliate", "affiliates",
	"analytics", "anonymous", "archive", "authentication", "backup", "banner", "banners", "billing", "business", "careers",
	"contact", "contest", "dashboard", "delete", "deleteme", "deleted", "download", "downloads", "favorite", "feedback",
	"guest", "information", "mailer", "mailing", "manager", "marketing", "newsletter", "operator", "password", "postmaster",
	"project", "projects", "random", "register", "registration", "settings", "subscribe", "support", "supportsystem", "username",
	"website", "websites", "webmaster", "webmail", "yourname", "yourusername", "yoursite", "yourdomain",
}

func UserRegister(ctx context.Context, pgPool *pgxpool.Pool, username string, password string) error {
	if len(username) == 0 {
		return infra.NewValidationError("Field 'username' is required!")
	}

	if len(password) == 0 {
		return infra.NewValidationError("Field 'password' is required!")
	}

	if !usernameRegex.MatchString(username) {
		return infra.NewValidationError("Supplied username is not valid!")
	}
	for _, element := range usernameBlacklist {
		if strings.EqualFold(username, element) {
			log.Println("WARN", "Banned username attempted:", element)
			return infra.NewValidationError("Supplied username is not permitted!")
		}
	}

	if len(password) < 10 {
		return infra.NewValidationError("Supplied password is too short!")
	}

	// get connection from pool
	conn, err := pgPool.Acquire(context.Background())
	if err != nil {
		return errors.New("oops")
	}

	// pretend query db
	var text string
	err = conn.QueryRow(context.Background(), "SELECT 'Hello, World!'").Scan(&text)
	if err != nil {
		return errors.New("oops")
	}

	return nil
}
