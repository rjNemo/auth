package main

import (
	"log"
	"net/http"

	"github.com/rjnemo/auth/internal/server"
)

func main() {
	srv := server.New()

	log.Println("Starting server on http://localhost:8000")
	if err := http.ListenAndServe(":8000", srv.Router()); err != nil {
		log.Fatalf("listen: %v", err)
	}
}
