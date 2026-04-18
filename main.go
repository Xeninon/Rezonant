package main

import (
	"log"
	"net/http"
	"sync/atomic"
)

func main() {
	cfg := apiConfig{}
	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app", cfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))))
	mux.HandleFunc("/healthz", handlerHealthz)
	mux.HandleFunc("/metrics", cfg.handlerMetrics)
	mux.HandleFunc("/reset", cfg.handlerReset)
	srv := &http.Server{
		Handler: mux,
		Addr:    ":8080",
	}
	log.Fatal(srv.ListenAndServe())
}

type apiConfig struct {
	fileserverHits atomic.Int32
}
