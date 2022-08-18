package functionextractor

import (
	ph "codesearch-ai-data/internal/parsinghelpers"
	sp "codesearch-ai-data/internal/sitterparsers"
	"context"
	"errors"
	"io"

	sitter "github.com/smacker/go-tree-sitter"
)

var ignoredJavascriptFunctionIdentifiers = []string{"toString", "toLocaleString", "valueOf"}

type javascriptFunctionExtractor struct {
	*functionExtractor
}

func NewJavascriptFunctionExtractor(minLines int) FunctionExtractor {
	return &javascriptFunctionExtractor{&functionExtractor{sp.GetJavascriptParser(), minLines}}
}

func getInlineFunctionIdentifierAndDocstring(node *sitter.Node, code []byte) (identifier string, docstring string) {
	nodeParent := node.Parent()
	if nodeParent == nil {
		return
	}

	nodeParentType := nodeParent.Type()
	if nodeParentType == "variable_declarator" {
		identifier = ph.FindNamedIdentifier(nodeParent, code)
		docstring = ph.GetPrecedingFunctionDocstring(nodeParent.Parent(), code)
	} else if nodeParentType == "pair" {
		identifier = ph.FindNamedIdentifier(nodeParent, code)
		docstring = ph.GetPrecedingFunctionDocstring(nodeParent, code)
	}

	return
}

func (jfe *javascriptFunctionExtractor) Extract(ctx context.Context, code []byte) ([]*ExtractedFunction, error) {
	tree, err := jfe.parser.ParseCtx(ctx, nil, code)
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
		docstring := ""
		identifier := ""

		nodeType := node.Type()
		if nodeType == "method_definition" || nodeType == "function_declaration" {
			docstring = ph.GetPrecedingFunctionDocstring(node, code)
			identifier = ph.FindNamedIdentifier(node, code)
		} else if nodeType == "arrow_function" || nodeType == "function" {
			identifier, docstring = getInlineFunctionIdentifierAndDocstring(node, code)
		} else {
			return nil
		}

		if contains(ignoredJavascriptFunctionIdentifiers, identifier) {
			return nil
		}

		filteredNodes, commentNodes := ph.StripComments(node, nil)
		inlineComments := ph.StripCommentNodesDelimiters(commentNodes, code)
		prettyFormattedCode := ph.PrettyFormatNodes(filteredNodes, code)

		if !isFunctionRightSize(prettyFormattedCode, jfe.minLines) {
			return nil
		}

		extractedFunctions = append(
			extractedFunctions,
			NewExtractedFunction(
				identifier,
				prettyFormattedCode,
				inlineComments,
				docstring,
				node,
				code,
			),
		)

		return nil
	})

	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}

	return extractedFunctions, nil
}
