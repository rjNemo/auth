package main

import (
	"log"
	"net/http"
)

var loggedIn = false

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	mux.HandleFunc("GET /in", func(w http.ResponseWriter, r *http.Request) {
		if loggedIn {
			http.ServeFile(w, r, "in.html")
		} else {
			w.Header().Add("Content-Type", "text/html")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`
				<head>
					<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@picocss/pico@2/css/pico.min.css" />
				</head>
				<body>
					<main class="container">
						<h1>Unauthorized</h1> <a href='/' role='button'> Back to safety </a>
					</main>
				</body>
				`))
		}
	})

	mux.HandleFunc("POST /login", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Login request received")
		loggedIn = true
		http.Redirect(w, r, "/in", http.StatusSeeOther)
	})

	log.Println("Starting server on http://localhost:8000")
	http.ListenAndServe(":8000", mux)
}
