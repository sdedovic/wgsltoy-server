package controllers

import (
	"net/http"
	"time"

	"github.com/sdedovic/wgsltoy-server/internal/domain"
	"github.com/sdedovic/wgsltoy-server/internal/server"
)

type ShaderController struct {
	shaders *domain.ShaderService
}

func NewShaderController(shaders *domain.ShaderService) *ShaderController {
	return &ShaderController{shaders}
}

func (c *ShaderController) Register(mux *http.ServeMux) {
	mux.Handle("POST /api/v1/shader", server.Handle(c.create))
	mux.Handle("GET /api/v1/shader/{id}", server.Handle(c.get))
	mux.Handle("PUT /api/v1/shader/{id}", server.Handle(c.update))
	mux.Handle("GET /api/v1/user/me/shader", server.Handle(c.listMine))
}

type createShaderRequest struct {
	Name        string   `json:"name"`
	Visibility  string   `json:"visibility"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Content     string   `json:"content"`
}

type updateShaderRequest struct {
	Name        *string   `json:"name"`
	Visibility  *string   `json:"visibility"`
	Description *string   `json:"description"`
	Tags        *[]string `json:"tags"`
	Content     *string   `json:"content"`
}

type shaderSummaryResponse struct {
	ID          string    `json:"id"`
	Location    string    `json:"location"`
	CreatedBy   string    `json:"createdBy"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Name        string    `json:"name"`
	Visibility  string    `json:"visibility"`
	Description string    `json:"description"`
	Tags        []string  `json:"tags"`
}

type shaderResponse struct {
	shaderSummaryResponse
	Content string `json:"content"`
}

func shaderLocation(id domain.ShaderID) string {
	return "/api/v1/shader/" + string(id)
}

func toShaderSummaryResponse(s domain.ShaderSummary) shaderSummaryResponse {
	return shaderSummaryResponse{
		ID:          string(s.ID),
		Location:    shaderLocation(s.ID),
		CreatedBy:   string(s.CreatedBy),
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
		Name:        string(s.Name),
		Visibility:  string(s.Visibility),
		Description: string(s.Description),
		Tags:        s.Tags,
	}
}

func toShaderResponse(s domain.Shader) shaderResponse {
	return shaderResponse{toShaderSummaryResponse(s.ShaderSummary), string(s.Content)}
}

// A malformed ID can't match any shader, so it's a 404.
func pathShaderID(r *http.Request) (domain.ShaderID, error) {
	id, err := domain.NewShaderID(r.PathValue("id"))
	if err != nil {
		return "", domain.ErrNotFound
	}
	return id, nil
}

//	@Summary	Create a shader
//	@Tags		shader
//	@Accept		json
//	@Security	BearerAuth
//	@Param		body	body	createShaderRequest	true	"New shader"
//	@Success	201
//	@Header		201	{string}	Location	"URL of the new shader"
//	@Failure	400	{object}	server.ErrorResponse
//	@Failure	401	{object}	server.ErrorResponse
//	@Router		/api/v1/shader [post]
func (c *ShaderController) create(w http.ResponseWriter, r *http.Request) error {
	var req createShaderRequest
	if err := server.DecodeJSON(r, &req); err != nil {
		return err
	}

	id, err := c.shaders.Create(r.Context(), domain.CreateShaderCommand{
		Name:        req.Name,
		Visibility:  req.Visibility,
		Description: req.Description,
		Tags:        req.Tags,
		Content:     req.Content,
	})
	if err != nil {
		return err
	}

	w.Header().Set("Location", shaderLocation(id))
	w.WriteHeader(http.StatusCreated)
	return nil
}

//	@Summary		Get a shader
//	@Description	Anonymous users see public and unlisted shaders. Logged-in users also see their own private ones.
//	@Tags			shader
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Shader ID"
//	@Success		200	{object}	shaderResponse
//	@Failure		404	{object}	server.ErrorResponse
//	@Router			/api/v1/shader/{id} [get]
func (c *ShaderController) get(w http.ResponseWriter, r *http.Request) error {
	id, err := pathShaderID(r)
	if err != nil {
		return err
	}

	shader, err := c.shaders.Get(r.Context(), id)
	if err != nil {
		return err
	}

	return server.WriteJSON(w, http.StatusOK, toShaderResponse(shader))
}

//	@Summary		Update a shader
//	@Description	Partial update. Omitted fields are left unchanged.
//	@Tags			shader
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string				true	"Shader ID"
//	@Param			body	body		updateShaderRequest	true	"Changes"
//	@Success		200		{object}	shaderResponse
//	@Failure		400		{object}	server.ErrorResponse
//	@Failure		401		{object}	server.ErrorResponse
//	@Failure		404		{object}	server.ErrorResponse
//	@Router			/api/v1/shader/{id} [put]
func (c *ShaderController) update(w http.ResponseWriter, r *http.Request) error {
	id, err := pathShaderID(r)
	if err != nil {
		return err
	}

	var req updateShaderRequest
	if err := server.DecodeJSON(r, &req); err != nil {
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

	return server.WriteJSON(w, http.StatusOK, toShaderResponse(shader))
}

//	@Summary	List the current user's shaders
//	@Tags		shader
//	@Produce	json
//	@Security	BearerAuth
//	@Success	200	{array}		shaderSummaryResponse
//	@Failure	401	{object}	server.ErrorResponse
//	@Router		/api/v1/user/me/shader [get]
func (c *ShaderController) listMine(w http.ResponseWriter, r *http.Request) error {
	shaders, err := c.shaders.ListMine(r.Context())
	if err != nil {
		return err
	}

	response := make([]shaderSummaryResponse, len(shaders))
	for i, shader := range shaders {
		response[i] = toShaderSummaryResponse(shader)
	}
	return server.WriteJSON(w, http.StatusOK, response)
}
