package parsinghelpers

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

var commentPrefixes = []string{"/**", "/*", "*", "#", "//", "\"\"\"", "\""}
var commentSuffixes = []string{"*/", "\"\"\"", "\""}

func isCommentNode(node *sitter.Node) bool {
	nodeType := node.Type()
	return nodeType == "comment" || nodeType == "line_comment" || nodeType == "block_comment"
}

func StripCommentDelimiters(comment string) string {
	commentLines := strings.Split(comment, "\n")
	strippedCommentLines := make([]string, 0, len(commentLines))
	for _, commentLine := range commentLines {
		trimmedLine := strings.TrimSpace(commentLine)

		for _, suffix := range commentSuffixes {
			if strings.HasSuffix(trimmedLine, suffix) {
				trimmedLine = strings.TrimSuffix(trimmedLine, suffix)
			}
		}

		for _, prefix := range commentPrefixes {
			if strings.HasPrefix(trimmedLine, prefix) {
				trimmedLine = strings.TrimPrefix(trimmedLine, prefix)
			}
		}

		trimmedLine = strings.TrimSpace(trimmedLine)

		if len(trimmedLine) > 0 {
			strippedCommentLines = append(strippedCommentLines, trimmedLine)
		}
	}

	return strings.Join(strippedCommentLines, " ")
}

func StripCommentNodesDelimiters(commentNodes []*sitter.Node, code []byte) []string {
	comments := make([]string, 0, len(commentNodes))
	for _, commentNode := range commentNodes {
		strippedComment := StripCommentDelimiters(commentNode.Content(code))
		if len(strippedComment) > 0 {
			comments = append(comments, strippedComment)
		}
	}
	return comments
}
