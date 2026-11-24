package server

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// Route is a ServeMux path pattern, like "/shaders/{id}". Handlers are served at it, and links to them
// are built from it with Path, so the two can't drift apart.
type Route string

// Path is the route with its wildcards filled in from params, in order, each escaped. A {name...}
// wildcard keeps the slashes in its param. Path panics when the number of params is wrong, since that's
// a mistake in the code rather than in a request.
func (rt Route) Path(params ...string) string {
	// nilaway can't see that the count check below guards indexing a nil params
	if params == nil {
		params = []string{}
	}
	segments := strings.Split(string(rt), "/")
	used := 0
	for i, segment := range segments {
		if !strings.HasPrefix(segment, "{") || !strings.HasSuffix(segment, "}") {
			continue
		}
		name := segment[1 : len(segment)-1]
		if name == "$" {
			segments[i] = ""
			continue
		}
		if used == len(params) {
			panic(fmt.Sprintf("route %s: %d params given for more wildcards", rt, len(params)))
		}
		param := params[used]
		used++

		if strings.HasSuffix(name, "...") {
			parts := strings.Split(param, "/")
			for j, part := range parts {
				parts[j] = url.PathEscape(part)
			}
			segments[i] = strings.Join(parts, "/")
		} else {
			segments[i] = url.PathEscape(param)
		}
	}
	if used != len(params) {
		panic(fmt.Sprintf("route %s: %d params given for %d wildcards", rt, len(params), used))
	}
	return strings.Join(segments, "/")
}

// Registration is a method and route that a handler is served at.
type Registration struct {
	Method string
	Route  Route
}

// Router serves handlers by method and Route.
type Router struct {
	mux    *http.ServeMux
	routes []Registration
}

func NewRouter() *Router {
	return &Router{mux: http.NewServeMux()}
}

func (rt *Router) Handle(method string, route Route, h http.Handler) {
	rt.mux.Handle(method+" "+string(route), h)
	rt.routes = append(rt.routes, Registration{method, route})
}

// Routes are the registrations so far, in the order they were made.
func (rt *Router) Routes() []Registration {
	return append([]Registration(nil), rt.routes...)
}

func (rt *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rt.mux.ServeHTTP(w, r)
}
