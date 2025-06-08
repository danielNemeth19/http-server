package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type Chirp struct {
	Body string `json:"body"`
}

type JSONResponse struct {
	Valid bool   `json:"valid,omitempty"`
	Error string `json:"error,omitempty"`
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

func validateChirp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	data := Chirp{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&data)
	if err != nil {
		log.Printf("Error decoding: %s\n", err)
		toResp := JSONResponse{Error: "Something went wrong"}
		errMessage, _ := json.Marshal(toResp)
		w.WriteHeader(500)
		w.Write(errMessage)
		return
	}
	fmt.Printf("chirp: %s\n", data.Body)

	if len(data.Body) > 140 {
		toResp := JSONResponse{Error: "Chirp is too long"}
		errMessage, _ := json.Marshal(toResp)
		w.WriteHeader(400)
		w.Write(errMessage)
		return
	}

	resp := JSONResponse{Valid: true}
	responseJson, _ := json.Marshal(resp)
	w.WriteHeader(200)
	w.Write(responseJson)
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
