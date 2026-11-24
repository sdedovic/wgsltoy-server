package controllers

import (
	"net/http"

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

func (c *ShaderController) Register(router *server.Router) {
	router.Handle("GET", homeRoute, server.Handle(c.gallery))
	router.Handle("POST", shadersRoute, server.Handle(c.create))
	router.Handle("GET", shaderRoute, server.Handle(c.view))
	router.Handle("PUT", shaderRoute, server.Handle(c.update))
	router.Handle("DELETE", shaderRoute, server.Handle(c.delete))
	router.Handle("GET", legacyShaderRoute, http.HandlerFunc(c.moved))
	router.Handle("GET", legacyShaderEditorRoute, http.HandlerFunc(c.moved))
}

// how many random shaders the home page previews
const featuredCount = 6

type galleryPageData struct {
	Principal *domain.Principal
	Featured  []domain.Shader
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

// @Summary		List public shaders
// @Description	The 100 most recently updated public shaders. This is the home page in a browser.
// @Tags			shaders
// @Produce		json
// @Success		200	{object}	galleryDTO
// @Router			/ [get]
func (c *ShaderController) gallery(w http.ResponseWriter, r *http.Request) error {
	shaders, err := c.shaders.ListPublic(r.Context())
	if err != nil {
		return err
	}
	if server.WantsJSON(r) {
		return server.WriteJSON(w, http.StatusOK, newGalleryDTO(shaders))
	}
	featured, err := c.shaders.ListRandomPublic(r.Context(), featuredCount)
	if err != nil {
		return err
	}
	return server.WriteHTML(w, http.StatusOK, c.tmpl.Gallery, galleryPageData{
		Principal: principalPtr(r),
		Featured:  featured,
		Shaders:   shaders,
	})
}

// @Summary		Get a shader
// @Description	Anyone can get public and unlisted shaders, only the author and admins can get private ones.
// @Tags			shaders
// @Produce		json
// @Param			id	path		string	true	"Shader ID"	format(uuid)
// @Success		200	{object}	shaderDTO
// @Failure		404	{object}	server.ErrorResponse
// @Router			/shaders/{id} [get]
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
	return server.Respond(w, r, http.StatusOK, c.tmpl.Shader, shaderPageData{
		Principal: principalPtr(r),
		Shader:    shader,
		CanEdit:   canEdit,
	}, newShaderDTO(shader))
}

// @Summary		Old shader link
// @Description	Redirects to /shaders/{id}.
// @Tags			shaders
// @Param			id	path	string	true	"Shader ID"
// @Success		302
// @Deprecated
// @Router			/shader/{id} [get]
// @Router			/shader/{id}/edit [get]
func (c *ShaderController) moved(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, shaderRoute.Path(r.PathValue("id")), http.StatusFound)
}

// @Summary	Create a shader
// @Tags		shaders
// @Accept		json,x-www-form-urlencoded,mpfd
// @Produce	json
// @Param		shader	body		shaderInputDTO	true	"The new shader"
// @Success	201		{object}	shaderDTO
// @Header		201		{string}	Location	"The new shader's URL"
// @Failure	401		{object}	server.ErrorResponse
// @Failure	415		{object}	server.ErrorResponse
// @Failure	422		{object}	server.ErrorResponse
// @Security	SessionCookie
// @Router		/shaders [post]
func (c *ShaderController) create(w http.ResponseWriter, r *http.Request) error {
	var req shaderInputDTO
	if err := server.Decode(r, &req); err != nil {
		return err
	}
	shader, err := c.shaders.Create(r.Context(), domain.CreateShaderCommand{
		Name:        req.Name,
		Visibility:  req.Visibility,
		Description: req.Description,
		Tags:        req.Tags,
		Content:     req.Content,
	})
	if err != nil {
		return err
	}

	location := shaderPath(shader.ID)
	if server.WantsJSON(r) {
		w.Header().Set("Location", location)
		return server.WriteJSON(w, http.StatusCreated, newShaderDTO(shader))
	}
	http.Redirect(w, r, location, http.StatusSeeOther)
	return nil
}

// @Summary		Replace a shader
// @Description	Every field is replaced, including ones left out. Only the author can update a shader.
// @Tags			shaders
// @Accept			json,x-www-form-urlencoded,mpfd
// @Produce		json
// @Param			id		path		string			true	"Shader ID"	format(uuid)
// @Param			shader	body		shaderInputDTO	true	"The shader's new fields"
// @Success		200		{object}	shaderDTO
// @Failure		400		{object}	server.ErrorResponse
// @Failure		401		{object}	server.ErrorResponse
// @Failure		404		{object}	server.ErrorResponse
// @Failure		415		{object}	server.ErrorResponse
// @Failure		422		{object}	server.ErrorResponse
// @Security		SessionCookie
// @Router			/shaders/{id} [put]
func (c *ShaderController) update(w http.ResponseWriter, r *http.Request) error {
	id, err := pathShaderID(r)
	if err != nil {
		return err
	}
	var req shaderInputDTO
	if err := server.Decode(r, &req); err != nil {
		return err
	}
	shader, err := c.shaders.Update(r.Context(), id, domain.UpdateShaderCommand{
		Name:        req.Name,
		Visibility:  req.Visibility,
		Description: req.Description,
		Tags:        req.Tags,
		Content:     req.Content,
	})
	if err != nil {
		return err
	}

	if server.WantsJSON(r) {
		return server.WriteJSON(w, http.StatusOK, newShaderDTO(shader))
	}
	http.Redirect(w, r, shaderPath(shader.ID), http.StatusSeeOther)
	return nil
}

// @Summary		Delete a shader
// @Description	The author and admins can delete a shader.
// @Tags			shaders
// @Produce		json
// @Param			id	path	string	true	"Shader ID"	format(uuid)
// @Success		204
// @Failure		401	{object}	server.ErrorResponse
// @Failure		404	{object}	server.ErrorResponse
// @Security		SessionCookie
// @Router			/shaders/{id} [delete]
func (c *ShaderController) delete(w http.ResponseWriter, r *http.Request) error {
	id, err := pathShaderID(r)
	if err != nil {
		return err
	}
	if err := c.shaders.Delete(r.Context(), id); err != nil {
		return err
	}

	if server.WantsJSON(r) {
		w.WriteHeader(http.StatusNoContent)
		return nil
	}
	http.Redirect(w, r, homeRoute.Path(), http.StatusSeeOther)
	return nil
}

func pathShaderID(r *http.Request) (domain.ShaderID, error) {
	id, err := domain.NewShaderID(r.PathValue("id"))
	if err != nil {
		return "", domain.ErrNotFound
	}
	return id, nil
}
