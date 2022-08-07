package main

import (
	"codesearch-ai-data/internal/database"
	"codesearch-ai-data/internal/web"
	"context"
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func mockSearchHandler(dataSource string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		conn, err := database.ConnectToDatabase(ctx)
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close(ctx)

		var results any
		if dataSource == "functions" {
			ids := []int{1, 100, 1000}
			hefs, err := web.GetExtractedFunctionsByID(ctx, conn, ids)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			highlightCodeLineRanges(hefs)
			results = hefs
		} else if dataSource == "so" {
			ids := []int{1006395, 1243079, 1163074}
			results, err = web.GetSOQuestionsWithAnswersByID(ctx, conn, ids)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(results)
	}
}
