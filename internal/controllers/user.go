package controllers

import (
	"errors"
	"net/http"
	"time"

	"github.com/sdedovic/wgsltoy-server/internal/domain"
	"github.com/sdedovic/wgsltoy-server/internal/server"
	"go.uber.org/zap"
)

type UserController struct {
	users *domain.UserService
	log   *zap.Logger
}

func NewUserController(users *domain.UserService, log *zap.Logger) *UserController {
	return &UserController{users, log}
}

func (c *UserController) Register(mux *http.ServeMux) {
	mux.Handle("POST /api/v1/user/register", server.Handle(c.registerUser))
	mux.Handle("POST /api/v1/user/login", server.Handle(c.login))
	mux.Handle("GET /api/v1/user/me", server.Handle(c.me))
}

type registerRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type userResponse struct {
	ID                      string    `json:"id"`
	CreatedAt               time.Time `json:"createdAt"`
	UpdatedAt               time.Time `json:"updatedAt"`
	Username                string    `json:"username"`
	Email                   string    `json:"email"`
	EmailVerificationStatus string    `json:"emailVerificationStatus"`
}

//	@Summary	Register a new user
//	@Tags		user
//	@Accept		json
//	@Produce	json
//	@Param		body	body	registerRequest	true	"New user"
//	@Success	201
//	@Failure	400	{object}	server.ErrorResponse
//	@Router		/api/v1/user/register [post]
func (c *UserController) registerUser(w http.ResponseWriter, r *http.Request) error {
	var req registerRequest
	if err := server.DecodeJSON(r, &req); err != nil {
		return err
	}

	err := c.users.Register(r.Context(), req.Username, req.Email, req.Password)
	if errors.Is(err, domain.ErrUsernameBanned) {
		c.log.Warn("banned username attempted", zap.String("username", req.Username))
	}
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusCreated)
	return nil
}

//	@Summary		Log in
//	@Description	Returns a JWT to send as a bearer token.
//	@Tags			user
//	@Accept			json
//	@Produce		plain
//	@Param			body	body		loginRequest	true	"Credentials"
//	@Success		200		{string}	string			"JWT"
//	@Failure		400		{object}	server.ErrorResponse
//	@Router			/api/v1/user/login [post]
func (c *UserController) login(w http.ResponseWriter, r *http.Request) error {
	var req loginRequest
	if err := server.DecodeJSON(r, &req); err != nil {
		return err
	}

	token, err := c.users.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		return err
	}

	// TODO: return JSON like every other route
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, err = w.Write([]byte(token))
	return err
}

//	@Summary	Get the current user
//	@Tags		user
//	@Produce	json
//	@Security	BearerAuth
//	@Success	200	{object}	userResponse
//	@Failure	401	{object}	server.ErrorResponse
//	@Router		/api/v1/user/me [get]
func (c *UserController) me(w http.ResponseWriter, r *http.Request) error {
	user, err := c.users.GetCurrent(r.Context())
	if err != nil {
		return err
	}

	return server.WriteJSON(w, http.StatusOK, userResponse{
		ID:                      string(user.ID),
		CreatedAt:               user.CreatedAt,
		UpdatedAt:               user.UpdatedAt,
		Username:                string(user.Username),
		Email:                   string(user.Email),
		EmailVerificationStatus: string(user.EmailVerification),
	})
}
