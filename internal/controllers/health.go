package controllers

import (
	"net/http"

	"github.com/sdedovic/wgsltoy-server/internal/server"
)

type HealthController struct{}

func NewHealthController() *HealthController {
	return &HealthController{}
}

func (c *HealthController) Register(mux *http.ServeMux) {
	mux.Handle("GET /api/v1/health", server.Handle(c.health))
}

type healthResponse struct {
	Status string `json:"status"`
}

//	@Summary	Health check
//	@Tags		health
//	@Produce	json
//	@Success	200	{object}	healthResponse
//	@Router		/api/v1/health [get]
func (c *HealthController) health(w http.ResponseWriter, r *http.Request) error {
	return server.WriteJSON(w, http.StatusOK, healthResponse{"ok"})
}
