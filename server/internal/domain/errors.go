package domain

import "errors"

type ValidationError struct {
	message string
}

func (e ValidationError) Error() string {
	return e.message
}

func NewValidationError(message string) error {
	return ValidationError{message}
}

var ErrBadLogin = errors.New("bad login")

var ErrUnauthorized = errors.New("unauthorized")

var ErrNotFound = errors.New("not found")
