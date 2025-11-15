package main

import (
	"log"
	"net/http"
	"os"

	"github.com/eif-courses/tce/internal/api"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

// Nauja programine iranga : ThesisCheck Engine (TCE)
func main() {
	r := chi.NewRouter()

	// Allow Nuxt dev server to talk to this API
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://127.0.0.1:3000"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// health check
	r.Get("/api/tce/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	// main check endpoint
	r.Post("/api/tce/check", api.CheckHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("TCE server listening on :%s\n", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err)
	}
}
