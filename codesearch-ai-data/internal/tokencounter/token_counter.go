package tokencounter

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"regexp"
	"strings"

	ph "codesearch-ai-data/internal/parsinghelpers"

	sitter "github.com/smacker/go-tree-sitter"
)

type TokenCounter map[string]int

func (tc TokenCounter) Add(token string) {
	count, ok := tc[token]
	if ok {
		tc[token] = count + 1
	} else {
		tc[token] = 1
	}
}

func (tc TokenCounter) addCount(token string, count int) {
	existingCount, ok := tc[token]
	if ok {
		tc[token] += existingCount + count
	} else {
		tc[token] = count
	}
}

func (tc TokenCounter) Extend(other TokenCounter) {
	for token, count := range other {
		tc.addCount(token, count)
	}
}

func (tc TokenCounter) Value() (driver.Value, error) {
	return json.Marshal(tc)
}

func (tc *TokenCounter) Scan(value any) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &tc)
}

// var SPECIAL_CHARS_REGEX = regexp.MustCompile("[-._’'\"{}()$/%`, " + `<>\\` + "]")
var SPECIAL_CHARS_REGEX = regexp.MustCompile(`[^\-_a-zA-Z0-9]`)
var VALID_TOKEN_REGEX = regexp.MustCompile("[_a-zA-Z][_a-zA-Z0-9]*")

func isValidToken(token string) bool {
	return len(token) > 1 && VALID_TOKEN_REGEX.MatchString(token)
}

func stringToTokens(str string) []string {
	tokens := SPECIAL_CHARS_REGEX.Split(str, -1)
	validTokens := []string{}
	for _, token := range tokens {
		tokenTrimmed := strings.TrimSpace(token)
		if isValidToken(tokenTrimmed) {
			validTokens = append(validTokens, tokenTrimmed)
		}
	}
	return validTokens
}

func isSingleLineNode(node *sitter.Node) bool {
	return node.StartPoint().Row == node.EndPoint().Row
}

func CountTokens(rootNode *sitter.Node, code []byte) TokenCounter {
	tokenCounter := TokenCounter{}

	nodesToVisit := []*sitter.Node{rootNode}
	var currentNode *sitter.Node
	for len(nodesToVisit) != 0 {
		currentNode, nodesToVisit = nodesToVisit[0], nodesToVisit[1:]
		currentNodeType := currentNode.Type()

		if strings.Contains(currentNodeType, "string") && isSingleLineNode(currentNode) {
			stringTokens := stringToTokens(currentNode.Content(code))
			for _, token := range stringTokens {
				tokenCounter.Add(token)
			}
		} else if ph.IsIdentifierType(currentNodeType) {
			identifier := strings.TrimSpace(currentNode.Content(code))
			if isValidToken(identifier) {
				tokenCounter.Add(identifier)
			}
		} else {
			children := make([]*sitter.Node, currentNode.ChildCount())
			for i := 0; i < int(currentNode.ChildCount()); i++ {
				children[i] = currentNode.Child(i)
			}
			nodesToVisit = append(children, nodesToVisit...)
		}
	}

	return tokenCounter
}
