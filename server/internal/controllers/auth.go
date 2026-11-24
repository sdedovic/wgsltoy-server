package controllers

import (
	"errors"
	"net/http"

	"go.uber.org/zap"

	"github.com/sdedovic/wgsltoy-server/internal/domain"
	"github.com/sdedovic/wgsltoy-server/internal/server"
	"github.com/sdedovic/wgsltoy-server/internal/views"
)

type AuthController struct {
	users *domain.UserService
	tmpl  *views.Templates
	log   *zap.Logger
}

func NewAuthController(users *domain.UserService, tmpl *views.Templates, log *zap.Logger) *AuthController {
	return &AuthController{users, tmpl, log}
}

func (c *AuthController) Register(router *server.Router) {
	router.Handle("GET", loginRoute, server.Handle(c.loginPage))
	router.Handle("POST", loginRoute, server.Handle(c.login))
	router.Handle("POST", logoutRoute, server.Handle(c.logout))
	router.Handle("GET", registerRoute, server.Handle(c.registerPage))
	router.Handle("POST", registerRoute, server.Handle(c.register))
}

type loginPageData struct {
	Principal *domain.Principal
	Error     string
}

type registerPageData struct {
	Principal *domain.Principal
	Error     string
}

// The login and register pages are only forms, so have no JSON to offer.
//
// @Summary		Login page
// @Description	Only HTML.
// @Tags			auth
// @Produce		html,json
// @Success		200	{string}	string	"The login page"
// @Failure		406	{object}	server.ErrorResponse
// @Router			/login [get]
func (c *AuthController) loginPage(w http.ResponseWriter, r *http.Request) error {
	if server.WantsJSON(r) {
		return server.ErrNotAcceptable
	}
	return server.WriteHTML(w, http.StatusOK, c.tmpl.Login, loginPageData{})
}

// @Summary		Sign in
// @Description	Sets the session cookie.
// @Tags			auth
// @Accept			json,x-www-form-urlencoded,mpfd
// @Produce		json
// @Param			credentials	body	loginDTO	true	"Credentials"
// @Success		204
// @Header			204	{string}	Set-Cookie	"The session cookie"
// @Failure		400	{object}	server.ErrorResponse
// @Failure		401	{object}	server.ErrorResponse	"The username or password is wrong"
// @Failure		415	{object}	server.ErrorResponse
// @Failure		422	{object}	server.ErrorResponse
// @Router			/login [post]
func (c *AuthController) login(w http.ResponseWriter, r *http.Request) error {
	var req loginDTO
	if err := server.Decode(r, &req); err != nil {
		return err
	}

	session, err := c.users.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		return server.WriteFormError(w, r, err, c.tmpl.Login, func(message string) any {
			return loginPageData{Error: message}
		})
	}

	server.SetSessionCookie(w, string(session.ID), session.ExpiresAt)
	if server.WantsJSON(r) {
		w.WriteHeader(http.StatusNoContent)
		return nil
	}
	http.Redirect(w, r, homeRoute.Path(), http.StatusSeeOther)
	return nil
}

// @Summary		Sign out
// @Description	Ends the session and clears its cookie. Signing out without a session is fine too.
// @Tags			auth
// @Produce		json
// @Success		204
// @Router			/logout [post]
func (c *AuthController) logout(w http.ResponseWriter, r *http.Request) error {
	if cookie, err := r.Cookie("session"); err == nil {
		_ = c.users.Logout(r.Context(), domain.SessionID(cookie.Value))
	}
	server.ClearSessionCookie(w)
	if server.WantsJSON(r) {
		w.WriteHeader(http.StatusNoContent)
		return nil
	}
	http.Redirect(w, r, loginRoute.Path(), http.StatusSeeOther)
	return nil
}

// @Summary		Register page
// @Description	Only HTML.
// @Tags			auth
// @Produce		html,json
// @Success		200	{string}	string	"The register page"
// @Failure		406	{object}	server.ErrorResponse
// @Router			/register [get]
func (c *AuthController) registerPage(w http.ResponseWriter, r *http.Request) error {
	if server.WantsJSON(r) {
		return server.ErrNotAcceptable
	}
	return server.WriteHTML(w, http.StatusOK, c.tmpl.Register, registerPageData{})
}

// @Summary		Create an account
// @Description	The account starts with an unverified email, and isn't signed in. Needs the invite code, when one is configured.
// @Tags			auth
// @Accept			json,x-www-form-urlencoded,mpfd
// @Produce		json
// @Param			account	body	registerDTO	true	"The new account"
// @Success		201
// @Header			201	{string}	Location	"The new user's profile URL"
// @Failure		400	{object}	server.ErrorResponse
// @Failure		409	{object}	server.ErrorResponse	"The username or email is taken"
// @Failure		415	{object}	server.ErrorResponse
// @Failure		422	{object}	server.ErrorResponse
// @Router			/register [post]
func (c *AuthController) register(w http.ResponseWriter, r *http.Request) error {
	var req registerDTO
	if err := server.Decode(r, &req); err != nil {
		return err
	}

	err := c.users.Register(r.Context(), req.Username, req.Email, req.Password, req.InviteCode)
	if err != nil {
		if errors.Is(err, domain.ErrUsernameBanned) {
			c.log.Warn("banned username attempted", zap.String("username", req.Username))
		}
		return server.WriteFormError(w, r, err, c.tmpl.Register, func(message string) any {
			return registerPageData{Error: message}
		})
	}

	if server.WantsJSON(r) {
		w.Header().Set("Location", userPath(domain.Username(req.Username)))
		w.WriteHeader(http.StatusCreated)
		return nil
	}
	http.Redirect(w, r, loginRoute.Path(), http.StatusSeeOther)
	return nil
}
