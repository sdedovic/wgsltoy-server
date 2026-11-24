package controllers

import (
	"net/http"
	"strings"

	"github.com/sdedovic/wgsltoy-server/internal/domain"
	"github.com/sdedovic/wgsltoy-server/internal/server"
	"github.com/sdedovic/wgsltoy-server/internal/views"
)

type ShaderController struct {
	shaders *domain.ShaderService
	tmpl    *views.Templates
}

func NewShaderController(shaders *domain.ShaderService, tmpl *views.Templates) *ShaderController {
	return &ShaderController{shaders, tmpl}
}

func (c *ShaderController) Register(mux *http.ServeMux) {
	mux.Handle("GET /", server.HandleHTML(c.gallery))
	mux.Handle("POST /shader", server.HandleHTML(c.create))
	mux.Handle("GET /shader/{id}", server.HandleHTML(c.view))
	mux.Handle("PUT /shader/{id}", server.HandleHTML(c.update))
	mux.Handle("DELETE /shader/{id}", server.HandleHTML(c.delete))
	// the editor moved onto the shader page, this keeps old links working
	mux.Handle("GET /shader/{id}/edit", http.HandlerFunc(c.editor))
}

type galleryPageData struct {
	Principal *domain.Principal
	Shaders   []domain.ShaderSummary
}

type shaderPageData struct {
	Principal *domain.Principal
	Shader    domain.Shader
	CanEdit   bool
}

func principalPtr(r *http.Request) *domain.Principal {
	if p, ok := domain.PrincipalFromContext(r.Context()); ok {
		return &p
	}
	return nil
}

func (c *ShaderController) gallery(w http.ResponseWriter, r *http.Request) error {
	shaders, err := c.shaders.ListPublic(r.Context())
	if err != nil {
		return err
	}
	return server.WriteHTML(w, http.StatusOK, c.tmpl.Gallery, galleryPageData{
		Principal: principalPtr(r),
		Shaders:   shaders,
	})
}

func (c *ShaderController) view(w http.ResponseWriter, r *http.Request) error {
	id, err := pathShaderID(r)
	if err != nil {
		return err
	}
	shader, err := c.shaders.Get(r.Context(), id)
	if err != nil {
		return err
	}
	canEdit := false
	if p, ok := domain.PrincipalFromContext(r.Context()); ok {
		canEdit = shader.IsEditableBy(p)
	}
	return server.WriteHTML(w, http.StatusOK, c.tmpl.Shader, shaderPageData{
		Principal: principalPtr(r),
		Shader:    shader,
		CanEdit:   canEdit,
	})
}

func (c *ShaderController) editor(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/shader/"+r.PathValue("id"), http.StatusFound)
}

func (c *ShaderController) create(w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return err
	}
	id, err := c.shaders.Create(r.Context(), domain.CreateShaderCommand{
		Name:        r.FormValue("name"),
		Visibility:  r.FormValue("visibility"),
		Description: r.FormValue("description"),
		Tags:        parseTags(r.FormValue("tags")),
		Content:     r.FormValue("content"),
	})
	if err != nil {
		return err
	}
	http.Redirect(w, r, "/shader/"+string(id), http.StatusSeeOther)
	return nil
}

func (c *ShaderController) update(w http.ResponseWriter, r *http.Request) error {
	id, err := pathShaderID(r)
	if err != nil {
		return err
	}
	if err := r.ParseForm(); err != nil {
		return err
	}
	shader, err := c.shaders.Update(r.Context(), id, domain.UpdateShaderCommand{
		Name:        r.FormValue("name"),
		Visibility:  r.FormValue("visibility"),
		Description: r.FormValue("description"),
		Tags:        parseTags(r.FormValue("tags")),
		Content:     r.FormValue("content"),
	})
	if err != nil {
		return err
	}
	return server.WriteHTML(w, http.StatusOK, c.tmpl.Shader, shaderPageData{
		Principal: principalPtr(r),
		Shader:    shader,
		CanEdit:   true,
	})
}

func (c *ShaderController) delete(w http.ResponseWriter, r *http.Request) error {
	id, err := pathShaderID(r)
	if err != nil {
		return err
	}
	if err := c.shaders.Delete(r.Context(), id); err != nil {
		return err
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}

func pathShaderID(r *http.Request) (domain.ShaderID, error) {
	id, err := domain.NewShaderID(r.PathValue("id"))
	if err != nil {
		return "", domain.ErrNotFound
	}
	return id, nil
}

func parseTags(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	tags := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			tags = append(tags, t)
		}
	}
	return tags
}
