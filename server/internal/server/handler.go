package server

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

// Handle serves h, writing any error it returns. Responses vary on Accept, since handlers answer with
// HTML or JSON depending on it.
func Handle(h HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Accept")
		if err := h(w, r); err != nil {
			WriteError(w, r, err)
		}
	})
}

// Template is a page that renders data as HTML.
type Template interface {
	Execute(w io.Writer, data any) error
}

// Respond writes data as JSON when the client asks for it, otherwise renders page with pageData.
func Respond(w http.ResponseWriter, r *http.Request, status int, page Template, pageData any, data any) error {
	if WantsJSON(r) {
		return WriteJSON(w, status, data)
	}
	return WriteHTML(w, status, page, pageData)
}

// WriteFormError renders a form page again with the error's message, from pageData. Errors for JSON
// clients, and errors that aren't the client's fault, are returned for Handle to write instead.
func WriteFormError(w http.ResponseWriter, r *http.Request, err error, page Template, pageData func(message string) any) error {
	status, message, ok := ClientError(err)
	if !ok || WantsJSON(r) {
		return err
	}
	return WriteHTML(w, status, page, pageData(message))
}

func WriteHTML(w http.ResponseWriter, status int, tmpl Template, data any) error {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return err
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(buf.Bytes())
	return nil
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	// marshal first, so a failure can still become an error response
	body, err := json.Marshal(v)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(append(body, '\n'))
	return nil
}
