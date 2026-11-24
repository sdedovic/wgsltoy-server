package domain

import (
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
	"uuid"
)

//==== UserID ====\\

type UserID string

func NewUserID(s string) (UserID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return "", NewValidationError("Supplied user id is not valid!")
	}
	return UserID(id.String()), nil
}

func GenerateUserID() UserID {
	return UserID(uuid.New().String())
}

//==== Username ====\\

// 5 to 15 chars: a letter, then letters, digits, -, _ or .
var usernameRegex = regexp.MustCompile(`^[[:alpha:]][[:alnum:]-_.]{4,14}$`)

// names that could pass for staff or system accounts
var usernameBlacklist = []string{
	"about", "access", "account", "accounts", "address", "admin", "administration", "advertising", "affiliate", "affiliates",
	"analytics", "anonymous", "archive", "authentication", "backup", "banner", "banners", "billing", "business", "careers",
	"contact", "contest", "dashboard", "delete", "deleteme", "deleted", "download", "downloads", "favorite", "feedback",
	"guest", "information", "mailer", "mailing", "manager", "marketing", "newsletter", "operator", "password", "postmaster",
	"project", "projects", "random", "register", "registration", "settings", "subscribe", "support", "supportsystem", "username",
	"website", "websites", "webmaster", "webmail", "yourname", "yourusername", "yoursite", "yourdomain",
}

// separate from other validation errors, so attempts can be logged
var ErrUsernameBanned = NewValidationError("Supplied username is not permitted!")

var ErrUsernameTaken = NewValidationError("Username is already taken!")

type Username string

func NewUsername(s string) (Username, error) {
	if len(s) == 0 {
		return "", NewValidationError("Field 'username' is required!")
	}
	if !usernameRegex.MatchString(s) {
		return "", NewValidationError("Supplied username is not valid!")
	}
	for _, element := range usernameBlacklist {
		if strings.EqualFold(s, element) {
			return "", ErrUsernameBanned
		}
	}
	return Username(s), nil
}

//==== Email ====\\

var ErrEmailTaken = NewValidationError("Email is already taken!")

type Email string

func NewEmail(s string) (Email, error) {
	if len(s) == 0 {
		return "", NewValidationError("Field 'email' is required!")
	}
	if _, err := mail.ParseAddress(s); err != nil {
		return "", NewValidationError("Field 'email' is not valid!")
	}
	return Email(s), nil
}

type EmailVerification string

const (
	EmailVerificationPending   EmailVerification = "pending"
	EmailVerificationCompleted EmailVerification = "completed"
)

//==== Password ====\\

type Password string

// Login doesn't use this, so tightening the policy doesn't lock out existing users.
func NewPassword(s string) (Password, error) {
	if len(s) == 0 {
		return "", NewValidationError("Field 'password' is required!")
	}
	if utf8.RuneCountInString(s) < 10 {
		return "", NewValidationError("Supplied password is too short!")
	}
	return Password(s), nil
}

//==== UserVisibility ====\\

// NewUserVisibility validates that the value is one of the two visibility levels supported for user accounts.
func NewUserVisibility(s string) (Visibility, error) {
	switch Visibility(s) {
	case VisibilityPrivate, VisibilityPublic:
		return Visibility(s), nil
	case "":
		return "", NewValidationError("Field 'visibility' may not be empty!")
	default:
		return "", NewValidationError("Field 'visibility' must be one of 'private' or 'public'!")
	}
}

//==== User ====\\

type User struct {
	ID                UserID
	CreatedAt         time.Time
	UpdatedAt         time.Time
	Username          Username
	Email             Email
	EmailVerification EmailVerification
	PasswordHash      string
	Visibility        Visibility
	IsAdmin           bool
}

type UserChanges struct {
	Username   Username
	Email      Email
	Visibility Visibility
}

func (u *User) Apply(changes UserChanges, updatedAt time.Time) {
	u.Username = changes.Username
	if changes.Email != u.Email {
		u.Email = changes.Email
		u.EmailVerification = EmailVerificationPending
	}
	u.Visibility = changes.Visibility
	u.UpdatedAt = updatedAt
}

func (u User) IsVisibleTo(p Principal) bool {
	return p.User.IsAdmin || u.Visibility != VisibilityPrivate || u.ID == p.User.ID
}

func (u User) IsEditableBy(p Principal) bool {
	return p.User.IsAdmin || u.ID == p.User.ID
}
