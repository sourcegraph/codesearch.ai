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
	// Static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./client/build/static"))))
	// API routes
	if isDevelopment {
		r.HandleFunc("/api/search/extracted-functions", mockExtractedFunctionsSearchHandler).Methods("GET")
		r.HandleFunc("/api/search/so", mockSoSearchHandler).Methods("GET")
	} else {
		r.HandleFunc("/api/search/extracted-functions", extractedFunctionsSearchHandler).Methods("GET")
		r.HandleFunc("/api/search/so", soSearchHandler).Methods("GET")
	}
	// Index
	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "client/build/index.html") }).Methods("GET")

	http.Handle("/", r)

	log.Info("Starting server at port 8000")
	if err := http.ListenAndServe("0.0.0.0:8000", nil); err != nil {
		log.Fatal(err)
	}
}
