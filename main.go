package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
)

type JSONResponse struct {
	CleanedBody string `json:"cleaned_body,omitempty"`
}

type JSONError struct {
	Error       string `json:"error,omitempty"`
}

type Chirp struct {
	Body string `json:"body"`
}

func (c *Chirp) cleanedWord(w string) string {
	invalidWords := []string{"kerfuffle", "sharbert", "fornax"}
	for _, invalidWord := range invalidWords {
		sanitizedWord := strings.ToLower(w)
		if invalidWord == sanitizedWord {
			return "****"
		}
	}
	return w
}

func (c *Chirp) cleanBody() string {
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
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	cfg.fileServerHits.Store(0)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func responsWithJSON(w http.ResponseWriter, code int, payload JSONResponse) {
	w.Header().Set("Content-Type", "application/json")
	response, _  := json.Marshal(payload)
	w.WriteHeader(code)
	w.Write(response)
}

func responsWithJSONError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	error := JSONError{Error: msg}
	response, _  := json.Marshal(error)
	w.WriteHeader(code)
	w.Write(response)
}

func validateChirp(w http.ResponseWriter, r *http.Request) {
	data := Chirp{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&data)
	if err != nil {
		log.Printf("Error decoding: %s\n", err)
		responsWithJSONError(w, 500, "Something went wrong")
		return
	}
	fmt.Printf("chirp: %s\n", data.Body)

	if len(data.Body) > 140 {
		responsWithJSONError(w, 400, "Chirp is too long")
		return
	}
	resp := JSONResponse{CleanedBody: data.cleanBody()}
	responsWithJSON(w, 200, resp)
}

func main() {
	cfg := apiConfig{}
	serverMux := http.NewServeMux()
	serverMux.Handle("/app/", http.StripPrefix("/app/", cfg.middleWareMetrics(http.FileServer(http.Dir(".")))))
	serverMux.HandleFunc("GET /api/healthz", healthCheck)
	serverMux.HandleFunc("POST /api/validate_chirp", validateChirp)
	serverMux.HandleFunc("GET /admin/metrics", cfg.counter)
	serverMux.HandleFunc("POST /admin/reset", cfg.reset)
	server := http.Server{Handler: serverMux, Addr: ":8080"}
	server.ListenAndServe()
}
