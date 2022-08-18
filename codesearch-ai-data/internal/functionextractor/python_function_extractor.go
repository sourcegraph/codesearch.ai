package functionextractor

import (
	ph "codesearch-ai-data/internal/parsinghelpers"
	sp "codesearch-ai-data/internal/sitterparsers"
	"context"
	"errors"
	"io"

	sitter "github.com/smacker/go-tree-sitter"
)

type pythonFunctionExtractor struct {
	*functionExtractor
}

func NewPythonFunctionExtractor(minLines int) FunctionExtractor {
	return &pythonFunctionExtractor{&functionExtractor{sp.GetPythonParser(), minLines}}
}

func (pfe *pythonFunctionExtractor) Extract(ctx context.Context, code []byte) ([]*ExtractedFunction, error) {
	tree, err := pfe.parser.ParseCtx(ctx, nil, code)
	if err != nil {
		return nil, err
	}

	rootNode := tree.RootNode()
	if rootNode.HasError() {
		return nil, errors.New("error encountered while parsing")
	}

	extractedFunctions := []*ExtractedFunction{}
	iter := sitter.NewNamedIterator(tree.RootNode(), sitter.BFSMode)
	err = iter.ForEach(func(node *sitter.Node) error {
		if node.Type() == "function_definition" {
			docstringNodes, err := ph.GetPythonDocstringNodes(node)
			if err != nil {
				return err
			}

			// We have to traverse the entire function subtree to remove nested function docstrings.
			filteredNodes, commentNodes := ph.StripComments(node, ph.SkipPythonDocstringNodesFn(docstringNodes))
			inlineComments := ph.StripCommentNodesDelimiters(commentNodes, code)
			prettyFormattedCode := ph.PrettyFormatNodes(filteredNodes, code)

			if !isFunctionRightSize(prettyFormattedCode, pfe.minLines) {
				return nil
			}

			docstringNode := ph.GetPythonDocstringNode(node)
			docstring := ph.GetPythonDocstring(docstringNode, code)

			extractedFunctions = append(
				extractedFunctions,
				NewExtractedFunction(
					ph.FindNamedIdentifier(node, code),
					prettyFormattedCode,
					inlineComments,
					docstring,
					node,
					code,
				),
			)
		}
		return nil
	})

	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}

	return extractedFunctions, nil
}
