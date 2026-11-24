package controllers

import (
	"net/http"
	"path/filepath"

	"github.com/sdedovic/wgsltoy-server/internal/server"
	"github.com/sdedovic/wgsltoy-server/internal/views"
)

// StaticController serves the frontend built by Vite. Only assets/ is served, not e.g. the manifest.
// Vite puts a content hash in every asset's file name, so they can be cached forever.
type StaticController struct {
	assets http.Handler
}

func NewStaticController(distDir string) *StaticController {
	files := http.FileServer(http.Dir(filepath.Join(distDir, "assets")))
	return &StaticController{http.StripPrefix(views.AssetBase+"assets/", files)}
}

func (c *StaticController) Register(router *server.Router) {
	router.Handle("GET", server.Route(views.AssetBase+"assets/"), http.HandlerFunc(c.serve))
}

func (c *StaticController) serve(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	c.assets.ServeHTTP(w, r)
}
