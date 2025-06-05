package main

import (
	"net/http"
	"sync/atomic"
	"fmt"
)

type apiConfig struct {
	fileServerHits atomic.Int32
}

func (cfg *apiConfig) middleWareMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileServerHits.Add(1)
		next.ServeHTTP(w,r )
	})
}

func (cfg *apiConfig) counter(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	count := fmt.Sprintf("Hits: %d\n", cfg.fileServerHits.Load())
	w.Write([]byte(count))
}

func (cfg *apiConfig) reset(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	cfg.fileServerHits.Store(0)
}

func myHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func main() {
	cfg := apiConfig{}
	serverMux := http.NewServeMux()
	serverMux.Handle("/app/", http.StripPrefix("/app/", cfg.middleWareMetrics(http.FileServer(http.Dir(".")))))
	serverMux.HandleFunc("/healthz", myHandler)
	serverMux.HandleFunc("/metrics", cfg.counter)
	serverMux.HandleFunc("/reset", cfg.counter)
	server := http.Server{Handler: serverMux, Addr: ":8080"}
	server.ListenAndServe()
}
