package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/Xeninon/Rezonant/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return
	}
	dbQueries := database.New(db)
	cfg := apiConfig{db: dbQueries}
	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app", cfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", handlerHealthz)
	mux.HandleFunc("GET /admin/metrics", cfg.handlerMetrics)
	mux.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)
	mux.HandleFunc("POST /admin/reset", cfg.handlerReset)
	srv := &http.Server{
		Handler: mux,
		Addr:    ":8080",
	}
	log.Fatal(srv.ListenAndServe())
}

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
}
