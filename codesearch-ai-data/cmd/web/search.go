package main

import (
	"codesearch-ai-data/internal/database"
	"codesearch-ai-data/internal/sg"
	"codesearch-ai-data/internal/web"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
)

type SearchResults struct {
	IDs []int `json:"ids"`
}

var searchCache map[string]*SearchResults

func search(query string, set string) (*SearchResults, error) {
	cacheKey := fmt.Sprintf("%s:%s", set, query)
	cachedResults, ok := searchCache[cacheKey]
	if ok {
		return cachedResults, nil
	}

	q := url.Values{}
	q.Add("query", query)

	resp, err := http.Get(fmt.Sprintf("http://localhost:8001/search/%s?%s", set, q.Encode()))
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result SearchResults
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	searchCache[cacheKey] = &result
	return &result, nil
}

func highlightCodeLineRanges(hefs []*web.HighlightedExtractedFunction) {
	wg := &sync.WaitGroup{}
	for _, hef := range hefs {
		wg.Add(1)
		go func(hef *web.HighlightedExtractedFunction) {
			defer wg.Done()
			if hef == nil {
				log.Error("nil extracted function")
				return
			}
			highlightedCode, err := sg.GetHighlightedCodeLineRange(hef.RepositoryName, hef.CommitID, hef.FilePath, hef.StartLine, hef.EndLine+1)
			if err != nil {
				log.Error("Error highlighting code ", err)
				return
			}
			hef.HighlightedHTML = template.HTML(fmt.Sprintf("<code><table>%s</table></code>", highlightedCode))
		}(hef)
	}
	wg.Wait()
}

func extractedFunctionsSearchHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	conn, err := database.ConnectToDatabase(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)

	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if len(q) > 256 {
		q = q[:256]
	}

	searchResults, err := search(q, "extracted-functions")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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

func soSearchHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	conn, err := database.ConnectToDatabase(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)

	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if len(q) > 256 {
		q = q[:256]
	}

	searchResults, err := search(q, "so")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	soqs, err := web.GetSOQuestionsWithAnswersByID(ctx, conn, searchResults.IDs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(soqs)
}
