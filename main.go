package main

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sdedovic/wgsltoy-server/src/go/web"
	"log"
	"net/http"
	"os"
)

//==== Main ====\\

func main() {
	pgPool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println("ERROR", "Unable to connect to database!", err.Error())
		os.Exit(1)
	}
	defer pgPool.Close()

	http.HandleFunc("/health", web.HealthCheck())
	http.HandleFunc("/user/register", web.UserRegister(pgPool))

	log.Println("INFO", "Starting server on 0.0.0.0:8080")

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Println("ERROR", err.Error())
		os.Exit(1)
	}
}
