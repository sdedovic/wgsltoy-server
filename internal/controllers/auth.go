package controllers

import (
	"errors"
	"net/http"

	"github.com/sdedovic/wgsltoy-server/internal/domain"
	"github.com/sdedovic/wgsltoy-server/internal/server"
	"github.com/sdedovic/wgsltoy-server/internal/views"
	"go.uber.org/zap"
)

type AuthController struct {
	users *domain.UserService
	tmpl  *views.Templates
	log   *zap.Logger
}

func NewAuthController(users *domain.UserService, tmpl *views.Templates, log *zap.Logger) *AuthController {
	return &AuthController{users, tmpl, log}
}

func (c *AuthController) Register(mux *http.ServeMux) {
	mux.Handle("GET /login", server.HandleHTML(c.loginPage))
	mux.Handle("POST /login", server.HandleHTML(c.login))
	mux.Handle("POST /logout", server.HandleHTML(c.logout))
	mux.Handle("GET /register", server.HandleHTML(c.registerPage))
	mux.Handle("POST /register", server.HandleHTML(c.register))
}

type loginPageData struct {
	Principal *domain.Principal
	Error     string
}

type registerPageData struct {
	Principal *domain.Principal
	Error     string
}

func (c *AuthController) loginPage(w http.ResponseWriter, r *http.Request) error {
	return server.WriteHTML(w, http.StatusOK, c.tmpl.Login, loginPageData{})
}

func (c *AuthController) login(w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return err
	}

	session, err := c.users.Login(r.Context(), r.FormValue("username"), r.FormValue("password"))
	if err != nil {
		return server.WriteHTML(w, http.StatusUnauthorized, c.tmpl.Login, loginPageData{Error: err.Error()})
	}

	server.SetSessionCookie(w, string(session.ID), session.ExpiresAt)
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}

func (c *AuthController) logout(w http.ResponseWriter, r *http.Request) error {
	if cookie, err := r.Cookie("session"); err == nil {
		_ = c.users.Logout(r.Context(), domain.SessionID(cookie.Value))
	}
	server.ClearSessionCookie(w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
	return nil
}

func (c *AuthController) registerPage(w http.ResponseWriter, r *http.Request) error {
	return server.WriteHTML(w, http.StatusOK, c.tmpl.Register, registerPageData{})
}

func (c *AuthController) register(w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return err
	}

	username := r.FormValue("username")
	err := c.users.Register(r.Context(), username, r.FormValue("email"), r.FormValue("password"))
	if err != nil {
		if errors.Is(err, domain.ErrUsernameBanned) {
			c.log.Warn("banned username attempted", zap.String("username", username))
		}
		return server.WriteHTML(w, http.StatusBadRequest, c.tmpl.Register, registerPageData{Error: err.Error()})
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
	return nil
}
