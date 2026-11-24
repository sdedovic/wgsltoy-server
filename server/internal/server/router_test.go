package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoute_Path(t *testing.T) {
	tests := []struct {
		route  Route
		params []string
		want   string
	}{
		{"/", nil, "/"},
		{"/{$}", nil, "/"},
		{"/shaders", nil, "/shaders"},
		{"/static/assets/", nil, "/static/assets/"},
		{"/shaders/{id}", []string{"abc"}, "/shaders/abc"},
		{"/shaders/{id}/edit", []string{"abc"}, "/shaders/abc/edit"},
		{"/users/{username}", []string{"a b/c?d"}, "/users/a%20b%2Fc%3Fd"},
		{"/a/{x}/b/{y}", []string{"1", "2"}, "/a/1/b/2"},
		{"/files/{path...}", []string{"a/b c/d"}, "/files/a/b%20c/d"},
	}
	for _, tc := range tests {
		t.Run(string(tc.route), func(t *testing.T) {
			assert.Equal(t, tc.want, tc.route.Path(tc.params...))
		})
	}
}

func TestRoute_PathPanicsOnWrongParams(t *testing.T) {
	assert.Panics(t, func() { Route("/shaders/{id}").Path() })
	assert.Panics(t, func() { Route("/shaders/{id}").Path("a", "b") })
	assert.Panics(t, func() { Route("/shaders").Path("a") })
}

// A path built from a route is served by that route, with the same params back out.
func TestRouter_PathRoundTrips(t *testing.T) {
	for _, param := range []string{"plain", "with space", "with/slash", "with?query#frag", "üñí", "%41"} {
		t.Run(param, func(t *testing.T) {
			route := Route("/users/{username}")
			var got string
			router := NewRouter()
			router.Handle("GET", route, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				got = r.PathValue("username")
			}))

			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest("GET", route.Path(param), nil))

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, param, got)
		})
	}
}

func TestRouter_MatchesMethod(t *testing.T) {
	router := NewRouter()
	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) })
	router.Handle("PUT", "/shaders/{id}", ok)
	router.Handle("DELETE", "/shaders/{id}", ok)

	for method, want := range map[string]int{
		"PUT":    http.StatusNoContent,
		"DELETE": http.StatusNoContent,
		"POST":   http.StatusMethodNotAllowed,
	} {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(method, "/shaders/abc", nil))
		assert.Equal(t, want, w.Code, method)
	}
}

func TestRouter_Routes(t *testing.T) {
	router := NewRouter()
	router.Handle("GET", "/shaders/{id}", http.NotFoundHandler())
	router.Handle("PUT", "/shaders/{id}", http.NotFoundHandler())

	assert.Equal(t, []Registration{{"GET", "/shaders/{id}"}, {"PUT", "/shaders/{id}"}}, router.Routes())
}
