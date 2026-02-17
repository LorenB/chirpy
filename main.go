package main

import (
	"chirpy/internal/auth"
	"chirpy/internal/database"
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

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	dbChirps, err := cfg.db.GetChirps(r.Context())
	if err != nil {
		log.Printf("error writing to database: %s", err)
		w.WriteHeader(500)
		return
	}
	var chirps []Chirp
	for _, dbChirp := range dbChirps {
		chirps = append(
			chirps,
			Chirp{
				UserID:    dbChirp.UserID,
				Body:      dbChirp.Body,
				CreatedAt: dbChirp.CreatedAt,
				UpdatedAt: dbChirp.CreatedAt,
			},
		)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(&chirps); err != nil {
		log.Panicf("Error encoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}
}

func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
	chirpIDStr := r.PathValue("chirpID")
	// conver to UUID
	chirpID, err := uuid.Parse(chirpIDStr)
	if err != nil {
		log.Panicf("Error parsing ID: %s", chirpIDStr)
		w.WriteHeader(500)
		return
	}
	dbChirp, err := cfg.db.GetChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, 404, "Chirp not found", err)
		return
	}

	chirp := Chirp{
		ID:        dbChirp.ID,
		UserID:    dbChirp.UserID,
		Body:      dbChirp.Body,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.CreatedAt,
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(chirp); err != nil {
		log.Panicf("Error encoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

}

func (cfg *apiConfig) handlerChirpCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("error decoding parameters: %s\n", err)
		w.WriteHeader(500)
		return
	}
	chripParams := database.CreateChirpParams{
		UserID: params.UserID,
		Body:   params.Body,
	}

	dbChirp, err := cfg.db.CreateChirp(r.Context(), chripParams)
	if err != nil {
		log.Printf("error writing to database: %s", err)
		w.WriteHeader(500)
		return
	}
	chirp := Chirp{
		ID:        dbChirp.ID,
		UserID:    dbChirp.UserID,
		Body:      dbChirp.Body,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.CreatedAt,
	}
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(&chirp); err != nil {
		log.Panicf("Error encoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}
}

type UserCredential struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

func (cfg *apiConfig) handlerUserLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	log.Printf("preparing to decode json...\n")
	decoder := json.NewDecoder(r.Body)
	log.Printf("created decoder\n")
	params := parameters{}
	log.Printf("decoding json...\n")
	err := decoder.Decode(&params)

	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}
	dbUser, err := cfg.db.GetUserForEmail(r.Context(), params.Email)
	if err != nil {
		log.Printf("No user forund for that email")
	}

	match, err := auth.CheckPasswordHash(params.Password, dbUser.HashedPassword)
	if err != nil {
		log.Println("Failed login")
		w.WriteHeader(http.StatusUnauthorized)
	}
	user := User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	}

	log.Printf("User after DB write:")
	fmt.Printf("%+v\n", user)
	if match {
		w.WriteHeader(http.StatusOK)
		log.Printf("Writing content type...")
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(user); err != nil {
			log.Panicf("Error encoding parameters: %s", err)
			w.WriteHeader(500)
			return
		}
	} else {
		w.WriteHeader(http.StatusUnauthorized)
	}
}

type CreateUserHandler struct {
	queries *database.Queries
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) handlerUserCreate(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling POST api/users")
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	log.Printf("preparing to decode json...\n")
	decoder := json.NewDecoder(r.Body)
	log.Printf("created decoder\n")
	params := parameters{}
	log.Printf("decoding json...\n")
	err := decoder.Decode(&params)

	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	// userHanlder, err := h.queries.CreateUser(r.Context(), params.Email)
	log.Printf("Getting user")
	log.Printf("Getting user - email: %v", params.Email)
	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		log.Printf("error hashing password")
	}

	userParams := database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
	}

	// dbUser, err := h.queries.CreateUser(r.Context())
	dbUser, err := cfg.db.CreateUser(r.Context(), userParams)
	if err != nil {
		log.Printf("error while getting user from db: %s", err)
	}
	log.Printf("local user struct")
	user := User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	}

	log.Printf("User after DB write:")
	fmt.Printf("%+v\n", user)
	w.WriteHeader(http.StatusCreated)
	log.Printf("Writing content type...")
	w.Header().Set("Content-Type", "application/json")
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
	db             *database.Queries
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

	godotenv.Load()
	dbURl := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURl)
	if err != nil {
		log.Fatal("error connecting to database")
	}
	dbQueries := database.New(db)

	mux := http.NewServeMux()
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	apiCfg := &apiConfig{}
	apiCfg.db = dbQueries
	wrappedHandler := apiCfg.middlewareMetricsInc(&RootHanlder{})
	mux.Handle("/app/", http.StripPrefix("/app", wrappedHandler))

	mux.Handle("GET /api/healthz", &HealthHanlder{})
	// mux.Handle("POST /api/users", &CreateUserHandler{})

	mux.HandleFunc("POST /api/validate_chirp", handlerChirpsValidate)
	mux.HandleFunc("POST /api/users", apiCfg.handlerUserCreate)
	mux.HandleFunc("POST /api/login", apiCfg.handlerUserLogin)
	mux.HandleFunc("GET /api/chirps", apiCfg.handlerGetChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerGetChirp)
	mux.HandleFunc("POST /api/chirps", apiCfg.handlerChirpCreate)

	metricsHndl := &MetricsHanlder{}
	metricsHndl.hits = &apiCfg.fileserverHits
	mux.Handle("GET /admin/metrics", metricsHndl)

	resetHndl := &ResetHanlder{}
	resetHndl.hits = &apiCfg.fileserverHits
	mux.Handle("POST /admin/reset", resetHndl)

	log.Println("Starting server on :8080")
	log.Fatal(srv.ListenAndServe())
}
