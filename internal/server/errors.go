package server

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/sdedovic/wgsltoy-server/internal/domain"
	"go.uber.org/zap"
)

type JSONParsingError struct {
	cause error
}

func (e JSONParsingError) Error() string {
	return fmt.Sprintf("Unable to parse JSON, caused by: %v", e.cause)
}

func (e JSONParsingError) Unwrap() error {
	return e.cause
}

type ErrorResponse struct {
	Class   string `json:"errorClass"`
	Message string `json:"causedBy"`
}

func WriteHTMLError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrUnauthorized):
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	case errors.Is(err, domain.ErrNotFound):
		http.Error(w, "404 not found", http.StatusNotFound)
	default:
		Logger(r.Context()).Error("request failed", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	var validationError domain.ValidationError
	var jsonParsingError JSONParsingError

	status, response := http.StatusInternalServerError, ErrorResponse{"UNKNOWN", "An unexpected error occurred!"}
	switch {
	case errors.As(err, &validationError):
		status, response = http.StatusBadRequest, ErrorResponse{"VALIDATION_FAILURE", err.Error()}
	case errors.Is(err, domain.ErrBadLogin):
		status, response = http.StatusBadRequest, ErrorResponse{"BAD_LOGIN", "Either 'username' or 'password' are incorrect."}
	case errors.Is(err, domain.ErrUnauthorized):
		status, response = http.StatusUnauthorized, ErrorResponse{"UNAUTHORIZED", "This resource requires authorization."}
	case errors.Is(err, domain.ErrNotFound):
		status, response = http.StatusNotFound, ErrorResponse{"NOT_FOUND", "The requested resource was not found."}
	case errors.As(err, &jsonParsingError):
		status, response = http.StatusBadRequest, ErrorResponse{"JSON_PARSING_FAILURE", "Failed to parse JSON payload."}
	default:
		// only unexpected errors are logged, client errors would be noise
		Logger(r.Context()).Error("request failed", zap.Error(err))
	}

	if err := WriteJSON(w, status, response); err != nil {
		Logger(r.Context()).Error("failed writing error response", zap.Error(err))
	}
}
