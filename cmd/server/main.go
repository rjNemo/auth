package main

import (
	"log"
	"net/http"

	"github.com/rjnemo/auth/internal/server"
)

func main() {
	srv, err := server.New()
	if err != nil {
		log.Fatalf("initialise server: %v", err)
	}

	log.Println("Starting server on http://localhost:8000")
	if err := http.ListenAndServe(":8000", srv.Router()); err != nil {
		log.Fatalf("listen: %v", err)
	}
}
