package infra

import (
	"errors"
	"fmt"
)

//==== JSON Parsing ====\\

// JsonParsingError wraps JSON unmarshalling errors
type JsonParsingError struct {
	cause error
}

func (e JsonParsingError) Error() string {
	return fmt.Sprintf("Unable to parse JSON, caused by: %v", e.cause)
}

func (e JsonParsingError) Unwrap() error {
	return e.cause
}

func NewJsonParsingError(cause error) error {
	return JsonParsingError{cause}
}

//==== Validation ====\\

// ValidationError occurs when user supplied inputs are rejected according to business logic
type ValidationError struct {
	message string
}

func (e ValidationError) Error() string {
	return e.message
}

func NewValidationError(message string) error {
	return ValidationError{message}
}

//==== Misc. ====\\

// BadLoginError occurs when the provided credentials fail to authenticate
var BadLoginError = errors.New("bad login")

// UnauthorizedError occurs when a user lacks access while attempting to perform an operation
var UnauthorizedError = errors.New("unauthorized")
