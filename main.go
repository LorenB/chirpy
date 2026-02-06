package main

import (
	"chirpy/internal/databse"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
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

type CreateUserHandler struct {
	queries *databse.Queries
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (h *CreateUserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
	}
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	// userHanlder, err := h.queries.CreateUser(r.Context(), params.Email)

	dbUser, err := h.queries.CreateUser(r.Context(), params.Email)
	user := User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	}
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(user); err != nil {
		log.Panicf("Error encoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

}

type MetricsHanlder struct {
	hits *atomic.Int32
}

func (h *MetricsHanlder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)
	hits := h.hits.Load()
	tmpl := `
	<html>
	<body>
		<h1>Welcome, Chirpy Admin</h1>
		<p>Chirpy has been visited %d times!</p>
	</body>
	</html>
	`
	t := fmt.Sprintf(tmpl, hits)

	fmt.Fprintln(w, t)
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
	queries        *databse.Queries
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler, queries *databse.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		fmt.Printf("Hits: %d\n", cfg.fileserverHits.Load())
		cfg.queries = queries
		next.ServeHTTP(w, r)
	})
}

func main() {
	const port = "8080"

	godotenv.Load()
	dbURl := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURl)
	if err != nil {
		log.Fatal("error connecting to database")
	}
	dbQueries := databse.New(db)

	mux := http.NewServeMux()
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	apiCfg := &apiConfig{}
	wrappedHandler := apiCfg.middlewareMetricsInc(&RootHanlder{}, dbQueries)
	mux.Handle("/app/", http.StripPrefix("/app", wrappedHandler))

	mux.Handle("GET /api/healthz", &HealthHanlder{})
	// mux.Handle("POST /api/users", &CreateUserHandler{})

	mux.HandleFunc("POST /api/validate_chirp", handlerChirpsValidate)

	metricsHndl := &MetricsHanlder{}
	metricsHndl.hits = &apiCfg.fileserverHits
	mux.Handle("GET /admin/metrics", metricsHndl)

	resetHndl := &ResetHanlder{}
	resetHndl.hits = &apiCfg.fileserverHits
	mux.Handle("POST /admin/reset", resetHndl)

	log.Println("Starting server on :8080")
	log.Fatal(srv.ListenAndServe())
}
