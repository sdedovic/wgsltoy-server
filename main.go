package main

import (
	"github.com/goioc/di"
	"github.com/sdedovic/wgsltoy-server/src/go/db"
	shaderService "github.com/sdedovic/wgsltoy-server/src/go/service/shader"
	userService "github.com/sdedovic/wgsltoy-server/src/go/service/user"
	"github.com/sdedovic/wgsltoy-server/src/go/web"
	"github.com/sdedovic/wgsltoy-server/src/go/web/shader"
	"github.com/sdedovic/wgsltoy-server/src/go/web/user"
	"log"
	"net/http"
	"os"
	"reflect"
)

//==== Main ====\\

func main() {
	pgClient, err := db.InitializePgClient()
	if err != nil {
		log.Println("ERROR", "Unable to connect to database caused by:", err.Error())
		os.Exit(1)
	}
	defer db.CloseStorageDb(pgClient)

	_, _ = di.RegisterBeanInstance("Repository", pgClient)
	_, _ = di.RegisterBean("UserService", reflect.TypeOf((*userService.Service)(nil)))
	_, _ = di.RegisterBean("ShaderService", reflect.TypeOf((*shaderService.Service)(nil)))
	_, _ = di.RegisterBean("UserController", reflect.TypeOf((*user.Controller)(nil)))
	_, _ = di.RegisterBean("ShaderController", reflect.TypeOf((*shader.Controller)(nil)))
	if err = di.InitializeContainer(); err != nil {
		log.Println("ERROR", "Unable to connect to initialize application caused by:", err.Error())
		os.Exit(1)
	}

	http.HandleFunc("/health", web.HealthCheck())

	userController := di.GetInstance("UserController").(*user.Controller)
	http.HandleFunc("/user/register", userController.UserRegister())
	http.HandleFunc("/user/login", userController.UserLogin())
	http.HandleFunc("/user/me", userController.UserMe())

	shaderController := di.GetInstance("ShaderController").(*shader.Controller)
	http.HandleFunc("/shader", shaderController.ShaderCreate())
	http.HandleFunc("/user/me/shader/", shaderController.ShaderInfoListOwn())
	http.HandleFunc("/shader/{id}", shaderController.ShaderById())

	log.Println("INFO", "Starting server on 0.0.0.0:8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Println("ERROR", err.Error())
		os.Exit(1)
	}
}
