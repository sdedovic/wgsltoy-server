package server

import (
	"context"
	"net/http"
	"strings"

	"github.com/sdedovic/wgsltoy-server/internal/domain"
	"go.uber.org/zap"
)

type loggerKey struct{}

func Logger(ctx context.Context) *zap.Logger {
	if log, ok := ctx.Value(loggerKey{}).(*zap.Logger); ok {
		return log
	}
	return zap.NewNop()
}

func withLogger(log *zap.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), loggerKey{}, log)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func withSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, r)
	})
}

// Requests without a token pass through, since some routes allow anonymous access.
func withAuthentication(tokens domain.TokenVerifier, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			next.ServeHTTP(w, r)
			return
		}

		token, ok := strings.CutPrefix(header, "Bearer ")
		if !ok {
			WriteError(w, r, domain.ErrUnauthorized)
			return
		}

		identity, err := tokens.Verify(token)
		if err != nil {
			WriteError(w, r, err)
			return
		}

		next.ServeHTTP(w, r.WithContext(domain.WithIdentity(r.Context(), identity)))
	})
}
