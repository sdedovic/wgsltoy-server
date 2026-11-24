package server

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

func NewHandler(log *zap.Logger, auth *AuthMiddleware, controllers []Controller) http.Handler {
	mux := http.NewServeMux()
	for _, controller := range controllers {
		controller.Register(mux)
	}

	// TODO: JSON bodies for the mux's own 404 and 405 responses
	return withLogger(log, withSecurityHeaders(auth.Handler(mux)))
}

func New(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}
}
