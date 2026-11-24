package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCrossOriginProtection(t *testing.T) {
	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) })
	handler := withCrossOriginProtection(ok)

	tests := []struct {
		name    string
		method  string
		headers map[string]string
		want    int
	}{
		{"same origin", "PUT", map[string]string{"Sec-Fetch-Site": "same-origin"}, http.StatusNoContent},
		{"typed into the address bar", "POST", map[string]string{"Sec-Fetch-Site": "none"}, http.StatusNoContent},
		{"not a browser", "DELETE", nil, http.StatusNoContent},
		{"cross site", "POST", map[string]string{"Sec-Fetch-Site": "cross-site"}, http.StatusForbidden},
		{"other origin, older browser", "PUT", map[string]string{"Origin": "https://evil.example"}, http.StatusForbidden},
		{"cross site read", "GET", map[string]string{"Sec-Fetch-Site": "cross-site"}, http.StatusNoContent},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, "http://wgsltoy.com/me", nil)
			for k, v := range tc.headers {
				r.Header.Set(k, v)
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)
			assert.Equal(t, tc.want, w.Code)
		})
	}
}

func TestCrossOriginProtection_JSONError(t *testing.T) {
	r := httptest.NewRequest("POST", "http://wgsltoy.com/shaders", nil)
	r.Header.Set("Sec-Fetch-Site", "cross-site")
	r.Header.Set("Accept", "application/json")
	w := httptest.NewRecorder()
	withCrossOriginProtection(http.NotFoundHandler()).ServeHTTP(w, r)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), `"errorClass":"CROSS_ORIGIN"`)
}
