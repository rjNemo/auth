package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/rjnemo/auth/web"
)

var (
	loggedIn  = false
	templates = template.Must(template.ParseFS(web.Templates, "templates/index.html", "templates/in.html", "templates/unauthorized.html"))
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", handleIndex)
	mux.HandleFunc("GET /in", handleIn)
	mux.HandleFunc("POST /login", handleLogin)

	log.Println("Starting server on http://localhost:8000")
	if err := http.ListenAndServe(":8000", mux); err != nil {
		log.Fatalf("listen: %v", err)
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		http.Error(w, "template render failed", http.StatusInternalServerError)
	}
}

func handleIn(w http.ResponseWriter, r *http.Request) {
	if !loggedIn {
		w.WriteHeader(http.StatusUnauthorized)
		err := templates.ExecuteTemplate(w, "unauthorized.html", nil)
		if err != nil {
			http.Error(w, "template render failed", http.StatusInternalServerError)
		}
		return
	}

	err := templates.ExecuteTemplate(w, "in.html", nil)
	if err != nil {
		http.Error(w, "template render failed", http.StatusInternalServerError)
	}
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	log.Println("Login request received")
	loggedIn = true
	http.Redirect(w, r, "/in", http.StatusSeeOther)
}
