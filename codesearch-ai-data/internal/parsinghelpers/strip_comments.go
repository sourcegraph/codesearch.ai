package parsinghelpers

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

type SkipNodeFn func(node *sitter.Node) bool

func StripComments(rootNode *sitter.Node, skipNodeFn SkipNodeFn) ([]*sitter.Node, []*sitter.Node) {
	leafNodes := []*sitter.Node{}
	nodesToVisit := []*sitter.Node{rootNode}

	var currentNode *sitter.Node
	for len(nodesToVisit) != 0 {
		currentNode, nodesToVisit = nodesToVisit[0], nodesToVisit[1:]

		// We consider string types as leaf nodes since they can contain further children (e.g. interpolation)
		// but tree-sitter does not provide a way to iterate over string "parts". So we have to do a separate check
		// for string types.
		if strings.Contains(currentNode.Type(), "string") || currentNode.ChildCount() == 0 {
			leafNodes = append(leafNodes, currentNode)
			continue
		}

		children := make([]*sitter.Node, currentNode.ChildCount())
		for i := 0; i < int(currentNode.ChildCount()); i++ {
			children[i] = currentNode.Child(i)
		}
		nodesToVisit = append(children, nodesToVisit...)
	}

	filteredNodes := []*sitter.Node{}
	commentNodes := []*sitter.Node{}
	for _, node := range leafNodes {
		if isCommentNode(node) {
			commentNodes = append(commentNodes, node)
			continue
		}
		if node.Type() == "\n" || (skipNodeFn != nil && skipNodeFn(node)) {
			continue
		}
		filteredNodes = append(filteredNodes, node)
	}

	return filteredNodes, commentNodes
}
