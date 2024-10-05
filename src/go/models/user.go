package models

import "time"

type UserRegister struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserLogin struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type User struct {
	Id                string    `json:"id" db:"user_id"`
	CreatedAt         time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt         time.Time `json:"updatedAt" db:"updated_at"`
	Username          string    `json:"username" db:"username"`
	Email             string    `json:"email" db:"email"`
	EmailVerification string    `json:"emailVerificationStatus" db:"email_verification"`
	Password          string    `json:"-" db:"password"`
}

// UserPublicProfile represents the public information about a user, omitting things such as email addresses.
type UserPublicProfile struct {
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	Username  string    `json:"username" db:"username"`
}
