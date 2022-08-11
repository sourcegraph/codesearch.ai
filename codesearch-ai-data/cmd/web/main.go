package main

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func main() {
	isDevelopment := os.Getenv("DEVELOPMENT") == "true"

	r := mux.NewRouter()
	if isDevelopment {
		// Mock API routes
		r.HandleFunc("/api/search/functions/by-text", mockSearchHandler("functions")).Methods("GET", "OPTIONS")
		r.HandleFunc("/api/search/functions/by-code", mockSearchHandler("functions")).Methods("GET", "OPTIONS")

		r.HandleFunc("/api/search/so/by-text", mockSearchHandler("so")).Methods("GET", "OPTIONS")
		r.HandleFunc("/api/search/so/by-code", mockSearchHandler("so")).Methods("GET", "OPTIONS")
	} else {
		// Static files
		r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./client/build/static"))))

		// API routes
		r.HandleFunc("/api/search/functions/by-text", searchFunctionsByTextHandler).Methods("GET", "OPTIONS")
		r.HandleFunc("/api/search/functions/by-code", searchFunctionsByCodeHandler).Methods("GET", "OPTIONS")

		r.HandleFunc("/api/search/so/by-text", searchSOByTextHandler).Methods("GET", "OPTIONS")
		r.HandleFunc("/api/search/so/by-code", searchSOByCodeHandler).Methods("GET", "OPTIONS")

		r.Path("/favicon.png").HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "client/build/favicon.png") }).Methods("GET")

		// Index
		r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "client/build/index.html") }).Methods("GET")
	}
	http.Handle("/", r)

	log.Info("Starting server at port 8000")
	if err := http.ListenAndServe("0.0.0.0:8000", nil); err != nil {
		log.Fatal(err)
	}
}
