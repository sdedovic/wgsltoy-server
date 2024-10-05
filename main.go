package main

import (
	"github.com/sdedovic/wgsltoy-server/src/go/db"
	"github.com/sdedovic/wgsltoy-server/src/go/web"
	"log"
	"net/http"
	"os"
)

//==== Main ====\\

func main() {
	storage, err := db.InitializeDataDbConnection()
	if err != nil {
		log.Println("ERROR", "Unable to connect to database caused by:", err.Error())
		os.Exit(1)
	}
	defer db.CloseStorageDb(storage)

	http.HandleFunc("/health", web.HealthCheck())

	http.HandleFunc("/user/register", web.UserRegister())
	http.HandleFunc("/user/login", web.UserLogin())
	http.HandleFunc("/user/me", web.UserMe())

	http.HandleFunc("/shader", web.ShaderCreate())
	http.HandleFunc("/user/me/shader/", web.ShaderInfoListOwn())
	http.HandleFunc("/shader/{id}", web.ShaderById())

	log.Println("INFO", "Starting server on 0.0.0.0:8080")

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Println("ERROR", err.Error())
		os.Exit(1)
	}
}
