// @title			WGSL Toy
// @version		1.0
// @description	Routes answer with JSON when the request's Accept prefers application/json, and with HTML
// @description	otherwise. Only the JSON is described here. Request bodies can be JSON or form fields.
//
// @securityDefinitions.apikey	SessionCookie
// @in							cookie
// @name						session
// @description				Set by POST /login.
package main

import (
	"go.uber.org/fx"

	"github.com/sdedovic/wgsltoy-server/internal/app"
)

func main() {
	fx.New(app.Module).Run()
}
