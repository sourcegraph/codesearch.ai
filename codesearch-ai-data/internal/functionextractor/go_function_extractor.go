package functionextractor

import (
	ph "codesearch-ai-data/internal/parsinghelpers"
	sp "codesearch-ai-data/internal/sitterparsers"
	"context"
	"errors"
	"io"

	sitter "github.com/smacker/go-tree-sitter"
)

type goFunctionExtractor struct {
	*functionExtractor
}

func NewGoFunctionExtractor(minLines int) FunctionExtractor {
	return &goFunctionExtractor{&functionExtractor{sp.GetGoParser(), minLines}}
}

func (gfe *goFunctionExtractor) Extract(ctx context.Context, code []byte) ([]*ExtractedFunction, error) {
	tree, err := gfe.parser.ParseCtx(ctx, nil, code)
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
		nodeType := node.Type()
		if nodeType == "method_declaration" || nodeType == "function_declaration" {
			filteredNodes, commentNodes := ph.StripComments(node, nil)
			inlineComments := ph.StripCommentNodesDelimiters(commentNodes, code)
			prettyFormattedCode := ph.PrettyFormatNodes(filteredNodes, code)

			if !isFunctionRightSize(prettyFormattedCode, gfe.minLines) {
				return nil
			}

			extractedFunctions = append(
				extractedFunctions,
				NewExtractedFunction(
					ph.FindNamedIdentifier(node, code),
					prettyFormattedCode,
					inlineComments,
					ph.GetPrecedingFunctionDocstring(node, code),
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
