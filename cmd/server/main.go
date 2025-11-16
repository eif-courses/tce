package main

import (
	"net/http"
	"os"

	"github.com/eif-courses/tce/internal/api"
	"github.com/eif-courses/tce/internal/util"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	// Load .env (optional)
	_ = godotenv.Load()

	// === Init Structured Logger ===
	util.InitLogger()
	defer util.SyncLogger()

	util.Log.Info("Starting ThesisCheck Engine (TCE)...")

	// === Router ===
	r := chi.NewRouter()

	// === CORS ===
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:3000",
			"http://127.0.0.1:3000",
		},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// === Logging middleware ===
	r.Use(api.LoggerMiddleware)

	// === Routes ===
	r.Get("/api/tce/health", api.HandleHealth)
	r.Post("/api/tce/check", api.HandleCheck)

	// === Server port ===
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	util.Log.Info("TCE API server listening",
		zap.String("port", port),
	)

	if err := http.ListenAndServe(":"+port, r); err != nil {
		util.Log.Fatal("server crashed", zap.Error(err))
	}
}
