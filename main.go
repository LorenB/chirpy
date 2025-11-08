package main

import (
	"fmt"
	"log"
	"net/http"
)

type RootHanlder struct{}

// Implements the http.Handler interface directly in the struct.
// This method's signature exactly matches the interface requirement.
func (h *RootHanlder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// r.URL.Path now has "/app" stripped
	filepath := r.URL.Path

	// If path is empty or just "/", serve index.html
	if filepath == "" || filepath == "/" {
		filepath = "/index.html"
	}

	// Remove leading slash to make it relative to current directory
	filepath = filepath[1:] // Remove the leading "/"

	fmt.Printf("Serving file: %v\n", filepath)
	http.ServeFile(w, r, filepath)
}

type HealthHanlder struct{}

func (h *HealthHanlder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(`OK`))
}

func main() {
	const port = "8080"
	mux := http.NewServeMux()

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	// fileServer := http.FileServer(http.Dir("."))
	// mux.Handle("/app/", http.StripPrefix("/app", fileServer))
	mux.Handle("/app/", http.StripPrefix("/app", &RootHanlder{}))
	mux.Handle("/healthz", &HealthHanlder{})
	log.Println("Starting server on :8080")
	log.Fatal(srv.ListenAndServe())
}
