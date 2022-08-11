package main

import (
	"codesearch-ai-data/internal/sg"
	"codesearch-ai-data/internal/web"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"sync"

	log "github.com/sirupsen/logrus"
)

const ML_API_BASE_URL = "http://localhost:8001"

type SearchResults struct {
	IDs []int `json:"ids"`
}

var searchCacheMutex sync.Mutex
var searchCache = map[string]*SearchResults{}

func search(source string, by string, query string, count int) (*SearchResults, error) {
	cacheKey := fmt.Sprintf("%s:%s:%s:%d", source, by, query, count)
	searchCacheMutex.Lock()
	cachedResults, ok := searchCache[cacheKey]
	searchCacheMutex.Unlock()
	if ok {
		return cachedResults, nil
	}

	q := url.Values{}
	q.Add("query", query)
	q.Add("count", strconv.Itoa(count))

	resp, err := http.Get(fmt.Sprintf("%s/search/%s/by-%s?%s", ML_API_BASE_URL, source, by, q.Encode()))
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

	searchCacheMutex.Lock()
	searchCache[cacheKey] = &result
	searchCacheMutex.Unlock()

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
