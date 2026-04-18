package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", handlerHealthz)
	mux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	srv := &http.Server{
		Handler: mux,
		Addr:    ":8080",
	}
	log.Fatal(srv.ListenAndServe())
}
