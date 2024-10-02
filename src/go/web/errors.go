package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sdedovic/wgsltoy-server/src/go/infra"
	"log"
	"net/http"
	"os"
	"strings"
)

type UnsupportedOperationError struct {
	allow []string
}

func (e UnsupportedOperationError) Error() string {
	return fmt.Sprintf("Supported operations are: [%s]", strings.ToUpper(strings.Join(e.allow, ", ")))
}

func NewUnsupportedOperationError(allow ...string) error {
	return UnsupportedOperationError{allow}
}

type ErrorDto struct {
	Class   string `json:"errorClass"`
	Message string `json:"causedBy"`
}

func WriteErrorResponse(w http.ResponseWriter, in error) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	var err error

	var validationError infra.ValidationError
	var badLoginError infra.BadLoginError
	var unsupportedOperationError UnsupportedOperationError
	var jsonParsingError infra.JsonParsingError
	switch {
	case errors.As(in, &validationError):
		w.WriteHeader(http.StatusBadRequest)
		err = json.NewEncoder(w).Encode(ErrorDto{"VALIDATION_FAILURE", in.Error()})
	case errors.As(in, &badLoginError):
		w.WriteHeader(http.StatusBadRequest)
		err = json.NewEncoder(w).Encode(ErrorDto{"BAD_LOGIN", "Either 'username' or 'password' are incorrect."})
	case errors.As(in, &jsonParsingError):
		w.WriteHeader(http.StatusBadRequest)
		err = json.NewEncoder(w).Encode(ErrorDto{"JSON_PARSING_FAILURE", "Failed to parse JSON payload"})
	case errors.As(in, &unsupportedOperationError):
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Header().Set("Allow", strings.ToUpper(strings.Join(unsupportedOperationError.allow, ", ")))
		err = json.NewEncoder(w).Encode(ErrorDto{"UNSUPPORTED_OPERATION", in.Error()})
	default:
		log.Println("ERROR", in.Error())
		w.WriteHeader(http.StatusInternalServerError)
		err = json.NewEncoder(w).Encode(ErrorDto{"UNKNOWN", "An unexpected error occurred!"})
	}

	if err != nil {
		log.Println("ERROR", err.Error())
		os.Exit(1)
	}
}
