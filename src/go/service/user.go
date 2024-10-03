package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sdedovic/wgsltoy-server/src/go/infra"
	"log"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
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

func UserRegister(ctx context.Context, pgPool *pgxpool.Pool, username string, email string, password string) error {
	if len(username) == 0 {
		return infra.NewValidationError("Field 'username' is required!")
	}

	if len(email) == 0 {
		return infra.NewValidationError("Field 'email' is required!")
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

	if utf8.RuneCountInString(password) < 10 {
		return infra.NewValidationError("Supplied password is too short!")
	}

	passwordHash, err := HashPassword(password)
	if err != nil {
		return fmt.Errorf("failed password hashing caused by: %w", err)
	}

	now := time.Now()
	_, err = pgPool.Exec(ctx, "INSERT INTO users (username, email, email_verification, password, created_at, updated_at) VALUES ($1, $2, 'pending', $3, $4, $4)", username, email, passwordHash, now)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {

			// uniqueness constraint violation
			if pgErr.Code == "23505" {
				if pgErr.ConstraintName == "unique_email" {
					return infra.NewValidationError("Email is already taken!")
				}

				if pgErr.ConstraintName == "unique_username" {
					return infra.NewValidationError("Username is already taken!")
				}
			}
		}

		// catchall
		return fmt.Errorf("failed inserting user caused by: %w", err)
	}

	return nil
}

func UserLogin(ctx context.Context, pgPool *pgxpool.Pool, username string, password string) (string, error) {
	if len(username) == 0 {
		return "", infra.NewValidationError("Field 'username' is required!")
	}

	if len(password) == 0 {
		return "", infra.NewValidationError("Field 'password' is required!")
	}

	var userId string
	var storedPassword string
	err := pgPool.QueryRow(ctx, "SELECT user_id, password FROM users WHERE username = $1;", username).Scan(&userId, &storedPassword)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", infra.BadLoginError
		}
		return "", fmt.Errorf("failed querying user caused by: %w", err)
	}

	isMatch, err := VerifyPassword(password, storedPassword)
	if err != nil {
		return "", fmt.Errorf("failed verifying user password: %w", err)
	}
	if !isMatch {
		return "", infra.BadLoginError
	}

	token, err := MakeToken(UserInfo(userId))
	if err != nil {
		return "", err
	}

	return token, nil
}

type User struct {
	Id                string    `db:"user_id"`
	Username          string    `db:"username"`
	Email             string    `db:"email"`
	EmailVerification string    `db:"email_verification"`
	CreatedAt         time.Time `db:"created_at"`
	UpdatedAt         time.Time `db:"updated_at"`
}

func UserFindOne(ctx context.Context, pgPool *pgxpool.Pool, userId string) (*User, error) {
	rows, err := pgPool.Query(ctx, "SELECT user_id, email, email_verification, username, created_at, updated_at FROM users WHERE user_id = $1 LIMIT 1", userId)
	if err != nil {
		return nil, fmt.Errorf("failed querying user caused by: %w", err)
	}

	user, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[User])
	if err != nil {
		return nil, fmt.Errorf("failed deserializing database rows caused by: %w", err)
	}

	return user, nil
}
