package main

import (
	"log"
	"net/http"
)

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		respondWithError(w, 403, "incorrect platform")
		return
	}

	cfg.fileserverHits.Store(0)
	if err := cfg.db.DeleteUsers(r.Context()); err != nil {
		log.Printf("Error deleting users: %s", err)
		respondWithError(w, 500, "error deleting users")
		return
	}

	r.Header.Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}
