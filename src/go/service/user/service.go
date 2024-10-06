package user

import (
	"context"
	"fmt"
	"github.com/sdedovic/wgsltoy-server/src/go/db"
	"github.com/sdedovic/wgsltoy-server/src/go/infra"
	"github.com/sdedovic/wgsltoy-server/src/go/models"
	"github.com/sdedovic/wgsltoy-server/src/go/service"
	"log"
	"regexp"
	"strings"
	"unicode/utf8"
)

type Service struct {
	repo db.IRepository `di.inject:"Repository"`
}

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

func (s *Service) Register(ctx context.Context, username string, email string, password string) error {
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

	_, err = s.repo.UserCreate(username, email, hashedPassword)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Login(ctx context.Context, username string, password string) (string, error) {
	if len(username) == 0 {
		return "", infra.NewValidationError("Field 'username' is required!")
	}

	if len(password) == 0 {
		return "", infra.NewValidationError("Field 'password' is required!")
	}

	user, err := s.repo.UserGetByUsername(username)
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

	token, err := service.MakeToken(service.UserInfo{user.Id})
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *Service) GetCurrent(ctx context.Context) (models.User, error) {
	userInfo := service.ExtractUserInfoFromContext(ctx)
	if userInfo == nil {
		return models.User{}, infra.UnauthorizedError
	}

	user, err := s.repo.UserGetById(userInfo.Id)
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}
