package controllers

import (
	"net/http"

	"github.com/sdedovic/wgsltoy-server/internal/domain"
	"github.com/sdedovic/wgsltoy-server/internal/server"
	"github.com/sdedovic/wgsltoy-server/internal/views"
)

type UserController struct {
	users *domain.UserService
	tmpl  *views.Templates
}

func NewUserController(users *domain.UserService, tmpl *views.Templates) *UserController {
	return &UserController{users, tmpl}
}

func (c *UserController) Register(mux *http.ServeMux) {
	mux.Handle("GET /me", server.HandleHTML(c.me))
	mux.Handle("PUT /me", server.HandleHTML(c.update))
	mux.Handle("DELETE /me", server.HandleHTML(c.delete))
	mux.Handle("GET /{username}", server.HandleHTML(c.profile))
}

type mePageData struct {
	Principal *domain.Principal
	Error     string
}

type profilePageData struct {
	Principal *domain.Principal
	User      domain.User
}

func (c *UserController) me(w http.ResponseWriter, r *http.Request) error {
	if _, err := c.users.GetCurrent(r.Context()); err != nil {
		return err
	}
	p, _ := domain.PrincipalFromContext(r.Context())
	return server.WriteHTML(w, http.StatusOK, c.tmpl.Me, mePageData{Principal: &p})
}

func (c *UserController) update(w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return err
	}
	_, err := c.users.Update(r.Context(), domain.UpdateUserCommand{
		Username:   r.FormValue("username"),
		Email:      r.FormValue("email"),
		Visibility: r.FormValue("visibility"),
	})
	if err != nil {
		p, _ := domain.PrincipalFromContext(r.Context())
		return server.WriteHTML(w, http.StatusUnprocessableEntity, c.tmpl.Me, mePageData{Principal: &p, Error: err.Error()})
	}
	http.Redirect(w, r, "/me", http.StatusSeeOther)
	return nil
}

func (c *UserController) delete(w http.ResponseWriter, r *http.Request) error {
	disposition, err := domain.NewShaderDisposition(r.URL.Query().Get("shaderDisposition"))
	if err != nil {
		return err
	}
	if err := c.users.Delete(r.Context(), disposition); err != nil {
		return err
	}
	server.ClearSessionCookie(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}

func (c *UserController) profile(w http.ResponseWriter, r *http.Request) error {
	user, err := c.users.GetByUsername(r.Context(), r.PathValue("username"))
	if err != nil {
		return err
	}
	var principal *domain.Principal
	if p, ok := domain.PrincipalFromContext(r.Context()); ok {
		principal = &p
	}
	return server.WriteHTML(w, http.StatusOK, c.tmpl.Profile, profilePageData{Principal: principal, User: user})
}
