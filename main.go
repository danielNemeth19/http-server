package main

import (
	"net/http"
)

func myHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func main() {
	serverMux := http.NewServeMux()
	serverMux.Handle("/app/", http.StripPrefix("/app/", http.FileServer(http.Dir("."))))
	serverMux.HandleFunc("/healthz", myHandler)
	server := http.Server{Handler: serverMux, Addr: ":8080"}
	server.ListenAndServe()
}
