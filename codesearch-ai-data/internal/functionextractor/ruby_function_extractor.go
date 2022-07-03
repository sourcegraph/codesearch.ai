package functionextractor

import (
	ph "codesearch-ai-data/internal/parsinghelpers"
	sp "codesearch-ai-data/internal/sitterparsers"
	"errors"
	"io"

	sitter "github.com/smacker/go-tree-sitter"
)

type rubyFunctionExtractor struct {
	*functionExtractor
}

func NewRubyFunctionExtractor(minLines int) FunctionExtractor {
	return &rubyFunctionExtractor{&functionExtractor{sp.GetRubyParser(), minLines}}
}

func (rfe *rubyFunctionExtractor) Extract(code []byte) ([]*ExtractedFunction, error) {
	tree := rfe.parser.Parse(nil, code)

	rootNode := tree.RootNode()
	if rootNode.HasError() {
		return nil, errors.New("Error encountered while parsing")
	}

	extractedFunctions := []*ExtractedFunction{}
	iter := sitter.NewNamedIterator(tree.RootNode(), sitter.BFSMode)
	err := iter.ForEach(func(node *sitter.Node) error {
		if node.Type() == "method" {
			filteredNodes, commentNodes := ph.StripComments(node, nil)
			inlineComments := ph.StripCommentNodesDelimiters(commentNodes, code)
			prettyFormattedCode := ph.PrettyFormatNodes(filteredNodes, code)

			if !isFunctionRightSize(prettyFormattedCode, rfe.minLines) {
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
