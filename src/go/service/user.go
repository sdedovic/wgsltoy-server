package service

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sdedovic/wgsltoy-server/src/go/infra"
	"github.com/sdedovic/wgsltoy-server/src/go/repository"
	"log"
	"regexp"
	"strings"
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

	_, err = repository.UserInsert(ctx, pgPool, username, email, passwordHash)
	if err != nil {
		return err
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

	userId, storedPassword, err := repository.UserGetUserIdPasswordByUsername(ctx, pgPool, username)

	isMatch, err := VerifyPassword(password, storedPassword)
	if err != nil {
		return "", fmt.Errorf("failed verifying user password: %w", err)
	}
	if !isMatch {
		return "", infra.BadLoginError
	}

	token, err := MakeToken(UserInfo{userId})
	if err != nil {
		return "", err
	}

	return token, nil
}

func UserGetCurrent(ctx context.Context, pgPool *pgxpool.Pool) (*repository.User, error) {
	userInfo := ExtractUserInfoFromContext(ctx)
	if userInfo == nil {
		return nil, infra.UnauthorizedError
	}

	user, err := repository.UserFindOneById(ctx, pgPool, userInfo.UserID())
	if err != nil {
		return nil, err
	}

	return user, nil
}
