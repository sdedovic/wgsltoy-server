package controllers

import (
	"html/template"

	"github.com/sdedovic/wgsltoy-server/internal/domain"
	"github.com/sdedovic/wgsltoy-server/internal/server"
)

// Every route the app serves. Handlers and templates build paths from these, so changing one is enough.
var (
	homeRoute     = server.Route("/")
	healthRoute   = server.Route("/health")
	loginRoute    = server.Route("/login")
	logoutRoute   = server.Route("/logout")
	registerRoute = server.Route("/register")
	meRoute       = server.Route("/me")
	userRoute     = server.Route("/users/{username}")
	shadersRoute  = server.Route("/shaders")
	shaderRoute   = server.Route("/shaders/{id}")

	// shaders moved under /shaders and the editor onto the shader page, these keep old links working
	legacyShaderRoute       = server.Route("/shader/{id}")
	legacyShaderEditorRoute = server.Route("/shader/{id}/edit")
)

func shaderPath(id domain.ShaderID) string {
	return shaderRoute.Path(string(id))
}

func userPath(username domain.Username) string {
	return userRoute.Path(string(username))
}

// URLFuncs build paths in templates, e.g. {{shaderURL .ID}}.
func URLFuncs() template.FuncMap {
	return template.FuncMap{
		"homeURL":     homeRoute.Path,
		"loginURL":    loginRoute.Path,
		"logoutURL":   logoutRoute.Path,
		"registerURL": registerRoute.Path,
		"meURL":       meRoute.Path,
		"userURL":     userPath,
		"shadersURL":  shadersRoute.Path,
		"shaderURL":   shaderPath,
	}
}
