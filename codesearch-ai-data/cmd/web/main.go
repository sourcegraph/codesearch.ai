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
	// API routes
	if isDevelopment {
		r.HandleFunc("/api/search/functions/by-text", mockSearchHandler("functions")).Methods("GET")
		r.HandleFunc("/api/search/functions/by-code", mockSearchHandler("functions")).Methods("GET")

		r.HandleFunc("/api/search/so/by-text", mockSearchHandler("so")).Methods("GET")
		r.HandleFunc("/api/search/so/by-code", mockSearchHandler("so")).Methods("GET")
	} else {
		r.HandleFunc("/api/search/functions/by-text", searchFunctionsByTextHandler).Methods("GET")
		r.HandleFunc("/api/search/functions/by-code", searchFunctionsByCodeHandler).Methods("GET")

		r.HandleFunc("/api/search/so/by-text", searchSOByTextHandler).Methods("GET")
		r.HandleFunc("/api/search/so/by-code", searchSOByCodeHandler).Methods("GET")
	}
	http.Handle("/", r)

	log.Info("Starting server at port 8000")
	if err := http.ListenAndServe("0.0.0.0:8000", nil); err != nil {
		log.Fatal(err)
	}
}
