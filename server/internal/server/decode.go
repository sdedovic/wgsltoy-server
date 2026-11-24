package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"reflect"
	"strings"
)

const maxJSONBody = 1 << 20

// Decode reads the request body into v, a pointer to a struct, from JSON or form fields. Form fields
// are matched by each struct field's json name, and only string and []string fields can be set.
func Decode(r *http.Request, v any) error {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		return ErrUnsupportedMediaType
	}

	switch mediaType {
	case "application/json":
		return decodeJSON(r, v)
	case "application/x-www-form-urlencoded", "multipart/form-data":
		return decodeForm(r, v)
	default:
		return ErrUnsupportedMediaType
	}
}

func decodeJSON(r *http.Request, v any) error {
	body := http.MaxBytesReader(nil, r.Body, maxJSONBody)
	err := json.NewDecoder(body).Decode(v)

	var tooLarge *http.MaxBytesError
	var typeErr *json.UnmarshalTypeError
	switch {
	case err == nil:
		return nil
	case errors.As(err, &tooLarge):
		return ErrBodyTooLarge
	case errors.As(err, &typeErr):
		// well-formed, but a value of the wrong type, which is a validation failure like any other
		return RequestError{http.StatusUnprocessableEntity, "VALIDATION_FAILURE",
			fmt.Sprintf("Field '%s' must be of type %s!", typeErr.Field, jsonTypeName(typeErr.Type))}
	case errors.Is(err, io.EOF):
		return RequestError{http.StatusBadRequest, "JSON_PARSING_FAILURE", "Request body is empty."}
	default:
		return RequestError{http.StatusBadRequest, "JSON_PARSING_FAILURE", "Failed to parse JSON payload."}
	}
}

func jsonTypeName(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Slice, reflect.Array:
		return "array"
	case reflect.Struct, reflect.Map:
		return "object"
	case reflect.Bool:
		return "boolean"
	case reflect.String:
		return "string"
	default:
		return "number"
	}
}

func decodeForm(r *http.Request, v any) error {
	if err := r.ParseMultipartForm(maxJSONBody); err != nil && !errors.Is(err, http.ErrNotMultipart) {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			return ErrBodyTooLarge
		}
		return RequestError{http.StatusBadRequest, "FORM_PARSING_FAILURE", "Failed to parse form payload."}
	}

	target := reflect.ValueOf(v).Elem()
	for i := range target.NumField() {
		field := target.Type().Field(i)
		name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
		if name == "" || name == "-" {
			continue
		}
		values, ok := r.Form[name]
		if !ok {
			continue
		}
		switch field.Type {
		case reflect.TypeFor[string]():
			target.Field(i).SetString(values[0])
		case reflect.TypeFor[[]string]():
			target.Field(i).Set(reflect.ValueOf(values))
		default:
			panic(fmt.Sprintf("server.Decode: form field %s has unsupported type %s", name, field.Type))
		}
	}
	return nil
}
