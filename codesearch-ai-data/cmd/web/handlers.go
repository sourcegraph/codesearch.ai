package main

import (
	"codesearch-ai-data/internal/database"
	"codesearch-ai-data/internal/web"
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

const MAX_RESULTS = 20

var languagesRegexp = regexp.MustCompile(`(?i)\b(python|java|javascript|js|py|go|golang|ruby|php)\b`)

func sliceQuery(query string) string {
	trimmedQuery := strings.TrimSpace(query)
	if len(trimmedQuery) > 512 {
		return trimmedQuery[:512]
	}
	return trimmedQuery
}

func transformCodeQuery(codeQuery string) string {
	return strings.ReplaceAll(codeQuery, "\t", " ")
}

func findLanguage(query string) string {
	return strings.ToLower(languagesRegexp.FindString(query))
}

func languageToFileExtension(language string) string {
	switch language {
	case "python":
	case "py":
		return "py"
	case "javascript":
	case "js":
		return "js"
	case "ruby":
		return "rb"
	case "java":
		return "java"
	case "php":
		return "php"
	case "go":
	case "golang":
		return "go"
	}
	return ""
}

func searchFunctionsByTextHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	conn, err := database.ConnectToDatabase(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)

	query := sliceQuery(r.URL.Query().Get("query"))
	searchResults, err := search("functions", "text", query, 512)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	results, err := web.GetExtractedFunctionsByID(ctx, conn, searchResults.IDs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	language := findLanguage(query)
	languageExtension := languageToFileExtension(language)
	filteredResults := make([]*web.HighlightedExtractedFunction, 0, MAX_RESULTS)
	for _, result := range results {
		if strings.HasSuffix(result.FilePath, languageExtension) {
			filteredResults = append(filteredResults, result)
		}
		if len(filteredResults) == MAX_RESULTS {
			break
		}
	}
	highlightCodeLineRanges(filteredResults)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(filteredResults)
}

func searchFunctionsByCodeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	conn, err := database.ConnectToDatabase(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)

	query := transformCodeQuery(sliceQuery(r.URL.Query().Get("query")))
	searchResults, err := search("functions", "code", query, MAX_RESULTS)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	results, err := web.GetExtractedFunctionsByID(ctx, conn, searchResults.IDs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	highlightCodeLineRanges(results)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(results)
}

func searchSOByTextHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	conn, err := database.ConnectToDatabase(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)

	query := sliceQuery(r.URL.Query().Get("query"))
	searchResults, err := search("so", "text", query, 512)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	results, err := web.GetSOQuestionsWithAnswersByID(ctx, conn, searchResults.IDs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	language := findLanguage(query)
	filteredResults := make([]*web.SOQuestionWithAnswers, 0, MAX_RESULTS)
	for _, result := range results {
		if strings.Contains(strings.ToLower(result.Tags), language) {
			filteredResults = append(filteredResults, result)
		}
		if len(filteredResults) == MAX_RESULTS {
			break
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(filteredResults)
}

func searchSOByCodeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	conn, err := database.ConnectToDatabase(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)

	query := transformCodeQuery(sliceQuery(r.URL.Query().Get("query")))
	searchResults, err := search("so", "code", query, MAX_RESULTS)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	results, err := web.GetSOQuestionsWithAnswersByID(ctx, conn, searchResults.IDs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(results)
}
