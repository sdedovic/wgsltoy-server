package server

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

func Handle(h HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			WriteError(w, r, err)
		}
	})
}

func DecodeJSON(r *http.Request, v any) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return JSONParsingError{err}
	}
	return nil
}

func HandleHTML(h HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			WriteHTMLError(w, r, err)
		}
	})
}

// Template is a page that renders data as HTML.
type Template interface {
	Execute(w io.Writer, data any) error
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
