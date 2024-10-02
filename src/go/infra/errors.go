package infra

import "fmt"

//==== JSON Parsing ====\\

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

type ValidationError struct {
	message string
}

func (e ValidationError) Error() string {
	return e.message
}

func NewValidationError(message string) error {
	return ValidationError{message}
}

//==== Bad Login ====\\

type BadLoginError struct{}

func (e BadLoginError) Error() string {
	return "bad login"
}

func NewBadLoginError() error {
	return BadLoginError{}
}
