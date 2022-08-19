package tokencounter

import (
	"regexp"
	"strings"

	ph "codesearch-ai-data/internal/parsinghelpers"

	"github.com/fatih/camelcase"
	sitter "github.com/smacker/go-tree-sitter"
)

type tokenCounter map[string]int

func (tc tokenCounter) Add(token string) {
	count, ok := tc[token]
	if ok {
		tc[token] = count + 1
	} else {
		tc[token] = 1
	}
}

var SPECIAL_CHARS_REGEX = regexp.MustCompile("[._'\"{}$/%` " + `\\` + "]")
var DOT_UNDERSCORE_REGEX = regexp.MustCompile("[._]")
var VALID_TOKEN_REGEX = regexp.MustCompile("[_a-zA-Z][_a-zA-Z0-9]*")

func identifierToTokens(identifier string) []string {
	if len(identifier) == 0 {
		return []string{}
	}

	tokens := DOT_UNDERSCORE_REGEX.Split(identifier, -1)
	validIdentifierTokens := []string{}
	for _, token := range tokens {
		camelCaseTokens := camelcase.Split(token)
		for _, camelCaseToken := range camelCaseTokens {
			camelCaseTokenTrimmed := strings.TrimSpace(camelCaseToken)
			if len(camelCaseTokenTrimmed) == 0 {
				continue
			}
			if VALID_TOKEN_REGEX.MatchString(camelCaseTokenTrimmed) {
				validIdentifierTokens = append(validIdentifierTokens, camelCaseTokenTrimmed)
			}
		}
	}

	return validIdentifierTokens
}

func stringToTokens(str string) []string {
	tokens := SPECIAL_CHARS_REGEX.Split(str, -1)
	validTokens := []string{}
	for _, token := range tokens {
		camelCaseTokenTrimmed := strings.TrimSpace(token)
		if len(camelCaseTokenTrimmed) == 0 {
			continue
		}
		if VALID_TOKEN_REGEX.MatchString(token) {
			validTokens = append(validTokens, camelCaseTokenTrimmed)
		}
	}
	return validTokens
}

func isSingleLineNode(node *sitter.Node) bool {
	return node.StartPoint().Row == node.EndPoint().Row
}

func countTokens(rootNode *sitter.Node, code []byte) (tokenCounter, error) {
	tokenCounter := tokenCounter{}

	nodesToVisit := []*sitter.Node{rootNode}
	var currentNode *sitter.Node
	for len(nodesToVisit) != 0 {
		currentNode, nodesToVisit = nodesToVisit[0], nodesToVisit[1:]
		currentNodeType := currentNode.Type()

		if strings.Contains(currentNodeType, "string") && isSingleLineNode(currentNode) {
			stringTokens := stringToTokens(currentNode.Content(code))
			for _, token := range stringTokens {
				tokenCounter.Add(strings.ToLower(token))
			}
		} else if ph.IsIdentifierType(currentNodeType) {
			identifierTokens := identifierToTokens(currentNode.Content(code))
			for _, token := range identifierTokens {
				tokenCounter.Add(strings.ToLower(token))
			}
		} else {
			children := make([]*sitter.Node, currentNode.ChildCount())
			for i := 0; i < int(currentNode.ChildCount()); i++ {
				children[i] = currentNode.Child(i)
			}
			nodesToVisit = append(children, nodesToVisit...)
		}
	}

	return tokenCounter, nil
}
