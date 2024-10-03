package web

import (
	"context"
	"github.com/sdedovic/wgsltoy-server/src/go/infra"
	"github.com/sdedovic/wgsltoy-server/src/go/service"
	"net/http"
	"strings"
)

// StartContext extracts the root context from the incoming HTTP request to be passed down to subsequent processing
func StartContext(r *http.Request) context.Context {
	return context.Background()
}

func Handler(handler func(context.Context, http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := StartContext(r)

		err := handler(ctx, w, r)

		w.Header().Set("X-Content-Type-Options", "nosniff")
		if err != nil {
			WriteErrorResponse(w, err)
		}
	}
}

func AuthenticatedHandler(handler func(context.Context, http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := StartContext(r)

		authorizationHeader := r.Header.Get("authorization")
		if authorizationHeader == "" {
			WriteErrorResponse(w, infra.UnauthorizedError)
			return
		}

		parts := strings.Split(authorizationHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			WriteErrorResponse(w, infra.UnauthorizedError)
			return
		}

		user, err := service.ParseToken(parts[1])
		if err != nil {
			WriteErrorResponse(w, err)
			return
		}

		ctx = context.WithValue(ctx, "user", user)

		err = handler(ctx, w, r)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		if err != nil {
			WriteErrorResponse(w, err)
		}
	}
}
