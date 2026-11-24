package server

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type decodeTarget struct {
	Name    string   `json:"name"`
	Tags    []string `json:"tags"`
	Ignored string   `json:"-"`
	Missing string   `json:"missing"`
}

func request(contentType, body string) *http.Request {
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	if contentType != "" {
		r.Header.Set("Content-Type", contentType)
	}
	return r
}

func TestDecode_JSON(t *testing.T) {
	var got decodeTarget
	err := Decode(request("application/json; charset=utf-8", `{"name":"a","tags":["b","c"]}`), &got)
	require.NoError(t, err)
	assert.Equal(t, decodeTarget{Name: "a", Tags: []string{"b", "c"}}, got)
}

func TestDecode_URLEncoded(t *testing.T) {
	var got decodeTarget
	err := Decode(request("application/x-www-form-urlencoded", "name=a&tags=b&tags=c&-=x"), &got)
	require.NoError(t, err)
	assert.Equal(t, decodeTarget{Name: "a", Tags: []string{"b", "c"}}, got)
}

func TestDecode_Multipart(t *testing.T) {
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	require.NoError(t, mw.WriteField("name", "a"))
	require.NoError(t, mw.WriteField("tags", "b"))
	require.NoError(t, mw.Close())

	var got decodeTarget
	err := Decode(request(mw.FormDataContentType(), body.String()), &got)
	require.NoError(t, err)
	assert.Equal(t, decodeTarget{Name: "a", Tags: []string{"b"}}, got)
}

func TestDecode_Errors(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		body        string
		status      int
	}{
		{"no content type", "", `{}`, http.StatusUnsupportedMediaType},
		{"unsupported content type", "text/plain", `{}`, http.StatusUnsupportedMediaType},
		{"empty json", "application/json", ``, http.StatusBadRequest},
		{"malformed json", "application/json", `{"name":`, http.StatusBadRequest},
		{"wrong json type", "application/json", `{"tags":"b"}`, http.StatusUnprocessableEntity},
		{"too large", "application/json", `{"name":"` + strings.Repeat("a", maxJSONBody) + `"}`, http.StatusRequestEntityTooLarge},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var got decodeTarget
			err := Decode(request(tc.contentType, tc.body), &got)
			status, _, ok := ClientError(err)
			assert.True(t, ok)
			assert.Equal(t, tc.status, status)
		})
	}
}
