package server

import (
	"errors"
	"net/http"

	"go.uber.org/zap"

	"github.com/sdedovic/wgsltoy-server/internal/domain"
)

// RequestError is a client error found before the request reaches the domain, like an unreadable body.
type RequestError struct {
	Status  int
	Class   string
	Message string
}

func (e RequestError) Error() string {
	return e.Message
}

var ErrUnsupportedMediaType = RequestError{http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE",
	"Content-Type must be application/json, application/x-www-form-urlencoded or multipart/form-data."}

var ErrBodyTooLarge = RequestError{http.StatusRequestEntityTooLarge, "BODY_TOO_LARGE", "Request body is too large."}

var ErrNotAcceptable = RequestError{http.StatusNotAcceptable, "NOT_ACCEPTABLE", "This resource is only available as HTML."}

type ErrorResponse struct {
	Class   string `json:"errorClass"`
	Message string `json:"causedBy"`
} // @name Error

// ClientError is the status and message for an error the client caused. ok is false for anything else,
// which is a server error and shouldn't be shown.
func ClientError(err error) (status int, message string, ok bool) {
	status, response := classify(err)
	return status, response.Message, status < http.StatusInternalServerError
}

func classify(err error) (int, ErrorResponse) {
	var requestError RequestError
	var validationError domain.ValidationError
	switch {
	case errors.As(err, &requestError):
		return requestError.Status, ErrorResponse{requestError.Class, requestError.Message}
	case errors.Is(err, domain.ErrUsernameTaken), errors.Is(err, domain.ErrEmailTaken):
		return http.StatusConflict, ErrorResponse{"CONFLICT", err.Error()}
	case errors.As(err, &validationError):
		return http.StatusUnprocessableEntity, ErrorResponse{"VALIDATION_FAILURE", err.Error()}
	case errors.Is(err, domain.ErrBadLogin):
		return http.StatusUnauthorized, ErrorResponse{"BAD_LOGIN", "Either 'username' or 'password' are incorrect."}
	case errors.Is(err, domain.ErrUnauthorized):
		return http.StatusUnauthorized, ErrorResponse{"UNAUTHORIZED", "This resource requires authorization."}
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound, ErrorResponse{"NOT_FOUND", "The requested resource was not found."}
	default:
		return http.StatusInternalServerError, ErrorResponse{"UNKNOWN", "An unexpected error occurred!"}
	}
}

// WriteError writes err as JSON when the client asks for it, otherwise as a plain page. A browser that
// needs to log in is sent to the login page instead.
func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	status, response := classify(err)
	if status >= http.StatusInternalServerError {
		// only unexpected errors are logged, client errors would be noise
		Logger(r.Context()).Error("request failed", zap.Error(err))
	}

	if WantsJSON(r) {
		if err := WriteJSON(w, status, response); err != nil {
			Logger(r.Context()).Error("failed writing error response", zap.Error(err))
		}
		return
	}
	if errors.Is(err, domain.ErrUnauthorized) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	http.Error(w, response.Message, status)
}
