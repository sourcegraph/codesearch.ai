package socode

import (
	"html"
	"regexp"
	"sort"
	"strings"
)

var startCodeTag = regexp.MustCompile("<code>")
var endCodeTag = regexp.MustCompile("</code>")

type codeTagMatch struct {
	isStart    bool
	startIndex int
	endIndex   int
}

func GetCodeSnippetIndices(htmlString string) [][]int {
	startCodeTagMatches := startCodeTag.FindAllStringIndex(htmlString, -1)
	endCodeTagMatches := endCodeTag.FindAllStringIndex(htmlString, -1)

	indices := [][]int{}
	if len(startCodeTagMatches) != len(endCodeTagMatches) {
		return indices
	}

	matches := []*codeTagMatch{}
	for _, m := range startCodeTagMatches {
		matches = append(matches, &codeTagMatch{true, m[0], m[1]})
	}
	for _, m := range endCodeTagMatches {
		matches = append(matches, &codeTagMatch{false, m[0], m[1]})
	}

	sort.SliceStable(matches, func(i, j int) bool {
		return matches[i].startIndex < matches[j].startIndex
	})

	depth := 0
	var lastStartMatch *codeTagMatch
	for _, m := range matches {
		if m.isStart {
			if depth == 0 {
				lastStartMatch = m
			}

			depth += 1
		} else if depth == 0 {
			// End tag at depth 0 is invalid
			break
		} else {
			// !isStart && depth != 0
			depth -= 1

			if depth == 0 {
				indices = append(indices, []int{lastStartMatch.endIndex, m.startIndex})
			}
		}
	}

	return indices
}

func GetCodeSnippetsFromHTML(htmlString string) []string {
	indices := GetCodeSnippetIndices(htmlString)
	codeSnippets := make([]string, 0, len(indices))
	for _, codeSnippetIndices := range indices {
		codeSnippets = append(codeSnippets, htmlString[codeSnippetIndices[0]:codeSnippetIndices[1]])
	}
	return codeSnippets
}

func EscapeCodeSnippetsInHTML(htmlString string) string {
	indices := GetCodeSnippetIndices(htmlString)

	if len(indices) == 0 {
		return htmlString
	}

	htmlStringParts := []string{
		htmlString[:indices[0][0]],
	}
	for i := 0; i < len(indices); i++ {
		codeStart, codeEnd := indices[i][0], indices[i][1]
		var nextCodeStart int
		if i+1 == len(indices) {
			nextCodeStart = len(htmlString)
		} else {
			nextCodeStart = indices[i+1][0]
		}
		htmlStringParts = append(htmlStringParts, html.EscapeString(htmlString[codeStart:codeEnd]), htmlString[codeEnd:nextCodeStart])
	}

	return strings.Join(htmlStringParts, "")
}
