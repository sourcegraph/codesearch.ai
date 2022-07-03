package parsinghelpers

import (
	"errors"
	"io"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

func areNodeLinesConsecutive(prevNode *sitter.Node, currentNode *sitter.Node) bool {
	// Caller has to ensure currentNode and prevNode are non-nil.
	currentNodeStartRow := currentNode.StartPoint().Row
	prevNodeEndRow := prevNode.EndPoint().Row
	return currentNodeStartRow-1 == prevNodeEndRow
}

func GetPrecedingFunctionDocstring(functionNode *sitter.Node, sourceCode []byte) string {
	if functionNode == nil {
		return ""
	}

	prevNode := functionNode.PrevNamedSibling()

	// If the function node does not have a previous consecutive comment node, then it does not have docstring.
	if prevNode == nil || !isCommentNode(prevNode) || !areNodeLinesConsecutive(prevNode, functionNode) {
		return ""
	}

	// We only consider previous comments of the same type (we do not mix line and block comments).
	commentType := prevNode.Type()
	comments := []string{}
	for prevNode != nil && prevNode.Type() == commentType {
		comment := StripCommentDelimiters(prevNode.Content(sourceCode))
		// Prepend comment to existing comments since we are traversing in the reverse order (bottom up).
		comments = append([]string{comment}, comments...)

		// Check that the start of the current comment matches the end of the previous comment,
		// i.e. stop when there is an empty line between comments.
		prevPrevNode := prevNode.PrevNamedSibling()
		if prevPrevNode != nil && !areNodeLinesConsecutive(prevPrevNode, prevNode) {
			break
		}
		prevNode = prevPrevNode
	}
	return strings.Join(comments, " ")
}

func SkipPythonDocstringNodesFn(docstringNodes []*sitter.Node) SkipNodeFn {
	return func(node *sitter.Node) bool {
		if node == nil {
			return false
		}

		for _, docstringNode := range docstringNodes {
			if docstringNode.Equal(node) {
				return true
			}
		}

		return false
	}
}

func GetPythonDocstringNodes(rootNode *sitter.Node) ([]*sitter.Node, error) {
	docstringNodes := []*sitter.Node{}
	iter := sitter.NewNamedIterator(rootNode, sitter.BFSMode)
	err := iter.ForEach(func(node *sitter.Node) error {
		if node.Type() == "function_definition" {
			docstringNode := GetPythonDocstringNode(node)
			if docstringNode != nil {
				docstringNodes = append(docstringNodes, docstringNode)
			}
		}
		return nil
	})

	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}

	return docstringNodes, nil
}

func GetPythonDocstringNode(functionNode *sitter.Node) *sitter.Node {
	var blockNode *sitter.Node
	// We are looking for the `(block (expression_statement (string)) ...)` pattern.
	for i := 0; i < int(functionNode.NamedChildCount()); i++ {
		namedChild := functionNode.NamedChild(i)
		if namedChild.Type() == "block" {
			blockNode = namedChild
		}
	}

	if blockNode == nil {
		return nil
	}

	firstBlockChild := blockNode.NamedChild(0)
	if firstBlockChild == nil || firstBlockChild.Type() != "expression_statement" {
		return nil
	}

	firstExpressionStatementChild := firstBlockChild.NamedChild(0)
	if firstExpressionStatementChild == nil || firstExpressionStatementChild.Type() != "string" {
		return nil
	}

	return firstExpressionStatementChild
}

func GetPythonDocstring(docstringNode *sitter.Node, code []byte) string {
	if docstringNode == nil {
		return ""
	}
	return StripCommentDelimiters(docstringNode.Content(code))
}
