package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
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

	fmt.Printf("Handling %v Serving file: %v\n", r.URL.Path, filepath)
	http.ServeFile(w, r, filepath)
}

type HealthHanlder struct{}

func (h *HealthHanlder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(`OK`))
}

type MetricsHanlder struct {
	hits *atomic.Int32
}

func (h *MetricsHanlder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	fmt.Fprintf(w, "Hits: %d", h.hits.Load())
}

type ResetHanlder struct {
	hits *atomic.Int32
}

func (h *ResetHanlder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	h.hits.Store(0)
	w.WriteHeader(200)
	w.Write([]byte(`OK`))
}

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		fmt.Printf("Hits: %d\n", cfg.fileserverHits.Load())
		next.ServeHTTP(w, r)
	})
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

	apiCfg := &apiConfig{}
	wrappedHandler := apiCfg.middlewareMetricsInc(&RootHanlder{})
	mux.Handle("/app/", http.StripPrefix("/app", wrappedHandler))

	// mux.Handle("/app/", http.StripPrefix("/app", &RootHanlder{}))

	mux.Handle("GET /api/healthz", &HealthHanlder{})

	metricsHndl := &MetricsHanlder{}
	metricsHndl.hits = &apiCfg.fileserverHits
	mux.Handle("GET /api/metrics", metricsHndl)

	resetHndl := &ResetHanlder{}
	resetHndl.hits = &apiCfg.fileserverHits
	mux.Handle("POST /api/reset", resetHndl)

	log.Println("Starting server on :8080")
	log.Fatal(srv.ListenAndServe())
}
