package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type RegisterUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Captcha  string `json:"Capture"`
}

type HealthResponse struct {
	Status string `json:"status"`
}

func main() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(HealthResponse{"ok"})
	})

	fmt.Println("Starting server on 0.0.0.0:8080")

	http.ListenAndServe(":8080", nil)

}
