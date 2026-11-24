package server

import (
	"context"
	"net/http"

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

type AuthMiddleware struct {
	auth domain.AuthService
}

func NewAuthMiddleware(auth domain.AuthService) *AuthMiddleware {
	return &AuthMiddleware{auth}
}

// Requests without a session cookie pass through as anonymous; authenticated routes check PrincipalFromContext themselves.
func (m *AuthMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookieName)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		principal, err := m.auth.Authenticate(r.Context(), cookie.Value)
		if err != nil {
			ClearSessionCookie(w)
			next.ServeHTTP(w, r)
			return
		}

		next.ServeHTTP(w, r.WithContext(domain.WithPrincipal(r.Context(), principal)))
	})
}
