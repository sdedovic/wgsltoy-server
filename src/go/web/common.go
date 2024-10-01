package web

import (
	"context"
	"net/http"
)

// StartContext extracts the root context from the incoming HTTP request to be passed down to subsequent processing
func StartContext(r *http.Request) context.Context {
	return context.Background()
}

func CreateHandler(handler func(context.Context, http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := StartContext(r)

		err := handler(ctx, w, r)

		w.Header().Set("X-Content-Type-Options", "nosniff")

		if err != nil {
			WriteErrorResponse(w, err)
		}
	}
}
