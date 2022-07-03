package main

import (
	"codesearch-ai-data/internal/database"
	"codesearch-ai-data/internal/sg"
	"codesearch-ai-data/internal/socode"
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

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4"
	log "github.com/sirupsen/logrus"
)

type SearchResults struct {
	SOQuestionIDs        []int `json:"soQuestionIds"`
	ExtractedFunctionIDs []int `json:"extractedFunctionIds"`
}

var searchCache map[string]*SearchResults

func main() {
	ctx := context.Background()
	conn, err := database.ConnectToDatabase(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)

	searchCache = map[string]*SearchResults{}

	r := mux.NewRouter()
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	r.HandleFunc("/search", searchHandler(ctx, conn)).Methods("GET")
	r.HandleFunc("/", indexHandler(ctx, conn)).Methods("GET")
	http.Handle("/", r)

	log.Info("Starting server at port 8000")
	if err := http.ListenAndServe("0.0.0.0:8000", nil); err != nil {
		log.Fatal(err)
	}
}

type Result struct {
	Distance            float64
	SOQuestionID        int
	ExtractedFunctionID int
}

func indexHandler(ctx context.Context, conn *pgx.Conn) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var indexTemplate = template.Must(template.ParseFiles("templates/web/index.html", "templates/web/query.html"))
		if err := indexTemplate.Execute(w, nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func search(query string) (*SearchResults, error) {
	cachedResults, ok := searchCache[query]
	if ok {
		return cachedResults, nil
	}

	q := url.Values{}
	q.Add("query", query)

	resp, err := http.Get(fmt.Sprintf("http://localhost:8001/search?%s", q.Encode()))
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
	searchCache[query] = &result
	return &result, nil
}

func highlightCodeLineRanges(hefs []*web.HighlightedExtractedFunction) {
	wg := &sync.WaitGroup{}
	for _, hef := range hefs {
		wg.Add(1)
		go func(hef *web.HighlightedExtractedFunction) {
			defer wg.Done()
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

func searchHandler(ctx context.Context, conn *pgx.Conn) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Move outside
		var indexTemplate = template.Must(template.New("search.html").Funcs(template.FuncMap{
			"escapeCodeSnippets": func(html string) template.HTML {
				return template.HTML(socode.EscapeCodeSnippetsInHTML(html))
			},
		}).ParseFiles("templates/web/search.html", "templates/web/query.html"))

		q := strings.TrimSpace(r.URL.Query().Get("q"))
		if len(q) > 256 {
			q = q[:256]
		}

		searchResults, err := search(q)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		soqs, err := web.GetSOQuestionsWithAnswersByID(ctx, conn, searchResults.SOQuestionIDs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		hefs, err := web.GetExtractedFunctionsByID(ctx, conn, searchResults.ExtractedFunctionIDs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		highlightCodeLineRanges(hefs)

		templateData := struct {
			Query     string
			Functions []*web.HighlightedExtractedFunction
			Questions []*web.SOQuestionWithAnswers
		}{
			Query:     q,
			Functions: hefs,
			Questions: soqs,
		}

		if err := indexTemplate.Execute(w, templateData); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
