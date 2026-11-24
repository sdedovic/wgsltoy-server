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

func (c *UserController) Register(router *server.Router) {
	router.Handle("GET", meRoute, server.Handle(c.me))
	router.Handle("PUT", meRoute, server.Handle(c.update))
	router.Handle("DELETE", meRoute, server.Handle(c.delete))
	router.Handle("GET", userRoute, server.Handle(c.profile))
}

type mePageData struct {
	Principal *domain.Principal
	Error     string
}

type profilePageData struct {
	Principal *domain.Principal
	Profile   domain.Profile
}

// @Summary	Get the signed in user
// @Tags		users
// @Produce	json
// @Success	200	{object}	meDTO
// @Failure	401	{object}	server.ErrorResponse
// @Security	SessionCookie
// @Router		/me [get]
func (c *UserController) me(w http.ResponseWriter, r *http.Request) error {
	user, err := c.users.GetCurrent(r.Context())
	if err != nil {
		return err
	}
	return server.Respond(w, r, http.StatusOK, c.tmpl.Me, mePageData{Principal: principalPtr(r)}, newMeDTO(user))
}

// @Summary		Replace the signed in user's account details
// @Description	Changing the email marks it as unverified.
// @Tags			users
// @Accept			json,x-www-form-urlencoded,mpfd
// @Produce		json
// @Param			user	body		updateMeDTO	true	"The account's new details"
// @Success		200		{object}	meDTO
// @Failure		400		{object}	server.ErrorResponse
// @Failure		401		{object}	server.ErrorResponse
// @Failure		409		{object}	server.ErrorResponse	"The username or email is taken"
// @Failure		415		{object}	server.ErrorResponse
// @Failure		422		{object}	server.ErrorResponse
// @Security		SessionCookie
// @Router			/me [put]
func (c *UserController) update(w http.ResponseWriter, r *http.Request) error {
	var req updateMeDTO
	if err := server.Decode(r, &req); err != nil {
		return err
	}
	user, err := c.users.Update(r.Context(), domain.UpdateUserCommand{
		Username:   req.Username,
		Email:      req.Email,
		Visibility: req.Visibility,
	})
	if err != nil {
		return server.WriteFormError(w, r, err, c.tmpl.Me, func(message string) any {
			return mePageData{Principal: principalPtr(r), Error: message}
		})
	}

	if server.WantsJSON(r) {
		return server.WriteJSON(w, http.StatusOK, newMeDTO(user))
	}
	http.Redirect(w, r, meRoute.Path(), http.StatusSeeOther)
	return nil
}

// @Summary		Delete the signed in user's account
// @Description	Signs out too. The user's shaders are handled by shaderDisposition: default deletes the
// @Description	private ones and keeps the rest without an author, remove_all deletes them all, and
// @Description	donate_all keeps them all without an author.
// @Tags			users
// @Produce		json
// @Param			shaderDisposition	query	string	false	"What happens to the user's shaders"	Enums(default, remove_all, donate_all)	default(default)
// @Success		204
// @Failure		401	{object}	server.ErrorResponse
// @Failure		422	{object}	server.ErrorResponse
// @Security		SessionCookie
// @Router			/me [delete]
func (c *UserController) delete(w http.ResponseWriter, r *http.Request) error {
	disposition, err := domain.NewShaderDisposition(r.URL.Query().Get("shaderDisposition"))
	if err != nil {
		return err
	}
	if err := c.users.Delete(r.Context(), disposition); err != nil {
		return err
	}
	server.ClearSessionCookie(w)

	if server.WantsJSON(r) {
		w.WriteHeader(http.StatusNoContent)
		return nil
	}
	http.Redirect(w, r, homeRoute.Path(), http.StatusSeeOther)
	return nil
}

// @Summary		Get a user's profile
// @Description	A private account's profile has only its username, unless it's the signed in user's or the
// @Description	viewer is an admin.
// @Tags			users
// @Produce		json
// @Param			username	path		string	true	"Username"
// @Success		200			{object}	profileDTO
// @Failure		404			{object}	server.ErrorResponse
// @Router			/users/{username} [get]
func (c *UserController) profile(w http.ResponseWriter, r *http.Request) error {
	profile, err := c.users.GetProfile(r.Context(), r.PathValue("username"))
	if err != nil {
		return err
	}
	return server.Respond(w, r, http.StatusOK, c.tmpl.Profile,
		profilePageData{Principal: principalPtr(r), Profile: profile}, newProfileDTO(profile))
}
