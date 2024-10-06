package main

import (
	"fmt"
	"github.com/goioc/di"
	"github.com/sdedovic/wgsltoy-server/src/go/db"
	shaderService "github.com/sdedovic/wgsltoy-server/src/go/service/shader"
	userService "github.com/sdedovic/wgsltoy-server/src/go/service/user"
	"github.com/sdedovic/wgsltoy-server/src/go/web"
	"github.com/sdedovic/wgsltoy-server/src/go/web/shader"
	"github.com/sdedovic/wgsltoy-server/src/go/web/user"
	"log"
	"net/http"
	"reflect"
)

func run() error {
	// set up Postgres connection
	pgClient, err := db.InitializePgClient()
	if err != nil {
		return fmt.Errorf("unable to connect to database caused by: %w", err)
	}
	defer db.CloseStorageDb(pgClient)

	// Initialize IOC container
	_, err = di.RegisterBeanInstance("PgClient", &pgClient)
	if err != nil {
		return fmt.Errorf("unable to register PgClient: %w", err)
	}
	_, err = di.RegisterBean("Repository", reflect.TypeOf((*db.Repository)(nil)))
	if err != nil {
		return fmt.Errorf("unable to register Repository: %w", err)
	}
	_, err = di.RegisterBean("UserService", reflect.TypeOf((*userService.Service)(nil)))
	if err != nil {
		return fmt.Errorf("unable to register UserService: %w", err)
	}
	_, err = di.RegisterBean("ShaderService", reflect.TypeOf((*shaderService.Service)(nil)))
	if err != nil {
		return fmt.Errorf("unable to register ShaderService: %w", err)
	}
	_, err = di.RegisterBean("UserController", reflect.TypeOf((*user.Controller)(nil)))
	if err != nil {
		return fmt.Errorf("unable to register UserController: %w", err)
	}
	_, err = di.RegisterBean("ShaderController", reflect.TypeOf((*shader.Controller)(nil)))
	if err != nil {
		return fmt.Errorf("unable to register ShaderController: %w", err)
	}
	if err = di.InitializeContainer(); err != nil {
		return fmt.Errorf("unable to connect to initialize application caused by: %w", err)
	}

	// register route handlers
	http.HandleFunc("/health", web.HealthCheck())

	userController := di.GetInstance("UserController").(*user.Controller)
	http.HandleFunc("/user/register", userController.UserRegister())
	http.HandleFunc("/user/login", userController.UserLogin())
	http.HandleFunc("/user/me", userController.UserMe())

	shaderController := di.GetInstance("ShaderController").(*shader.Controller)
	http.HandleFunc("/shader", shaderController.ShaderCreate())
	http.HandleFunc("/user/me/shader/", shaderController.ShaderInfoListOwn())
	http.HandleFunc("/shader/{id}", shaderController.ShaderById())

	// start server
	log.Println("INFO", "Starting server on 0.0.0.0:8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	log.SetFlags(log.Lshortfile)

	err := run()
	if err != nil {
		log.Println("FATAL", err)
	}
}
