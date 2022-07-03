package parsinghelpers

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func PrettyFormatNodes(nodes []*sitter.Node, code []byte) string {
	if len(nodes) == 0 {
		return ""
	}

	initialColumnOffset := int(nodes[0].StartPoint().Column)
	for i := 1; i < len(nodes); i++ {
		curStartColumn := int(nodes[i].StartPoint().Column)
		if curStartColumn < initialColumnOffset {
			initialColumnOffset = curStartColumn
		}
	}

	prettyPrinted := []string{
		nodes[0].Content(code),
	}

	for i := 1; i < len(nodes); i++ {
		prev, cur := nodes[i-1], nodes[i]
		sameLine := prev.EndPoint().Row == cur.StartPoint().Row

		if sameLine {
			diff := int(cur.StartPoint().Column) - int(prev.EndPoint().Column)
			prettyPrinted = append(prettyPrinted, strings.Repeat(" ", min(diff, 1)))
		} else {
			columnOffset := max(int(cur.StartPoint().Column)-initialColumnOffset, 0)
			prettyPrinted = append(prettyPrinted, "\n", strings.Repeat(" ", columnOffset))
		}

		prettyPrinted = append(prettyPrinted, cur.Content(code))
	}

	return strings.TrimSpace(strings.Join(prettyPrinted, ""))
}
