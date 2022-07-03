package parsinghelpers

import (
	sp "codesearch-ai-data/internal/sitterparsers"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/hexops/autogold"
	sitter "github.com/smacker/go-tree-sitter"
)

func TestStripComments(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		parser *sitter.Parser
	}{
		{
			name:   "RubyStripComments",
			path:   "../testdata/test.rb",
			parser: sp.GetRubyParser(),
		},
		{
			name:   "GoStripComments",
			path:   "../testdata/test.go",
			parser: sp.GetGoParser(),
		},
		{
			name:   "PythonStripComments",
			path:   "../testdata/test.py",
			parser: sp.GetPythonParser(),
		},
		{
			name:   "JavascriptStripComments",
			path:   "../testdata/test.js",
			parser: sp.GetJavascriptParser(),
		},
		{
			name:   "JavaStripComments",
			path:   "../testdata/test.java",
			parser: sp.GetJavaParser(),
		},
		{
			name:   "PhpStripComments",
			path:   "../testdata/test.php",
			parser: sp.GetPhpParser(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testCode, err := ioutil.ReadFile(tt.path)
			if err != nil {
				t.Fatal(err)
			}

			tree := tt.parser.Parse(nil, testCode)
			rootNode := tree.RootNode()
			if rootNode.HasError() {
				t.Fatal("Error encountered while parsing")
			}

			// TODO: Refactor
			var skipNodeFn SkipNodeFn = nil
			if tt.name == "PythonStripComments" {
				docstringNodes, err := GetPythonDocstringNodes(rootNode)
				if err != nil {
					t.Fatal(err)
				}
				skipNodeFn = SkipPythonDocstringNodesFn(docstringNodes)
			}

			filteredNodes, commentNodes := StripComments(rootNode, skipNodeFn)
			prettyFormattedCode := PrettyFormatNodes(filteredNodes, testCode)
			comments := StripCommentNodesDelimiters(commentNodes, testCode)

			autogold.Equal(t, prettyFormattedCode+"\n\n"+strings.Join(comments, "\n"))
		})
	}
}
