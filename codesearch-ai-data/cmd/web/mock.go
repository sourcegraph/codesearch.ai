package main

import (
	"codesearch-ai-data/internal/database"
	"codesearch-ai-data/internal/web"
	"context"
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func mockExtractedFunctionsSearchHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	conn, err := database.ConnectToDatabase(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)

	searchResults := &SearchResults{IDs: []int{1, 100, 1000}}

	hefs, err := web.GetExtractedFunctionsByID(ctx, conn, searchResults.IDs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	highlightCodeLineRanges(hefs)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(hefs)
}

func mockSoSearchHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	conn, err := database.ConnectToDatabase(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)

	searchResults := &SearchResults{IDs: []int{59628, 91367, 143233}}

	soqs, err := web.GetSOQuestionsWithAnswersByID(ctx, conn, searchResults.IDs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(soqs)
}
