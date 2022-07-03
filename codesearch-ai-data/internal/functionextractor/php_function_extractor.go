package functionextractor

import (
	ph "codesearch-ai-data/internal/parsinghelpers"
	sp "codesearch-ai-data/internal/sitterparsers"
	"errors"
	"io"

	sitter "github.com/smacker/go-tree-sitter"
)

var ignoredPhpFunctionIdentifiers = []string{
	"__construct",
	"__destruct",
	"__call",
	"__callStatic",
	"__get",
	"__set",
	"__isset",
	"__unset",
	"__sleep",
	"__wakeup",
	"__toString",
	"__invoke",
	"__set_state",
	"__clone",
	"__debugInfo",
	"__serialize",
	"__unserialize",
}

type phpFunctionExtractor struct {
	*functionExtractor
}

func NewPhpFunctionExtractor(minLines int) FunctionExtractor {
	return &phpFunctionExtractor{&functionExtractor{sp.GetPhpParser(), minLines}}
}

func (pfe *phpFunctionExtractor) Extract(code []byte) ([]*ExtractedFunction, error) {
	tree := pfe.parser.Parse(nil, code)

	rootNode := tree.RootNode()
	if rootNode.HasError() {
		return nil, errors.New("Error encountered while parsing")
	}
	extractedFunctions := []*ExtractedFunction{}
	iter := sitter.NewNamedIterator(tree.RootNode(), sitter.BFSMode)
	err := iter.ForEach(func(node *sitter.Node) error {
		nodeType := node.Type()

		if nodeType == "method_declaration" || nodeType == "function_definition" {
			identifier := ph.FindNamedIdentifier(node, code)
			if contains(ignoredPhpFunctionIdentifiers, identifier) {
				return nil
			}

			filteredNodes, commentNodes := ph.StripComments(node, nil)
			inlineComments := ph.StripCommentNodesDelimiters(commentNodes, code)
			prettyFormattedCode := ph.PrettyFormatNodes(filteredNodes, code)

			if !isFunctionRightSize(prettyFormattedCode, pfe.minLines) {
				return nil
			}

			extractedFunctions = append(
				extractedFunctions,
				NewExtractedFunction(
					identifier,
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
