package controllers

import (
	"net/http"

	"github.com/sdedovic/wgsltoy-server/internal/server"
)

type HealthController struct{}

func NewHealthController() *HealthController {
	return &HealthController{}
}

func (c *HealthController) Register(router *server.Router) {
	router.Handle("GET", healthRoute, server.Handle(c.health))
}

// @Summary	Health check
// @Tags		health
// @Produce	json
// @Success	200	{object}	healthDTO
// @Router		/health [get]
// Always JSON, it's for probes rather than people.
func (c *HealthController) health(w http.ResponseWriter, r *http.Request) error {
	return server.WriteJSON(w, http.StatusOK, healthDTO{"ok"})
}
