package server

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

func NewHandler(log *zap.Logger, auth *AuthMiddleware, controllers []Controller) http.Handler {
	router := NewRouter()
	for _, controller := range controllers {
		controller.Register(router)
	}

	// TODO: JSON bodies for the mux's own 404 and 405 responses
	return withLogger(log, withSecurityHeaders(withCrossOriginProtection(auth.Handler(router))))
}

func New(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}
}
