package server

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWantsJSON(t *testing.T) {
	tests := []struct {
		accept string
		want   bool
	}{
		{"", false},
		// what fetch sends by default
		{"*/*", false},
		{"text/html", false},
		{"application/json", true},
		{"application/JSON", true},
		{"application/*", true},
		// what browsers send for a page
		{"text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8", false},
		{"application/json, */*", true},
		{"application/json, text/plain, */*", true},
		{"application/json;q=0.5, text/html", false},
		{"text/html;q=0.5, application/json", true},
		{"application/json;q=0", false},
		{"application/json;q=nonsense", false},
		{"text/html, application/json", false},
		{"text/*, application/json", true},
		{"image/png", false},
	}
	for _, tc := range tests {
		t.Run(tc.accept, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", nil)
			if tc.accept != "" {
				r.Header.Set("Accept", tc.accept)
			}
			assert.Equal(t, tc.want, WantsJSON(r))
		})
	}
}
