package sg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

var SOURCEGRAPH_API_TOKEN = os.Getenv("SOURCEGRAPH_API_TOKEN")

func requestGraphQL(token, query string, variables map[string]any, target any) error {
	body, err := json.Marshal(map[string]any{
		"query":     query,
		"variables": variables,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://sourcegraph.com/.api/graphql", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("token %s", token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%d: %s", resp.StatusCode, string(body))
	}

	if target == nil {
		return nil
	}

	return json.Unmarshal(body, &target)
}

const HIGHLIGHT_QUERY_GRAPHQL = `
query (
  $repoName: String!
  $commitID: String!
  $filePath: String!
  $ranges: [HighlightLineRange!]!
) {
	repository(name: $repoName) {
		commit(rev: $commitID) {
			file(path: $filePath) {
				highlight(disableTimeout: false, isLightTheme: true) {
					lineRanges(ranges: $ranges)
				}
			}
		}
	}
}
`

type highlightLineRange struct {
	StartLine int `json:"startLine"`
	EndLine   int `json:"endLine"`
}

type highlightQueryResponse struct {
	Data struct {
		Repository struct {
			Commit struct {
				File struct {
					Highlight struct {
						LineRanges [][]string `json:"lineRanges"`
					} `json:"highlight"`
				} `json:"file"`
			} `json:"commit"`
		} `json:"repository"`
	} `json:"data"`
}

var highlightCache = map[string]string{}

func GetHighlightedCodeLineRange(repoName string, commitID string, filePath string, startLine int, endLine int) (string, error) {
	cacheKey := fmt.Sprintf("%s:%s:%s:%d:%d", repoName, commitID, filePath, startLine, endLine)
	cachedHighlightedCode, ok := highlightCache[cacheKey]
	if ok {
		return cachedHighlightedCode, nil
	}

	variables := map[string]any{
		"repoName": repoName,
		"commitID": commitID,
		"filePath": filePath,
		"ranges":   []highlightLineRange{{StartLine: startLine, EndLine: endLine}},
	}

	var resp highlightQueryResponse
	err := requestGraphQL(SOURCEGRAPH_API_TOKEN, HIGHLIGHT_QUERY_GRAPHQL, variables, &resp)
	if err != nil {
		return "", nil
	}
	highlightedCode := strings.Join(resp.Data.Repository.Commit.File.Highlight.LineRanges[0], "")
	highlightCache[cacheKey] = highlightedCode
	return highlightedCode, nil
}
