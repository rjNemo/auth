package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/rjnemo/auth/internal/server"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("run: %v", err)
	}
}

func run() error {
	srv, err := server.New()
	if err != nil {
		return fmt.Errorf("initialise server: %v", err)
	}

	log.Println("Starting server on http://localhost:8000")
	if err := http.ListenAndServe(":8000", srv.Router()); err != nil {
		return fmt.Errorf("listen: %v", err)
	}

	return nil
}
