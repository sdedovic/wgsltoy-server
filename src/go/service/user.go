package service

import (
	"context"
	"fmt"
	"github.com/sdedovic/wgsltoy-server/src/go/db"
	"github.com/sdedovic/wgsltoy-server/src/go/infra"
	"github.com/sdedovic/wgsltoy-server/src/go/models"
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

func UserRegister(ctx context.Context, username string, email string, password string) error {
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

	hashedPassword, err := HashPassword(password)
	if err != nil {
		return fmt.Errorf("failed password hashing caused by: %w", err)
	}

	_, err = db.UserCreate(username, email, hashedPassword)
	if err != nil {
		return err
	}

	return nil
}

func UserLoginGenerateToken(ctx context.Context, username string, password string) (string, error) {
	if len(username) == 0 {
		return "", infra.NewValidationError("Field 'username' is required!")
	}

	if len(password) == 0 {
		return "", infra.NewValidationError("Field 'password' is required!")
	}

	user, err := db.UserGetByUsername(username)
	if err != nil {
		return "", err
	}

	isMatch, err := VerifyPassword(password, user.Password)
	if err != nil {
		return "", fmt.Errorf("failed verifying user password: %w", err)
	}
	if !isMatch {
		return "", infra.BadLoginError
	}

	token, err := MakeToken(UserInfo{user.Id})
	if err != nil {
		return "", err
	}

	return token, nil
}

func UserGetCurrent(ctx context.Context) (models.User, error) {
	userInfo := ExtractUserInfoFromContext(ctx)
	if userInfo == nil {
		return models.User{}, infra.UnauthorizedError
	}

	user, err := db.UserGetById(userInfo.Id)
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}
