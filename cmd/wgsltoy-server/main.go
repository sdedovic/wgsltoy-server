// @title						WGSL Toy API
// @version					1.0
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
// @description				"Bearer <token>", with the token from /api/v1/user/login
package main

import (
	"github.com/sdedovic/wgsltoy-server/internal/app"
	"go.uber.org/fx"
)

func main() {
	fx.New(app.Module).Run()
}
