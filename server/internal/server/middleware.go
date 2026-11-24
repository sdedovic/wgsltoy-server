package server

import (
	"context"
	"net/http"

	"go.uber.org/zap"

	"github.com/sdedovic/wgsltoy-server/internal/domain"
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

var ErrCrossOrigin = RequestError{http.StatusForbidden, "CROSS_ORIGIN", "Cross-origin requests that change data are not allowed."}

// withCrossOriginProtection rejects browser requests from other sites that aren't GET, HEAD or OPTIONS,
// since the session cookie would go with them.
func withCrossOriginProtection(next http.Handler) http.Handler {
	protection := http.NewCrossOriginProtection()
	protection.SetDenyHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		WriteError(w, r, ErrCrossOrigin)
	}))
	return protection.Handler(next)
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
