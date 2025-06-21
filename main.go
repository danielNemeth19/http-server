package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"

	"github.com/danielNemeth19/http-server/internal/auth"
	"github.com/danielNemeth19/http-server/internal/database"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type JSONError struct {
	Error string `json:"error,omitempty"`
}

type UserRequestParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

type ChirpRequestParams struct {
	Body   string        `json:"body"`
	UserID uuid.NullUUID `json:"user_id"`
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (c *ChirpRequestParams) cleanedWord(w string) string {
	invalidWords := []string{"kerfuffle", "sharbert", "fornax"}
	for _, invalidWord := range invalidWords {
		sanitizedWord := strings.ToLower(w)
		if invalidWord == sanitizedWord {
			return "****"
		}
	}
	return w
}

func (c *ChirpRequestParams) cleanBody() string {
	var cleanedWords []string
	words := strings.Split(c.Body, " ")

	for _, word := range words {
		cleanedWord := c.cleanedWord(word)
		cleanedWords = append(cleanedWords, cleanedWord)
	}
	return strings.Join(cleanedWords, " ")
}

type apiConfig struct {
	fileServerHits atomic.Int32
	db             *database.Queries
	env            string
}

func (cfg *apiConfig) middleWareMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileServerHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) counter(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)
	count := fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", cfg.fileServerHits.Load())
	w.Write([]byte(count))
}

func (cfg *apiConfig) reset(w http.ResponseWriter, r *http.Request) {
	if cfg.env != "dev" {
		responsWithJSONError(w, 403, "Forbidden")
	}
	rowCount, err := cfg.db.DeleteUsers(r.Context())
	if err != nil {
		log.Printf("Error deleting users: %s\n", err)
		responsWithJSONError(w, 500, "Something went wrong during db save")
		return
	}
	log.Printf("Deleted %d number of users\n", rowCount)
	responsWithJSON(w, 200, rowCount)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func responsWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	response, _ := json.Marshal(payload)
	w.WriteHeader(code)
	w.Write(response)
}

func responsWithJSONError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	error := JSONError{Error: msg}
	response, _ := json.Marshal(error)
	w.WriteHeader(code)
	w.Write(response)
}

func (cfg *apiConfig) addChirp(w http.ResponseWriter, r *http.Request) {
	data := ChirpRequestParams{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&data)
	if err != nil {
		log.Printf("Error decoding: %s\n", err)
		responsWithJSONError(w, 500, "Something went wrong")
		return
	}
	log.Printf("chirp: %s\n", data.Body)

	if len(data.Body) > 140 {
		responsWithJSONError(w, 400, "Chirp is too long")
		return
	}
	chirpParams := database.CreateChirpParams{
		Body:   data.cleanBody(),
		UserID: data.UserID,
	}
	chirp, err := cfg.db.CreateChirp(r.Context(), chirpParams)
	if err != nil {
		log.Printf("Error decoding: %s\n", err)
		responsWithJSONError(w, 500, "Database saved failed")
		return
	}
	chirpCreated := Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID.UUID,
	}
	responsWithJSON(w, 201, chirpCreated)
}

func (cfg *apiConfig) getChirps(w http.ResponseWriter, r *http.Request) {
	chirpsDB, err := cfg.db.GetChirps(r.Context())
	if err != nil {
		log.Printf("Error running database query: %s\n", err)
		responsWithJSONError(w, 500, "Something went wrong")
		return
	}
	var chirps []Chirp
	for _, msg := range chirpsDB {
		chirp := Chirp{
			ID:        msg.ID,
			CreatedAt: msg.CreatedAt,
			UpdatedAt: msg.UpdatedAt,
			Body:      msg.Body,
			UserID:    msg.UserID.UUID,
		}
		chirps = append(chirps, chirp)

	}
	responsWithJSON(w, 200, chirps)
}

func (cfg *apiConfig) getChirp(w http.ResponseWriter, r *http.Request) {
	param := r.PathValue("chirpID")
	uuidParam, err := uuid.Parse(param)
	if err != nil {
		log.Printf("Error parsing incoming param as UUID: %s\n", err)
		responsWithJSONError(w, 400, "Bad value for UUID")
		return
	}
	msg, err := cfg.db.GetChirp(r.Context(), uuidParam)
	if err == sql.ErrNoRows {
		log.Printf("Not found error: %s\n", err)
		responsWithJSONError(w, 404, "No record has been found")
		return
	} else if err != nil {
		log.Printf("Error running database query: %s\n", err)
		responsWithJSONError(w, 500, "Something went wrong")
		return
	}
	chirp := Chirp{
		ID:        msg.ID,
		CreatedAt: msg.CreatedAt,
		UpdatedAt: msg.UpdatedAt,
		Body:      msg.Body,
		UserID:    msg.UserID.UUID,
	}
	responsWithJSON(w, 200, chirp)
}

func (cfg *apiConfig) createUser(w http.ResponseWriter, r *http.Request) {
	data := UserRequestParams{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&data)
	if err != nil {
		log.Printf("Error decoding: %s\n", err)
		responsWithJSONError(w, 500, "Params could not be parsed")
		return
	}
	hashedPassword, err := auth.HashPassword(data.Password)
	if err != nil {
		log.Printf("Error hashing password: %s\n", err)
		responsWithJSONError(w, 500, "Error hashing password")
	}
	params := database.CreateUserParams{
		Email:          sql.NullString{String: data.Email, Valid: true},
		HashedPassword: sql.NullString{String: hashedPassword, Valid: true},
	}
	user, err := cfg.db.CreateUser(r.Context(), params)
	if err != nil {
		log.Printf("Error creating user: %s\n", err)
		responsWithJSONError(w, 500, "Something went wrong during db save")
		return
	}
	userResponse := User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email.String,
	}
	log.Printf("User created: %s\n", userResponse.Email)
	responsWithJSON(w, 201, userResponse)
}

func (cfg *apiConfig) login(w http.ResponseWriter, r *http.Request) {
	data := UserRequestParams{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&data)
	if err != nil {
		log.Printf("Error decoding: %s\n", err)
		responsWithJSONError(w, 500, "Params could not be parsed")
		return
	}
	user, err := cfg.db.GetUser(r.Context(), sql.NullString{String: data.Email, Valid: true})
	if err != nil {
		log.Printf("User lookup failed in database: %s\n", err)
		responsWithJSONError(w, 401, "Unauthorized")
		return
	}
	err = auth.CheckPasswordHash(user.HashedPassword.String, data.Password)
	if err != nil {
		log.Printf("Password check failed: %s\n", err)
		responsWithJSONError(w, 401, "Unauthorized")
		return
	}
	userResponse := User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email.String,
	}
	log.Printf("Login successfull for user %s\n", userResponse.Email)
	responsWithJSON(w, 200, userResponse)
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	env := os.Getenv("PLATFORM")
	db, err := sql.Open("postgres", dbURL)
	dbQueries := database.New(db)

	if err != nil {
		log.Printf("Error connecting: %s\n", err)
	}
	cfg := apiConfig{db: dbQueries, env: env}
	serverMux := http.NewServeMux()
	serverMux.Handle("/app/", http.StripPrefix("/app/", cfg.middleWareMetrics(http.FileServer(http.Dir(".")))))
	serverMux.HandleFunc("GET /api/healthz", healthCheck)
	serverMux.HandleFunc("POST /api/chirps", cfg.addChirp)
	serverMux.HandleFunc("GET /api/chirps", cfg.getChirps)
	serverMux.HandleFunc("GET /api/chirps/{chirpID}", cfg.getChirp)
	serverMux.HandleFunc("POST /api/users", cfg.createUser)
	serverMux.HandleFunc("POST /api/login", cfg.login)
	serverMux.HandleFunc("GET /admin/metrics", cfg.counter)
	serverMux.HandleFunc("POST /admin/reset", cfg.reset)
	server := http.Server{Handler: serverMux, Addr: ":8080"}
	server.ListenAndServe()
}
